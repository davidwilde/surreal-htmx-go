package main

import "net/http"
import "net/url"
import "os"

// LogoutHandler handles the /logout/ request
// It invalidates the session and redirects the user to the logout endpoint of the OIDC provider
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Logout handler")
	session, _ := sessionStore.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	url := os.Getenv("OIDC_PROVIDER_URL") + "/logout?redirect=" + url.PathEscape(getLogoutRedirectURL())
	logger.Debug("Redirecting to", "url", url)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
