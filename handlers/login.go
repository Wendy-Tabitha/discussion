package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			RenderError(w, "Error parsing file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var user User
		err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&user.ID, &user.Password)
		if err != nil {
			RenderError(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			RenderError(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Check if the user is already logged in
		var existingSessionID string
		err = db.QueryRow("SELECT session_id FROM sessions WHERE user_id = ?", user.ID).Scan(&existingSessionID)
		if err == nil {
			_, err = db.Exec("DELETE FROM sessions WHERE user_id = ?", user.ID)
			if err != nil {
				RenderError(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		}
		sessionID := uuid.New().String()

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   sessionID,                        // Use the UUID as the session ID
			Expires: time.Now().Add(24 * time.Hour), // Set expiration
		})

		// Store session in the database
		_, err = db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, user.ID)
		if err != nil {
			RenderError(w, "Error creating session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/post", http.StatusSeeOther)
		return
	} else {
		RenderError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}
