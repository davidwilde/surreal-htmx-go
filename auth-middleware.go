package main

import (
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyAccessTokenWithJWK(accessToken string, jwksURL string) (*jwt.Token, error) {
	// Create the keyfunc options. Use an error handler that logs. Timeout the initial JWKS refresh request after 10
	// seconds. This timeout is also used to create the initial context.Context for keyfunc.Get.
	options := keyfunc.Options{
		RefreshTimeout: time.Second * 10,
		RefreshErrorHandler: func(err error) {
			logger.Error("There was an error with the jwt.Keyfunc", "error", err.Error())
			return
		},
	}

	// Create the JWKS from the resource at the given URL.
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		logger.Debug("Failed to create JWKS from resource at the given URL.", "Error", err.Error())
		return nil, err
	}

	// Parse the JWT.
	token, err := jwt.Parse(accessToken, jwks.Keyfunc)
	if err != nil {
		logger.Debug("Failed to parse the JWT", "Error", err.Error())
		return nil, err
	}

	// Check if the token is valid.
	if !token.Valid {
		logger.Debug("The token is invalid", "error", err.Error())
		return nil, err
	}
	logger.Debug("The token is valid.")
	return token, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("AuthMiddleware")

		// Retrieve the access token from the session
		session, err := sessionStore.Get(r, "session")
		if (err != nil) || (session == nil) {
			logger.Warn("error getting session", "error", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		accessToken := session.Values["access_token"]
		tokenString, ok := accessToken.(string)
		if !ok {
			logger.Warn("error getting access token", "error", ok)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !ok {
			logger.Warn("error getting access token", "error", ok)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if tokenString == "" {
			http.Error(w, "Access token missing", http.StatusUnauthorized)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		_, err = VerifyAccessTokenWithJWK(tokenString, "https://hobby.kinde.com/.well-known/jwks")
		if err != nil {
			logger.Warn("error verifying token", "error", err, "token", tokenString)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Access token is valid; you can proceed to the protected endpoint.
		// You can also add claims validation or other checks here if needed.

		next.ServeHTTP(w, r)
	})
}
