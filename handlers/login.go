package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show login form
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			log.Printf("Error parsing login template: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/")
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Error executing login template: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/")
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			RenderError(w, r, "Invalid email or password", http.StatusBadRequest, "/login")
			return
		}

		// Get user from database
		var user User
		var hashedPassword string
		err := db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &hashedPassword)
		if err == sql.ErrNoRows {
			log.Printf("Login attempt failed for email %s: user not found", email)
			RenderError(w, r, "Invalid email or password", http.StatusUnauthorized, "/login")
			return
		} else if err != nil {
			log.Printf("Database error during login: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/login")
			return
		}

		// Compare passwords
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			log.Printf("Login attempt failed for email %s: invalid password", email)
			RenderError(w, r, "Invalid email or password", http.StatusUnauthorized, "/login")
			return
		}

		// Check if the user is already logged in
		var existingSessionID string
		err = db.QueryRow("SELECT session_id FROM sessions WHERE user_id = ?", user.ID).Scan(&existingSessionID)
		if err == nil {
			// Delete existing session
			_, err = db.Exec("DELETE FROM sessions WHERE user_id = ?", user.ID)
			if err != nil {
				log.Printf("Error deleting existing session: %v", err)
				RenderError(w, r, "Server Error", http.StatusInternalServerError, "/login")
				return
			}
		}

		// Create new session
		sessionID := uuid.New().String()
		_, err = db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, user.ID)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/login")
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// Redirect to home page after successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Handle invalid methods
	RenderError(w, r, "Method Not Allowed", http.StatusMethodNotAllowed, "/")
}
