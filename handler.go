package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "session")
	profile := session.Values["profile"]
	accessToken := session.Values["access_token"]
	slog.Debug("Access Token", "accessToken", accessToken)
	slog.Debug("Profile", "profile", profile)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		slog.Info("Handling contact request")
		contactID, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, "Invalid contact ID", http.StatusBadRequest)
			return
		}

		slog.Info("Query the database to retrieve the person by ID.")
		person, err := getPersonByID(contactID)
		if err != nil {
			http.Error(w, "Failed to retrieve contact", http.StatusInternalServerError)
			return
		}

		if len(parts) == 4 {
			slog.Info("Handling contact request with action")
			action := parts[3]
			if action == "edit" {
				slog.Info("Handling contact request with action edit")
				t := template.Must(template.ParseFiles("edit.html"))
				t.ExecuteTemplate(w, "person", person)
				return
			}
		}

		if r.Method == "PUT" {
			slog.Info("Handling contact request with PUT")
			err := updatePerson(person, r.FormValue("name"), r.FormValue("email"))
			if err != nil {
				http.Error(w, "Failed to update contact", http.StatusInternalServerError)
				return
			}
		}

		t := template.Must(template.ParseFiles("row.html"))
		t.ExecuteTemplate(w, "person", person)
		return
	}

	slog.Info("Handling /contact/ request with no ID")
	if r.Method == "POST" {
		slog.Info("Handling contact request with POST")
	}
	slog.Info("Query the database to retrieve all people.")
	people, err := getAllPeople()

	if err != nil {
		slog.Error("Failed to retrieve people", "error", err)
		http.Error(w, "Failed to retrieve people", http.StatusInternalServerError)
		return
	}
	slog.Debug("People", "people", len(people))

	t := template.Must(template.ParseFiles("contact.html", "row.html"))
	t.Execute(w, people)
	return

}
