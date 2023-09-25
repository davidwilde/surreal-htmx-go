package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Person struct {
	Id    int
	Name  string
	Email string
}

var db *sql.DB

func init() {
	// Initialize the database connection.
	dbURL := os.Getenv("POSTGRES_URL")
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	db = conn
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8080", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 && parts[1] == "contact" {
		contactID, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, "Invalid contact ID", http.StatusBadRequest)
			return
		}

		// Query the database to retrieve the person by ID.
		person, err := getPersonByID(contactID)
		if err != nil {
			http.Error(w, "Failed to retrieve contact", http.StatusInternalServerError)
			return
		}

		if len(parts) == 4 {
			action := parts[3]
			if action == "edit" {
				t := template.Must(template.ParseFiles("edit.html"))
				t.ExecuteTemplate(w, "person", person)
				return
			}
		}

		if r.Method == "PUT" {
			// Handle the PUT request to update person data in the database.
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

	// Handle the index page with a list of people.
	people, err := getAllPeople()
	if err != nil {
		http.Error(w, "Failed to retrieve people", http.StatusInternalServerError)
		return
	}

	t := template.Must(template.ParseFiles("index.html", "row.html"))
	t.Execute(w, people)
}

func getPersonByID(id int) (Person, error) {
	var person Person
	err := db.QueryRow("SELECT id, name, email FROM people WHERE id = $1", id).Scan(&person.Id, &person.Name, &person.Email)
	if err != nil {
		return person, err
	}
	return person, nil
}

func updatePerson(person Person, newName, newEmail string) error {
	_, err := db.Exec("UPDATE people SET name = $1, email = $2 WHERE id = $3", newName, newEmail, person.Id)
	return err
}

func getAllPeople() ([]Person, error) {
	rows, err := db.Query("SELECT id, name, email FROM people")
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
