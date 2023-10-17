// A middleware that will allow us to specify log levels from a header.
// Our logging is managed with logger.
// When the header is not present, it will default to the level specified.
// The header looks like x-log-level: debug

package main

import (
	"log/slog"
	"net/http"
	"os"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var level slog.Level
		if r.Header.Get("x-log-level") != "" {
			headerLevel := r.Header.Get("x-log-level")
			levelBytes := []byte(headerLevel)
			err := level.UnmarshalText(levelBytes)
			if err != nil {
				slog.Error("Error parsing log level", "Error", err)
			} else {
				slog.Info("Setting log level", "Level", level)
				logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
			}
		}

		next.ServeHTTP(w, r)
	})
}
