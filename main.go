package main

import (
	"context"
	"database/sql"
	"encoding/gob"
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
	db            *sql.DB
	logLevel      string
	oauthConfig   *oauth2.Config
	sessionStore  *sessions.CookieStore
	provider      *oidc.Provider
	logger        *slog.Logger
	requestLogger *slog.Logger
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
		"CALLBACK_URL", getRedirectURL(),
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
		RedirectURL:  getRedirectURL(),
		Scopes:       []string{"openid", "profile", "email", "offline"},
	}

	sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	gob.Register(UserProfile{})
}

func getLogoutRedirectURL() string {
	url := os.Getenv("RENDER_EXTERNAL_URL")
	if url == "" {
		url = "http://localhost:3000"
	}
	return url
}

func getRedirectURL() string {
	url := os.Getenv("RENDER_EXTERNAL_URL")
	if url == "" {
		url = "http://localhost:3000"
	}
	return url + "/callback"
}

func getProfile(r *http.Request) UserProfile {
	session, err := sessionStore.Get(r, "session")
	if err != nil || len(session.Values) == 0 {
		return UserProfile{}
	}
	profile, exists := session.Values["profile"]
	if !exists {
		return UserProfile{}
	}

	profileValue, ok := profile.(UserProfile)
	if !ok {
		logger.Debug("Profile did not parse into User Profile", "profile", profile)
		return UserProfile{}
	}
	logger.Debug("Profile", "profile", profileValue)
	return profileValue
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	webErrorLogger := slog.NewLogLogger(handler, slog.LevelError)
	logger.Info("Starting server on port 3000")

	http.Handle("/contact/", LoggingMiddleware((AuthMiddleware(http.HandlerFunc(Handler)))))
	http.Handle("/contact", LoggingMiddleware((AuthMiddleware(http.HandlerFunc(Handler)))))
	http.Handle("/login", LoggingMiddleware(http.HandlerFunc(LoginHandler)))
	http.Handle("/callback", LoggingMiddleware(http.HandlerFunc(CallbackHandler)))
	http.Handle("/logout", LoggingMiddleware(http.HandlerFunc(LogoutHandler)))
	http.Handle("/", LoggingMiddleware(http.HandlerFunc(IndexHandler)))

	server := http.Server{
		ErrorLog: webErrorLogger,
		Addr:     ":3000",
	}
	server.ListenAndServe()
}
