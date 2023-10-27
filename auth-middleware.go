package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

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
		refreshToken := session.Values["refresh_token"]
		accessTokenString, ok := accessToken.(string)
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

		if accessTokenString == "" {
			http.Error(w, "Access token missing", http.StatusUnauthorized)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		refreshTokenString, ok := refreshToken.(string)
		if !ok {
			logger.Warn("error getting refresh token", "error", ok)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if refreshTokenString == "" {
			http.Error(w, "Refresh token missing", http.StatusUnauthorized)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		_, err = VerifyAccessTokenWithJWK(accessTokenString, "https://hobby.kinde.com/.well-known/jwks", refreshTokenString)
		if err != nil {
			logger.Warn("error verifying token", "error", err, "token", accessTokenString)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Access token is valid; you can proceed to the protected endpoint.
		// You can also add claims validation or other checks here if needed.

		next.ServeHTTP(w, r)
	})
}

func VerifyAccessTokenWithJWK(accessToken string, jwksURL string, refreshToken string) (*jwt.Token, error) {
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

		// Check if there is a refresh token.
		if refreshToken == "" {
			return nil, errors.New("token is expired and no refresh token provided")
		}

		// Use the refresh token to obtain a new access token.
		newAccessToken, err := refreshAccessToken(refreshToken)
		if err != nil {
			return nil, err
		}

		// Parse the new access token.
		token, err = jwt.Parse(newAccessToken, jwks.Keyfunc)
		if err != nil {
			logger.Debug("Failed to parse the new access token", "Error", err.Error())
			return nil, err
		}
	}

	return token, nil
}

func refreshAccessToken(refreshToken string) (string, error) {
	// Create a new HTTP request to the token endpoint.
	req, err := http.NewRequest("POST", oauthConfig.Endpoint.TokenURL, nil)
	if err != nil {
		logger.Debug("Failed to create HTTP request", "Error", err.Error())
		return "", err
	}

	// Set the request headers.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set the request body.
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", refreshToken)
	form.Add("client_id", oauthConfig.ClientID)
	form.Add("client_secret", oauthConfig.ClientSecret)
	req.Body = ioutil.NopCloser(strings.NewReader(form.Encode()))

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Debug("Failed to send HTTP request", "Error", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Debug("Failed to read response body", "Error", err.Error())
		return "", err
	}

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		logger.Debug("Failed to refresh access token", "StatusCode", resp.StatusCode, "Body", string(body))
		return "", errors.New("failed to refresh access token")
	}

	// Parse the response body.
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		logger.Debug("Failed to parse token response", "Error", err.Error())
		return "", err
	}

	return tokenResponse.AccessToken, nil
}
