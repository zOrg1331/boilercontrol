package main

import (
	"crypto/subtle"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

func boilerOnHandler(w http.ResponseWriter, r *http.Request) {
	exec.Command("/usr/bin/boiler_on").Output()
	http.Redirect(w, r, "/", http.StatusFound)
}

func boilerOffHandler(w http.ResponseWriter, r *http.Request) {
	exec.Command("/usr/bin/boiler_off").Output()
	http.Redirect(w, r, "/", http.StatusFound)
}

var templates = template.Must(template.ParseFiles("./templates/index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^(/|/static/[a-zA-Z0-9]+|/boiler/(on|off))$")

func makeHandler(fn func(http.ResponseWriter, *http.Request), username string, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userToCheck, passToCheck, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(userToCheck), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(passToCheck), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="zHome"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	authData, err := ioutil.ReadFile("./auth")
	if err != nil {
		panic(err)
	}
	creds := strings.Split(strings.TrimSuffix(string(authData), "\n"), ":")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", makeHandler(indexHandler, creds[0], creds[1]))
	http.HandleFunc("/boiler/on", makeHandler(boilerOnHandler, creds[0], creds[1]))
	http.HandleFunc("/boiler/off", makeHandler(boilerOffHandler, creds[0], creds[1]))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
