package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"
)

func RenderError(w http.ResponseWriter, r *http.Request, message string, statuscode int, backLink string) {
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == sql.ErrNoRows {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})
		} else if err != nil {
			RenderError(w, r, "Database Error", http.StatusInternalServerError, backLink)
			return
		}
	}
	w.WriteHeader(statuscode)
	tmpl, err := template.ParseFiles("./templates/error.html")
	if err != nil {
		http.Error(w, "Error parsing error template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"ErrorMessage": message,
		"StatusCode":   statuscode,
		"IsLoggedIn":   userID != "",
		"BackLink":     backLink, // Pass the back link to the template
	})
}