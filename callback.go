package main

import (
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

type UserProfile struct {
	// Define the fields you need in the session data
	Name  string
	Email string
	// Add more fields as needed
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling callback request")
	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	rawIdToken, ok := token.Extra("id_token").(string)

	if !ok {
		http.Error(w, "Failed to get id token", http.StatusInternalServerError)
		return
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID})
	idToken, err := verifier.Verify(r.Context(), rawIdToken)
	if err != nil {
		http.Error(w, "Failed to verify id token", http.StatusInternalServerError)
		return
	}
	var profile UserProfile
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, "Failed to get claims", http.StatusInternalServerError)
		return
	}

	// Save the profile data in the session
	session, err := sessionStore.Get(r, "session")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	// Save the UserProfile in the session
	session.Values["profile"] = profile
	session.Values["access_token"] = token.AccessToken

	if err := session.Save(r, w); err != nil {
		logger.Debug("Failed to save session", "error", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to another page or render a template with the profile data
	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}