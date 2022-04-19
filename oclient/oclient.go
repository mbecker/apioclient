package oclient

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	// "fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/sessions"
)

const (
	KEYCLOAK     = "keycloak"
	STRAVA       = "strava"
	LINKEDIN     = "linkedin"
	SPOTIFY      = "spotify"
	GITHUB       = "github"
	FITBIT       = "fitbit"
	OURA         = "oura"
	AUTHORIZE    = "authorization_code"
	REFRESH      = "refresh_token"
	SECRET       = "secret"
	PKCE         = "pkce"
	SESSION_NAME = "oclient"
)

type OClient struct {
	services map[string]map[string]string
	store    *sessions.CookieStore
}

func InitOclient(sessionKey string, servicesFile string) (*OClient, error) {
	PkceInit()

	oclient := OClient{
		services: map[string]map[string]string{},
		store:    sessions.NewCookieStore([]byte(sessionKey)),
	}
	err := oclient.loadConfig(servicesFile)
	return &oclient, err
}

//== Services

func (oclient *OClient) loadConfig(fname string) (err error) {
	file, err := os.Open(fname)
	if err != nil {
		return
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	json.Unmarshal([]byte(byteValue), &oclient.services)
	for k, v := range oclient.services {
		v["client_id"] = os.Getenv(v["client_id"])
		if v["client_id"] == "" {
			err = errors.New("Missing service client_id for " + k)
			return
		}
		v["client_secret"] = os.Getenv(v["client_secret"])
		if v["client_id"] == "" {
			err = errors.New("Missing service client_secret for " + k)
			return
		}
	}
	return
}

//== PKCE

func PkceInit() {
	rand.Seed(time.Now().UnixNano())
}

//string of pkce allowed chars
func PkceVerifier(length int) string {
	if length > 128 {
		length = 128
	}
	if length < 43 {
		length = 43
	}
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

//base64-URL-encoded SHA256 hash of verifier, per rfc 7636
func PkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	return (challenge)
}

//== State Management

const (
	GcPeriod        = 60  //minutes - minimum ideal time between GC runs (unless MaxState)
	InitAuthTimeout = 10  //minutes - amount of time user has to complete Authorization and get Access Code from Authorization Server
	MaxState        = 400 //max allowed length of state map, to prevent malicious memory overflow
)

type State struct {
	CreatedAt     time.Time
	Service       string
	AuthType      string
	PkceVerifier  string
	PkceChallenge string
}

var state = make(map[string]*State)
var lastGc = time.Now().UTC()
var mutex = &sync.Mutex{}

//get the payload for a state, check expiration, and delete
func getState(key string) (value *State) {
	mutex.Lock()
	v, exists := state[key]
	if exists {
		n := time.Now().UTC()
		if n.After(v.CreatedAt.Add(InitAuthTimeout * time.Minute)) {
			value = nil //don't accept expired state
		} else {
			value = v
		}
		delete(state, key)
	} else {
		value = nil
	}
	defer mutex.Unlock()
	return
}

//set the payload for a state, set expiration, do gc as needed
func setState(key string, value *State) {
	mutex.Lock()
	n := time.Now().UTC()
	value.CreatedAt = n
	state[key] = value
	//gc
	authTimeout := InitAuthTimeout * time.Minute //type Duration
	gcTime := lastGc.Add(GcPeriod * time.Minute)
	if n.After(gcTime) || len(state) >= MaxState {
		for ok := true; ok; ok = len(state) >= MaxState { //keep going till below MaxState, 1/2 each cycle
			for k, v := range state {
				expiresAt := v.CreatedAt.Add(authTimeout)
				if n.After(expiresAt) {
					delete(state, k)
				}
			}
			authTimeout /= 2
		}
		lastGc = time.Now().UTC()
	}
	defer mutex.Unlock()
	return
}

//== Cookie Helpers

const CookiePrefix = "oclient-"

func cookieName(service string) string {
	return (CookiePrefix + service)
}

//generic cookie setter
func (oclient *OClient) setCookie(w http.ResponseWriter, r *http.Request, kcToken *KeycloakToken, kcClaims *KeycloakClaims, cookieName string) {
	tokens, err := kcToken.String()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tok64 := base64.StdEncoding.EncodeToString([]byte(tokens))

	idtokens, err := kcClaims.String()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	idToken64 := base64.StdEncoding.EncodeToString([]byte(idtokens))
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    tok64,
		HttpOnly: true,
		Secure:   false, //use true for production
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	idTokenCookie := http.Cookie{
		Name:     fmt.Sprintf("%s-id", cookieName),
		Value:    idToken64,
		HttpOnly: true,
		Secure:   false, //use true for production
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &idTokenCookie)

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := oclient.store.Get(r, SESSION_NAME)
	// Set some session values.
	session.Values["email"] = kcClaims.Email
	session.Values["isAuthenticated"] = true

	log.Printf("SET COOKIE - Save keycloak token access token: %s\n", kcToken.AccessToken)
	session.Values["token"] = &kcToken.AccessToken
	// kc, _ := kcClaims.String()
	// log.Printf("SET COOKIE - Save keycloak claims: %s\n", kc)
	// session.Values["idtoken"] = kc
	// Save it before we write to the response/return from the handler.
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func (oclient *OClient) DeleteCookieSession(w http.ResponseWriter, r *http.Request) {
	cookieName := cookieName("keycloak")
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, //use true for production
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	idTokenCookie := http.Cookie{
		Name:     fmt.Sprintf("%s-id", cookieName),
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, //use true for production
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &idTokenCookie)

	session, _ := oclient.store.Get(r, SESSION_NAME)
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

//generic cookie getter
func getCookie(r *http.Request, cookieName string) (kcToken *KeycloakToken, err error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return
	}
	tokb, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return
	}
	kcToken, err = newKeycloakToken(string(tokb))
	return
}

func getCookieIdToken(r *http.Request, cookieName string) (token string, err error) {
	cookie, err := r.Cookie(fmt.Sprintf("%s-id", cookieName))
	if err != nil {
		return
	}
	tokb, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return
	}
	token = string(tokb)
	return
}

