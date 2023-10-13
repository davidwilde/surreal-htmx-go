package main

import (
	"log/slog"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling login request")
	url := oauthConfig.AuthCodeURL("WeHH_yy2irpl8UY")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
