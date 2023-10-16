package main

import (
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Replace 'YOUR_OIDC_PROVIDER_URL' with the URL of your OIDC provider.
		verifier := provider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID})

		// Retrieve the access token from the session
		session, _ := sessionStore.Get(r, "session")
		tokenString := session.Values["access_token"].(string)

		if tokenString == "" {
			http.Error(w, "Access token missing", http.StatusUnauthorized)
			return
		}

		_, err := verifier.Verify(r.Context(), tokenString)
		if err != nil {
			logger.Error("error verifying token", "error", err, "token", tokenString)
			http.Error(w, "Invalid access token", http.StatusUnauthorized)
			return
		}

		// Access token is valid; you can proceed to the protected endpoint.
		// You can also add claims validation or other checks here if needed.

		next.ServeHTTP(w, r)
	})
}
