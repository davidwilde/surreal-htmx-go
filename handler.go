package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type details = struct {
	Profile UserProfile
	People  []Person
}

func Handler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		logger.Info("Handling contact request")
		contactID, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, "Invalid contact ID", http.StatusBadRequest)
			return
		}

		logger.Info("Query the database to retrieve the person by ID.")
		person, err := getPersonByID(contactID)
		if err != nil {
			http.Error(w, "Failed to retrieve contact", http.StatusInternalServerError)
			return
		}

		if len(parts) == 4 {
			logger.Info("Handling contact request with action")
			action := parts[3]
			if action == "edit" {
				logger.Info("Handling contact request with action edit")
				t := template.Must(template.ParseFiles("edit.html"))
				t.ExecuteTemplate(w, "person", person)
				return
			}
		}

		if r.Method == "PUT" {
			logger.Info("Handling contact request with PUT")
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

	logger.Info("Handling /contact/ request with no ID")
	if r.Method == "POST" {
		logger.Info("Handling contact request with POST")
	}
	logger.Info("Query the database to retrieve all people.")
	people, err := getAllPeople()

	if err != nil {
		logger.Error("Failed to retrieve people", "error", err)
		http.Error(w, "Failed to retrieve people", http.StatusInternalServerError)
		return
	}
	logger.Debug("People", "people", len(people))

	t := template.Must(template.ParseFiles("base.html", "contact.html", "row.html"))
	profile := getProfile(r)
	details := details{
		Profile: profile,
		People:  people,
	}

	t.Execute(w, details)
	return

}

func getPersonByID(id int) (Person, error) {
	var person Person
	logger.Debug("Getting person by ID", "id", id, "query", "SELECT id, name, email FROM people WHERE id = $1")
	err := db.QueryRow("SELECT id, name, email FROM people WHERE id = $1", id).Scan(&person.Id, &person.Name, &person.Email)
	if err != nil {
		return person, err
	}
	return person, nil
}

func updatePerson(person Person, newName, newEmail string) error {
	query := "UPDATE people SET name = $1, email = $2 WHERE id = $3"
	logger.Debug("Updating person", "query", query, "$1", newName, "$2", newEmail, "$3", person.Id)
	_, err := db.Exec(query, newName, newEmail, person.Id)
	return err
}

func getAllPeople() ([]Person, error) {
	query := "SELECT id, name, email FROM people"
	logger.Debug("Getting all people", "query", query)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []Person
	for rows.Next() {
		var person Person
		err := rows.Scan(&person.Id, &person.Name, &person.Email)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}

	return people, nil
}
