package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mbecker/apioclient/oclient"
)

const (
	SESSION_KEY = "SESSION_KEY"
)

var oc *oclient.OClient

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	oc, err = oclient.InitOclient(os.Getenv(SESSION_KEY), "services.json")
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	r := mux.NewRouter()
	fs := http.FileServer(http.Dir("./tailwind/dist/"))
	r.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", fs))
	r.HandleFunc("/", PageHomeHandler)
	r.HandleFunc("/signout", SignoutHandler)
	r.HandleFunc("/page/api", PageApiHandler)
	r.HandleFunc("/authlink/{authtype}/{service}", AuthlinkHandler)
	r.HandleFunc("/redirect", RedirectHandler)
	r.HandleFunc("/strava/get/athlete", StravaGetAthleteHandler)
	r.HandleFunc("/strava/get/activities", StravaGetActivitiesHandler)
	r.HandleFunc("/linkedin/get/me", LinkedinGetMeHandler)
	r.HandleFunc("/spotify/get/me", SpotifyGetMeHandler)
	r.HandleFunc("/spotify/get/newreleases", SpotifyGetNewReleasesHandler)
	r.HandleFunc("/spotify/put/rename", SpotifyPutRenameHandler)
	r.HandleFunc("/github/get/user", GithubGetUserHandler)
	r.HandleFunc("/fitbit/get/user", FitbitGetUserHandler)
	r.HandleFunc("/fitbit/get/heartrate", FitbitGetHeartrateHandler)
	r.HandleFunc("/fitbit/get/sleep", FitbitGetSleepHandler)
	r.HandleFunc("/oura/get/user", OuraGetUserHandler)
	r.HandleFunc("/oura/get/sleep", OuraGetSleepHandler)
	r.HandleFunc("/oura/get/activity", OuraGetActivityHandler)
	r.HandleFunc("/oura/get/readiness", OuraGetReadinessHandler)
	http.Handle("/", r)
	fmt.Println(">>>>>>> OClient started at:", port)
	log.Fatal(http.ListenAndServeTLS(":"+port, os.Getenv("CERT"), os.Getenv("CERT_KEY"), nil))

	// Create a new engine by passing the template folder
	// and template extension using <engine>.New(dir, ext string)
	// engine := html.New("./views", ".html")

	// // Reload the templates on each render, good for development
	// engine.Reload(true) // Optional. Default: false

	// // Debug will print each template that is parsed, good for debugging
	// engine.Debug(true) // Optional. Default: false

	// // Layout defines the variable name that is used to yield templates within layouts
	// engine.Layout("embed") // Optional. Default: "embed"

	// // Delims sets the action delimiters to the specified strings
	// engine.Delims("{{", "}}") // Optional. Default: engine delimiters

	// // AddFunc adds a function to the template's global function map.
	// engine.AddFunc("greet", func(name string) string {
	// 	return "Hello, " + name + "!"
	// })

	// // After you created your engine, you can pass it to Fiber's Views Engine
	// app := fiber.New(fiber.Config{
	// 	Views: engine,
	// })

	// app.Get("/", func(c *fiber.Ctx) error {
	// 	return c.Render("home", fiber.Map{
	// 		"Title": "Hello, World!",
	// 	})
	// })

	// app.Listen("pengun.linux.test:3000")
	return
}

func PageHomeHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, nil, "index.html")
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {
	oc.DeleteCookieSession(w, r)
	http.Redirect(w, r, "/", 307)
	return
}

func PageApiHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, nil, "main.html")
}

func pageHandler(w http.ResponseWriter, r *http.Request, data map[string]interface{}, filenames ...string) {
	dir := "tailwind/src"
	var files []string
	for _, file := range filenames {
		files = append(files, path.Join(dir, file))
	}
	files = append(files, path.Join(dir, "_colorspalette.html"))
	files = append(files, path.Join(dir, "_header.html"))
	files = append(files, path.Join(dir, "_sidebar.html"))
	files = append(files, path.Join(dir, "_scripts.html"))

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Session data
	email, isAuthenticated, err := oc.GetIdToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if data == nil {
		data = map[string]interface{}{}
	}

	data["email"] = email
	data["isAuthenticated"] = isAuthenticated

	log.Printf("Template / Session Data: %#v\n", data)

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AuthlinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authtype := vars["authtype"]
	service := vars["service"]
	authlink := oc.AuthLink(r, authtype, service)
	log.Println("AuthlinkHandler: " + authlink)
	fmt.Fprintln(w, authlink)
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	m, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	code := m.Get("code")
	state := m.Get("state")
	err = oc.ExchangeCode(w, r, code, state) //do not write to w before this call
	if err != nil {
		http.Error(w, "Exchange Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Fprintln(w, "Code: ", code, " Scope: ", scope)
	http.Redirect(w, r, "/page/api", 302)
}

//== API

func StravaGetAthleteHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://www.strava.com/api/v3/athlete"
	resp, err := oc.ApiRequest(w, r, oclient.STRAVA, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func StravaGetActivitiesHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=30"
	resp, err := oc.ApiRequest(w, r, oclient.STRAVA, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func LinkedinGetMeHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.linkedin.com/v2/me"
	resp, err := oc.ApiRequest(w, r, oclient.LINKEDIN, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func SpotifyGetMeHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.spotify.com/v1/me"
	resp, err := oc.ApiRequest(w, r, oclient.SPOTIFY, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func SpotifyGetNewReleasesHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.spotify.com/v1/browse/new-releases"
	resp, err := oc.ApiRequest(w, r, oclient.SPOTIFY, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func SpotifyPutRenameHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{
		"name":        "Updated Playlist Name",
		"description": "Updated playlist description",
		"public":      false,
	}

	url := "https://api.spotify.com/v1/playlists/2RmnrZSPoYtVyjou7DU8We"
	resp, err := oc.ApiRequest(w, r, oclient.SPOTIFY, "PUT", url, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func GithubGetUserHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.github.com/user"
	resp, err := oc.ApiRequest(w, r, oclient.GITHUB, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}
func FitbitGetUserHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.fitbit.com/1/user/-/profile.json"
	resp, err := oc.ApiRequest(w, r, oclient.FITBIT, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func FitbitGetHeartrateHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.fitbit.com/1/user/-/activities/heart/date/today/1d/1sec.json"
	resp, err := oc.ApiRequest(w, r, oclient.FITBIT, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func FitbitGetSleepHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.fitbit.com/1.2/user/-/sleep/date/2021-08-08.json?timezone=UTC"
	resp, err := oc.ApiRequest(w, r, oclient.FITBIT, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func OuraGetUserHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.ouraring.com/v1/userinfo"
	resp, err := oc.ApiRequest(w, r, oclient.OURA, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func OuraGetSleepHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.ouraring.com/v1/sleep"
	resp, err := oc.ApiRequest(w, r, oclient.OURA, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func OuraGetActivityHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.ouraring.com/v1/activity"
	resp, err := oc.ApiRequest(w, r, oclient.OURA, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}

func OuraGetReadinessHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://api.ouraring.com/v1/readiness"
	resp, err := oc.ApiRequest(w, r, oclient.OURA, "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(body))
}
