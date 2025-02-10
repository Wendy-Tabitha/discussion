package handlers

import (
	"log"
	"net/http"
	"time"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, r, "Method Not Allowed", http.StatusMethodNotAllowed, "/")
		return
	}

	// Get the session cookie
	sessionCookie, err := r.Cookie("session_id")
	if err == nil {
		// Delete the session from the database
		_, err = db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionCookie.Value)
		if err != nil {
			log.Printf("Error deleting session: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/")
			return
		}

		// Expire the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	}

	// Redirect to home page after successful logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
