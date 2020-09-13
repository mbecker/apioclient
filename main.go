package main

import (
	"fmt"
	"github.com/exyzzy/oclient/oclient"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
)

func main() {
	err := oclient.InitOclient()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r := mux.NewRouter()
	r.HandleFunc("/", PageHomeHandler)
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
	http.Handle("/", r)
	fmt.Println(">>>>>>> OClient started at:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
	return
}

func PageHomeHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, nil, "templates", "home.html")
}

func PageApiHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, nil, "templates", "api.html")
}

func pageHandler(w http.ResponseWriter, r *http.Request, data interface{}, dir string, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, path.Join(dir, file))
	}
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AuthlinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authtype := vars["authtype"]
	service := vars["service"]
	authlink := oclient.AuthLink(r, authtype, service)
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
	err = oclient.ExchangeCode(w, r, code, state) //do not write to w before this call
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
	resp, err := oclient.ApiRequest(w, r, oclient.STRAVA, "GET", url, nil)
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
	resp, err := oclient.ApiRequest(w, r, oclient.STRAVA, "GET", url, nil)
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
	resp, err := oclient.ApiRequest(w, r, oclient.LINKEDIN, "GET", url, nil)
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
	resp, err := oclient.ApiRequest(w, r, oclient.SPOTIFY, "GET", url, nil)
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
	resp, err := oclient.ApiRequest(w, r, oclient.SPOTIFY, "GET", url, nil)
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
	resp, err := oclient.ApiRequest(w, r, oclient.SPOTIFY, "PUT", url, data)
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
	resp, err := oclient.ApiRequest(w, r, oclient.GITHUB, "GET", url, nil)
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
