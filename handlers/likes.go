package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// LikeResponse is the response for like/dislike actions
type LikeResponse struct {
	LikeCount    int   `json:"likeCount"`
	DislikeCount int   `json:"dislikeCount"`
	UserLiked    *bool `json:"userLiked"`
}

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve session cookie
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "You must be logged in to like or dislike a post", http.StatusUnauthorized)
		return
	}

	// Get user ID from session
	var userID string
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
		http.Error(w, "You must be logged in to like or dislike a post", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Read form values
	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	isLike := r.FormValue("is_like") == "true"

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if the user has already liked/disliked this post
	var existingIsLike sql.NullBool
	err = tx.QueryRow("SELECT is_like FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingIsLike)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Error checking like status", http.StatusInternalServerError)
		return
	}

	if existingIsLike.Valid {
		if existingIsLike.Bool == isLike {
			// Remove the like/dislike if clicking the same button
			_, err = tx.Exec("DELETE FROM likes WHERE post_id = ? AND user_id = ?",
				postID, userID)
		} else {
			// Update from like to dislike or vice versa
			_, err = tx.Exec("UPDATE likes SET is_like = ? WHERE post_id = ? AND user_id = ?",
				isLike, postID, userID)
		}
	} else {
		// Add new like/dislike
		_, err = tx.Exec("INSERT INTO likes (post_id, user_id, is_like) VALUES (?, ?, ?)",
			postID, userID, isLike)
	}

	if err != nil {
		http.Error(w, "Error updating like status", http.StatusInternalServerError)
		return
	}

	// Get updated counts and user's current like status
	var response LikeResponse
	var userLiked sql.NullBool
	err = tx.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 1),
			(SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 0),
			CASE 
				WHEN EXISTS (SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)
				THEN (SELECT is_like FROM likes WHERE post_id = ? AND user_id = ?)
				ELSE NULL 
			END
	`, postID, postID, postID, userID, postID, userID).Scan(&response.LikeCount, &response.DislikeCount, &userLiked)

	if err != nil {
		http.Error(w, "Error getting updated counts", http.StatusInternalServerError)
		return
	}

	if userLiked.Valid {
		response.UserLiked = &userLiked.Bool
	} else {
		response.UserLiked = nil
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
