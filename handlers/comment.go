package handlers

import (
	"database/sql"
	"net/http"
	"time"
)

func CommentHandler(w http.ResponseWriter, r *http.Request) {
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
			RenderError(w, r, "Database Error", http.StatusInternalServerError, "/post")
			return
		}
	}
	if r.Method == http.MethodPost {
		postID := r.FormValue("post_id")
		comment := r.FormValue("comment")
		// Use the actual user ID retrieved from the session
		// userID is already set above

		// Insert the comment into the database
		_, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, comment)
		if err != nil {
			http.Error(w, "Error posting comment", http.StatusInternalServerError)
			return
		}

		// Redirect back to the post or home page
		http.Redirect(w, r, "/post", http.StatusSeeOther)
		return
	}
}