package main

import "net/http"
import "net/url"
import "os"

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Logout handler")
	session, _ := sessionStore.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
    url := os.Getenv("OIDC_PROVIDER_URL") + "/logout?redirect_uri=" + url.PathEscape(getLogoutRedirectURL())
    logger.Debug("Redirecting to", "url", url)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