func (oclient *OClient) GetIdToken(r *http.Request) (email string, isAuthenticated bool, token string, err error) {
	session, err := oclient.store.Get(r, SESSION_NAME)
	if err != nil {
		return
	}
	emaill, ok := session.Values["email"].(string)
	if ok {
		email = emaill
	}
	isAuthenticatedd, ok := session.Values["isAuthenticated"].(bool)
	if ok {
		isAuthenticated = isAuthenticatedd
	}
	tokenn, ok := session.Values["token"].(string)
	if ok {
		token = tokenn
	}
	return
}

//== API Helpers

//build service Code Authorize Link and save state as pkceVerifier (128)
func (oclient *OClient) AuthLink(r *http.Request, authtype string, service string) (result string) {
	stData := State{Service: service, AuthType: authtype}
	st := PkceVerifier(128)
	result = oclient.services[service]["authorize_endpoint"]
	result += "?client_id=" + oclient.services[service]["client_id"]
	result += "&response_type=code&redirect_uri="
	result += url.QueryEscape(oclient.services[service]["redirect_uri"])
	result += "&scope=" + oclient.services[service]["scope"]
	result += oclient.services[service]["prompt"]
	if authtype == PKCE {
		stData.PkceVerifier = PkceVerifier(128)
		stData.PkceChallenge = PkceChallenge(stData.PkceVerifier)
		result += "&code_challenge=" + stData.PkceChallenge
		result += "&code_challenge_method=S256"
	}
	result += "&state=" + st
	setState(st, &stData)
	fmt.Println("oclient - Debug Authorize Link: ", result)
	return
}

//make call to a resource api, add oauth bearer token
func (oclient *OClient) ApiRequest(w http.ResponseWriter, r *http.Request, service, method, url string, data map[string]interface{}) (response *http.Response, err error) {
	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	var body io.Reader
	if data == nil {
		body = nil
	} else {
		var requestBody []byte
		requestBody, err = json.Marshal(data)
		if err != nil {
			return
		}
		body = bytes.NewBuffer(requestBody)
	}
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	err = oclient.setHeader(w, r, service, request)
	if err != nil {
		err = errors.New("Unable to set Header: " + err.Error())
		return
	}
	response, err = client.Do(request)
	return
}

