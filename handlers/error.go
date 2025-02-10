 package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
)

// RenderError renders an error page with the given message and status code
func RenderError(w http.ResponseWriter, r *http.Request, message string, statusCode int, redirectPath string) {
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
			log.Printf("Database error in RenderError: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(statusCode)
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		log.Printf("Error parsing error template: %v", err)
		http.Error(w, "Error parsing error template", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ErrorMessage": message,
		"StatusCode":   statusCode,
		"IsLoggedIn":   userID != "",
		"RedirectPath": redirectPath,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing error template: %v", err)
		http.Error(w, "Error rendering error page", http.StatusInternalServerError)
		return
	}
}

// HandleDatabaseError handles database-related errors and logs them appropriately
func HandleDatabaseError(w http.ResponseWriter, r *http.Request, err error, redirectPath string) {
	if err != nil {
		log.Printf("Database error: %v", err)
		RenderError(w, r, "Database Error", http.StatusInternalServerError, redirectPath)
		return
	}
}
