package main

import "net/http"

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Logout handler out")
	session, _ := sessionStore.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
