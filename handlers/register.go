package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			log.Printf("Error parsing register template: %v", err)
			RenderError(w, r, "Error loading page", http.StatusInternalServerError, "/")
			return
		}

		if err = tmpl.Execute(w, nil); err != nil {
			log.Printf("Error executing register template: %v", err)
			RenderError(w, r, "Error rendering page", http.StatusInternalServerError, "/")
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		email := strings.TrimSpace(r.FormValue("email"))
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate input
		if email == "" || username == "" || password == "" || confirmPassword == "" {
			RenderError(w, r, "All fields are required", http.StatusBadRequest, "/register")
			return
		}

		// Validate email format
		emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !emailRegex.MatchString(email) {
			RenderError(w, r, "Invalid email format", http.StatusBadRequest, "/register")
			return
		}

		// Validate username length
		if len(username) < 3 {
			RenderError(w, r, "Username must be at least 3 characters long", http.StatusBadRequest, "/register")
			return
		}

		// Check password length
		if len(password) < 6 {
			RenderError(w, r, "Password must be at least 6 characters long", http.StatusBadRequest, "/register")
			return
		}

		// Check if passwords match
		if password != confirmPassword {
			RenderError(w, r, "Passwords do not match", http.StatusBadRequest, "/register")
			return
		}

		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/register")
			return
		}
		defer tx.Rollback()

		// Check if email already exists
		var exists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
		if err != nil {
			log.Printf("Error checking email existence: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/register")
			return
		}
		if exists {
			RenderError(w, r, "Email already registered", http.StatusConflict, "/register")
			return
		}

		// Check if username already exists
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
		if err != nil {
			log.Printf("Error checking username existence: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/register")
			return
		}
		if exists {
			RenderError(w, r, "Username already taken", http.StatusConflict, "/register")
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/register")
			return
		}

		// Generate user ID
		userID := uuid.New().String()

		// Create user
		_, err = tx.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)",
			userID, email, username, hashedPassword)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			RenderError(w, r, "Error creating account", http.StatusInternalServerError, "/register")
			return
		}

		// Create session for the new user
		sessionID := uuid.New().String()
		_, err = tx.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, userID)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			RenderError(w, r, "Error creating session", http.StatusInternalServerError, "/register")
			return
		}

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			RenderError(w, r, "Error creating account", http.StatusInternalServerError, "/register")
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

		// Redirect to home page after successful registration
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	RenderError(w, r, "Method Not Allowed", http.StatusMethodNotAllowed, "/")
}
