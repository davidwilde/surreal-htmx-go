package main

import (
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling login request")
	url := oauthConfig.AuthCodeURL("WeHH_yy2irpl8UY")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