func epochSeconds() int64 {
	now := time.Now()
	secs := now.Unix()
	return secs
}

//get Access Token via cookie, refresh if expired, set header bearer token
func (oclient *OClient) setHeader(w http.ResponseWriter, r *http.Request, service string, newReq *http.Request) error {
	kcToken, err := getCookie(r, cookieName(service))
	if err != nil {
		return err
	}

	if epochSeconds() > kcToken.ExpiresAt { //token has expired, refresh it
		if oclient.services[service]["refresh_allowed"] == "false" {
			return errors.New("Non-refreshable Token Expired, Re-authorize")
		}

		if kcToken.RefreshToken == "" {
			return errors.New("Refresh Token Not Found")
		}

		kcToken, kcClaims, err := oclient.getToken(w, r, service, REFRESH, kcToken.RefreshToken, SECRET, "")
		if err != nil {
			return errors.New("Refresh Token Not Found")
		}
		oclient.setCookie(w, r, kcToken, kcClaims, cookieName(service)) //note: must set cookie before writing to responsewriter

	}
	newReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", kcToken.AccessToken))
	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("Accept", "application/json")
	return nil
}

//== Access Token

//exchange the Authorization Code for Access Token
func (oclient *OClient) ExchangeCode(w http.ResponseWriter, r *http.Request, code string, state string) (err error) {
	statePtr := getState(state)
	if statePtr == nil {
		err = errors.New("State Key not found")
		return
	}
	token, idToken, err := oclient.getToken(w, r, statePtr.Service, AUTHORIZE, code, statePtr.AuthType, statePtr.PkceVerifier)
	if err != nil {
		return
	}
	oclient.setCookie(w, r, token, idToken, cookieName(statePtr.Service)) //note: must set cookie before writing to responsewriter
	return
}

//wrapper to set accept header
func jsonPost(url string, body io.Reader) (resp *http.Response, err error) {
	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return client.Do(req)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func basicPost(url string, body io.Reader, ba string) (resp *http.Response, err error) {
	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+ba)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	return client.Do(req)
}

//subtract a small delta from exires_at to account for transport time
const DELTASECS = 5

//get a token from authorization endpoint
func (oclient *OClient) getToken(w http.ResponseWriter, r *http.Request, service string, tokType string, code string, authType string, verifier string) (kcToken *KeycloakToken, kcClaims *KeycloakClaims, err error) {
	rParams := map[string]string{
		"client_id":    oclient.services[service]["client_id"],
		"redirect_uri": oclient.services[service]["redirect_uri"],
	}
	switch tokType {
	case AUTHORIZE:
		rParams["code"] = code
		rParams["grant_type"] = AUTHORIZE
	case REFRESH:
		rParams["refresh_token"] = code
		rParams["grant_type"] = REFRESH
	default:
		err = errors.New("Unknown tokType")
		return
	}
	switch authType {
	case SECRET:
		rParams["client_secret"] = oclient.services[service]["client_secret"]
	case PKCE:
		rParams["code_verifier"] = verifier
	default:
		err = errors.New("Unknown authType")
		return
	}
	var resp *http.Response
	switch oclient.services[service]["post_type"] {
	case "basic":
		form := url.Values{}
		for k, v := range rParams {
			form.Set(k, v)
		}

		basic := basicAuth(rParams["client_id"], rParams["client_secret"])

		resp, err = basicPost(oclient.services[service]["token_endpoint"], strings.NewReader(form.Encode()), basic)
		if err != nil {
			return
		}
	case "json":
		var requestBody []byte
		requestBody, err = json.Marshal(rParams)
		if err != nil {
			return
		}
		resp, err = jsonPost(oclient.services[service]["token_endpoint"], bytes.NewBuffer(requestBody))
		if err != nil {
			return
		}

	case "form":
		vals := url.Values{}
		for k, v := range rParams {
			vals.Set(k, v)
		}
		resp, err = http.PostForm(oclient.services[service]["token_endpoint"], vals)
		if err != nil {
			return
		}
	default:
		err = errors.New("Unknown post_type")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = errors.New(string(body))
		return
	}

	//Create new Keycloak Token
	kcToken, err = newKeycloakToken(string(body))
	if err != nil {
		return
	}
	// Deleting the tokMap.IDToken that the cookie value fits
	kcClaims, err = parseIdToken(kcToken.IDToken)
	if err != nil {
		return
	}
	kcToken.IDToken = ""

	writeJsonFile("token.json", kcToken)
	writeJsonFile("claims.json", kcClaims)

	return
}

