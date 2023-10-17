package main

import (
	"context"
	"database/sql"
	"encoding/gob"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

type Person struct {
	Id    int
	Name  string
	Email string
}

var (
	db           *sql.DB
	logLevel     string
	oauthConfig  *oauth2.Config
	sessionStore *sessions.CookieStore
	provider     *oidc.Provider
	logger       *slog.Logger
)

func init() {
	slog.Info("Initializing application")

	// Initialize the logger.
	logLevelEnvironmentVariable := os.Getenv("LOG_LEVEL")
	slog.Info("Environment Variables", "LOG_LEVEL", logLevelEnvironmentVariable)
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
	logger = slog.New(textHandler)

	logger.Debug(
		"Environment variables",
		"POSTGRES_URL", os.Getenv("POSTGRES_URL"),
		"OIDC_PROVIDER_URL", os.Getenv("OIDC_PROVIDER_URL"),
		"CLIENT_ID", os.Getenv("CLIENT_ID"),
		"CLIENT_SECRET", os.Getenv("CLIENT_SECRET"),
		"REDIRECT_URL", os.Getenv("REDIRECT_URL"),
		"SESSION_KEY", os.Getenv("SESSION_KEY"),
	)

	dbURL := os.Getenv("POSTGRES_URL")
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	db = conn

	provider, err = oidc.NewProvider(context.Background(), os.Getenv("OIDC_PROVIDER_URL"))
	if err != nil {
		panic(err)
	}

	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Endpoint:     provider.Endpoint(),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes:       []string{"openid", "profile", "email"},
	}

	sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	gob.Register(UserProfile{})
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "session")
	profile := session.Values["profile"]
	logger.Info("Handling index request")
<<<<<<< HEAD
	logger.Debug("Profile", "profile", profile)
=======
>>>>>>> 2af7e215a998a67f22b8be3ca5158a02994f0d9b
	t := template.Must(template.ParseFiles("index.html", "contact.html"))
	t.Execute(w, profile)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	webErrorLogger := slog.NewLogLogger(handler, slog.LevelError)
	logger.Info("Starting server on port 3000")

	http.Handle("/", LoggingMiddleware(http.HandlerFunc(IndexHandler)))
	http.Handle("/contact", LoggingMiddleware((AuthMiddleware(http.HandlerFunc(Handler)))))
	http.Handle("/login", LoggingMiddleware(http.HandlerFunc(LoginHandler)))
	http.Handle("/callback", LoggingMiddleware(http.HandlerFunc(CallbackHandler)))
	http.Handle("/logout", LoggingMiddleware(http.HandlerFunc(LogoutHandler)))

	server := http.Server{
		ErrorLog: webErrorLogger,
		Addr:     ":3000",
	}
	server.ListenAndServe()
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
