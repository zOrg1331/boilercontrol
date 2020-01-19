package main

import (
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"regexp"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

func boilerOnHandler(w http.ResponseWriter, r *http.Request) {
	exec.Command("/usr/bin/boiler_on")
	http.Redirect(w, r, "/", http.StatusFound)
}

func boilerOffHandler(w http.ResponseWriter, r *http.Request) {
	exec.Command("/usr/bin/boiler_off")
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

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/boiler/on", makeHandler(boilerOnHandler))
	http.HandleFunc("/boiler/off", makeHandler(boilerOffHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