func parseIdToken(idToken string) (*KeycloakClaims, error) {
	var kcClaims KeycloakClaims
	// Get the JWKS URL from an environment variable.
	jwksURL := "https://localhost/realms/azureapidev/protocol/openid-connect/certs"

	// Confirm the environment variable is not empty.
	if jwksURL == "" {
		return &kcClaims, errors.New("JWKS_URL environment variable must be populated.")
	}
	// Create the JWKS from the resource at the given URL.
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		return &kcClaims, fmt.Errorf("Failed to get the JWKS from the given URL.\nError:%s", err.Error())
	}

	token, err := jwt.ParseWithClaims(idToken, &KeycloakClaims{}, jwks.Keyfunc)
	if err != nil {
		return &kcClaims, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return &kcClaims, errors.New("Token is not valid")
	}
	kcClaimss, ok := token.Claims.(*KeycloakClaims)
	if !ok {
		return &kcClaims, errors.New("Claims are not of type KeyclokClaims")
	}
	return kcClaimss, nil
}

type KeycloakToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	ExpiresAt        int64  `json:"expires_at"`
	IDToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	SessionState     string `json:"session_state"`
	TokenType        string `json:"token_type"`
}

func newKeycloakToken(body string) (*KeycloakToken, error) {
	var kcToken KeycloakToken
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	err := decoder.Decode(&kcToken)
	if err != nil {
		return &kcToken, errors.New("decoder.Decode: " + err.Error())
	}
	var expiresIn int64
	if kcToken.ExpiresIn == 0 { //no expiration, so make it a year
		kcToken.ExpiresIn = 31536000
	}
	kcToken.ExpiresAt = epochSeconds() + expiresIn - DELTASECS
	return &kcToken, nil
}

func (kcToken *KeycloakToken) String() (string, error) {
	b, err := json.Marshal(kcToken)
	if err != nil {
		err = errors.New("json.Marshal: " + err.Error())
		return "", err
	}
	result := string(b)
	return result, nil

}

func writeJsonFile(filename string, tok interface{}) {
	// #### WRITE JSON FILE
	file, _ := json.MarshalIndent(tok, "", " ")
	_ = ioutil.WriteFile(filename, file, 0644)
	// #### WRITE JSON FILE (END)

}

type KeycloakClaims struct {
	jwt.StandardClaims
	Exp               int    `json:"exp,omitempty"`
	Iat               int    `json:"iat,omitempty"`
	AuthTime          int    `json:"auth_time,omitempty"`
	Jti               string `json:"jti,omitempty"`
	Iss               string `json:"iss,omitempty"`
	Audience          string `json:"aud,omitempty"`
	Sub               string `json:"sub,omitempty"`
	Typ               string `json:"typ,omitempty"`
	Azp               string `json:"azp,omitempty"`
	SessionState      string `json:"session_state,omitempty"`
	AtHash            string `json:"at_hash,omitempty"`
	Acr               string `json:"acr,omitempty"`
	Sid               string `json:"sid,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	Email             string `json:"email,omitempty"`
}

func (kc *KeycloakClaims) String() (string, error) {
	var s string
	b, err := json.Marshal(&kc)
	if err != nil {
		return s, err
	}
	return string(b), nil
}
