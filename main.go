package main

import (
	"database/sql"
	"html/template"
	"log/slog"
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
var logLevel = os.Getenv("LOG_LEVEL")

func init() {

	// Initialize the logger.
	logLevelEnvironmentVariable := os.Getenv("LOG_LEVEL")
	logLevel := slog.LevelInfo

	switch logLevelEnvironmentVariable {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo

	}

	opts := &slog.HandlerOptions{
		Level: slog.Level(logLevel),
	}
	textHandler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(textHandler)
	slog.SetDefault(logger)

	dbURL := os.Getenv("POSTGRES_URL")
	slog.Debug("POSTGRES_URL: %s", dbURL)
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	db = conn
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	webErrorLogger := slog.NewLogLogger(handler, slog.LevelError)
	slog.Info("Starting server on port 80")
	http.HandleFunc("/", Handler)
	server := http.Server{
		ErrorLog: webErrorLogger,
	}
	server.ListenAndServe()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 && parts[1] == "contact" {
		slog.Info("Handling contact request")
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

	// Handle the index page with a list of people.
	slog.Info("Handle the index page with a list of people")
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
	slog.Debug("Getting person by ID", "id", id, "query", "SELECT id, name, email FROM people WHERE id = $1")
	err := db.QueryRow("SELECT id, name, email FROM people WHERE id = $1", id).Scan(&person.Id, &person.Name, &person.Email)
	if err != nil {
		return person, err
	}
	return person, nil
}

func updatePerson(person Person, newName, newEmail string) error {
	query := "UPDATE people SET name = $1, email = $2 WHERE id = $3"
	slog.Debug("Updating person", "query", query, "$1", newName, "$2", newEmail, "$3", person.Id)
	_, err := db.Exec(query, newName, newEmail, person.Id)
	return err
}

func getAllPeople() ([]Person, error) {
	query := "SELECT id, name, email FROM people"
	slog.Debug("Getting all people", "query", query)
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
