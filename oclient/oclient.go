package oclient

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

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
)

const (
	KEYCLOAK  = "keycloak"
	STRAVA    = "strava"
	LINKEDIN  = "linkedin"
	SPOTIFY   = "spotify"
	GITHUB    = "github"
	FITBIT    = "fitbit"
	OURA      = "oura"
	AUTHORIZE = "authorization_code"
	REFRESH   = "refresh_token"
	SECRET    = "secret"
	PKCE      = "pkce"
)

func InitOclient() error {
	PkceInit()
	return loadConfig("oclient/services.json", &services)
}

//== Services

var services map[string]map[string]string

func loadConfig(fname string, config *map[string]map[string]string) (err error) {
	file, err := os.Open(fname)
	if err != nil {
		return
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	json.Unmarshal([]byte(byteValue), config)
	// for k, v := range *config {
	// 	v["client_id"] = os.Getenv(v["client_id"])
	// 	if v["client_id"] == "" {
	// 		err = errors.New("Missing service client_id for " + k)
	// 		return
	// 	}
	// 	v["client_secret"] = os.Getenv(v["client_secret"])
	// 	if v["client_id"] == "" {
	// 		err = errors.New("Missing service client_secret for " + k)
	// 		return
	// 	}
	// }
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

const CookiePrefix = "OClient-"

func cookieName(service string) string {
	fmt.Printf("Cookie Name: %s\n", service)
	return (CookiePrefix + service)
}

//generic cookie setter
func setCookie(w http.ResponseWriter, token string, email string, cookieName string) {
	fmt.Printf("Set cookie: %s\n", cookieName)
	tok64 := base64.StdEncoding.EncodeToString([]byte(token))
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    tok64,
		HttpOnly: false,
		Secure:   true, //use true for production
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	// cookie2 := http.Cookie{
	// 	Name:     "email",
	// 	Value:    email,
	// 	HttpOnly: true,
	// 	Secure:   false, //use true for production
	// 	Path:     "/",
	// 	SameSite: http.SameSiteLaxMode,
	// }
	// http.SetCookie(w, &cookie2)

	return
}

//generic cookie getter
func getCookie(r *http.Request, cookieName string) (token string, err error) {
	cookie, err := r.Cookie(cookieName)
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

//== API Helpers

//build service Code Authorize Link and save state as pkceVerifier (128)
func AuthLink(r *http.Request, authtype string, service string) (result string) {
	stData := State{Service: service, AuthType: authtype}
	st := PkceVerifier(128)
	result = services[service]["authorize_endpoint"]
	result += "?client_id=" + services[service]["client_id"]
	result += "&response_type=code&redirect_uri="
	result += url.QueryEscape(services[service]["redirect_uri"])
	result += "&scope=" + services[service]["scope"]
	result += services[service]["prompt"]
	if authtype == PKCE {
		stData.PkceVerifier = PkceVerifier(128)
		stData.PkceChallenge = PkceChallenge(stData.PkceVerifier)
		result += "&code_challenge=" + stData.PkceChallenge
		result += "&code_challenge_method=S256"
	}
	result += "&state=" + st
	setState(st, &stData)
	fmt.Println("Debug Authorize Link: ", result)
	return
}

//make call to a resource api, add oauth bearer token
func ApiRequest(w http.ResponseWriter, r *http.Request, service, method, url string, data map[string]interface{}) (response *http.Response, err error) {
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
	err = setHeader(w, r, service, request)
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
func setHeader(w http.ResponseWriter, r *http.Request, service string, newReq *http.Request) (err error) {
	token, err := getCookie(r, cookieName(service))
	if err != nil {
		return
	}
	var tokMap map[string]interface{}

	// err = json.Unmarshal([]byte(token), &tokMap)
	// normally as above, but we want numbers as ints vs floats
	decoder := json.NewDecoder(strings.NewReader(token))
	decoder.UseNumber()
	err = decoder.Decode(&tokMap)

	expiresAt, err := tokMap["expires_at"].(json.Number).Int64()
	if err != nil {
		return
	}
	if epochSeconds() > expiresAt { //token has expired, refresh it
		if services[service]["refresh_allowed"] == "false" {
			err = errors.New("Non-refreshable Token Expired, Re-authorize")
			return
		}
		refresh, exists := tokMap["refresh_token"]
		if !exists {
			err = errors.New("Refresh Token Not Found")
			return
		}
		var newToken string
		var email string
		newToken, email, err = getToken(w, r, service, REFRESH, refresh.(string), SECRET, "")
		if err != nil {
			return
		}
		setCookie(w, newToken, email, cookieName(service)) //note: must set cookie before writing to responsewriter
		decoder = json.NewDecoder(strings.NewReader(newToken))
		decoder.UseNumber()
		tokMap = make(map[string]interface{})
		err = decoder.Decode(&tokMap)
		if err != nil {
			return
		}
	}
	newReq.Header.Add("Authorization", "Bearer "+tokMap["access_token"].(string))
	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("Accept", "application/json")
	return
}

//== Access Token

//exchange the Authorization Code for Access Token
func ExchangeCode(w http.ResponseWriter, r *http.Request, code string, state string) (err error) {
	statePtr := getState(state)
	if statePtr == nil {
		err = errors.New("State Key not found")
		return
	}
	token, email, err := getToken(w, r, statePtr.Service, AUTHORIZE, code, statePtr.AuthType, statePtr.PkceVerifier)
	if err != nil {
		return
	}
	setCookie(w, token, email, cookieName(statePtr.Service)) //note: must set cookie before writing to responsewriter
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
func getToken(w http.ResponseWriter, r *http.Request, service string, tokType string, code string, authType string, verifier string) (result string, email string, err error) {
	rParams := map[string]string{
		"client_id":    services[service]["client_id"],
		"redirect_uri": services[service]["redirect_uri"],
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
		rParams["client_secret"] = services[service]["client_secret"]
	case PKCE:
		rParams["code_verifier"] = verifier
	default:
		err = errors.New("Unknown authType")
		return
	}
	var resp *http.Response
	switch services[service]["post_type"] {
	case "basic":
		form := url.Values{}
		for k, v := range rParams {
			form.Set(k, v)
		}

		basic := basicAuth(rParams["client_id"], rParams["client_secret"])

		resp, err = basicPost(services[service]["token_endpoint"], strings.NewReader(form.Encode()), basic)
		if err != nil {
			return
		}
	case "json":
		var requestBody []byte
		requestBody, err = json.Marshal(rParams)
		if err != nil {
			return
		}
		resp, err = jsonPost(services[service]["token_endpoint"], bytes.NewBuffer(requestBody))
		if err != nil {
			return
		}

	case "form":
		vals := url.Values{}
		for k, v := range rParams {
			vals.Set(k, v)
		}
		resp, err = http.PostForm(services[service]["token_endpoint"], vals)
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
	//check for expires_at
	var tokMap map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	err = decoder.Decode(&tokMap)
	if err != nil {
		err = errors.New("decoder.Decode: " + err.Error())
		return
	}
	expire, exists := tokMap["expires_at"]

	if exists {
		result = string(body)
		return
	}
	var expiresIn int64
	expire, exists = tokMap["expires_in"]
	if !exists { //no expiration, so make it a year
		expiresIn = 31536000
	} else {
		expiresIn, err = expire.(json.Number).Int64()
	}
	tokMap["expires_at"] = epochSeconds() + expiresIn - DELTASECS
	b, err := json.Marshal(tokMap)
	if err != nil {
		err = errors.New("json.Marshal: " + err.Error())
		return
	}
	idt, ok := tokMap["id_token"]
	if !ok {
		err = errors.New("No id_token")
		return
	}
	idtoken, ok := idt.(string)
	if !ok {
		err = errors.New("No id_token string")
		return
	}
	// Get the JWKS URL from an environment variable.
	jwksURL := "https://localhost/realms/azureapidev/protocol/openid-connect/certs"

	// Confirm the environment variable is not empty.
	if jwksURL == "" {
		err = errors.New("JWKS_URL environment variable must be populated.")
		return
	}
	// Create the JWKS from the resource at the given URL.
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		err = fmt.Errorf("Failed to get the JWKS from the given URL.\nError:%s", err.Error())
		return
	}

	token, err := jwt.ParseWithClaims(idtoken, &MyCustomClaims{}, jwks.Keyfunc)
	if err != nil {
		err = fmt.Errorf("failed to parse token: %w", err)
		return
	}
	if !token.Valid {
		fmt.Println("Token is not valid")
	}
	if claims, ok := token.Claims.(*MyCustomClaims); ok {
		email = claims.Email
	} else {
		fmt.Println("No custom claims in token")
	}
	result = string(b)
	return
}

type MyCustomClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}
