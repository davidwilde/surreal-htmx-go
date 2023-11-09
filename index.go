package main

import (
	"html/template"
	"net/http"
)

// IndexHandler handles the / request
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	profile := getProfile(r)
	details := struct {
		Profile UserProfile
	}{
		Profile: profile,
	}
	t := template.Must(template.ParseFiles("base.html", "index.html"))

	t.Execute(w, details)
}
