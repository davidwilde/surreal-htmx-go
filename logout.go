package main

import "net/http"

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Logout handler")
	session, _ := sessionStore.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
    url := oauthConfig.Endpoint.AuthURL + "/logout"
	http.Redirect(w, r, url, http.StatusSeeOther)
}
