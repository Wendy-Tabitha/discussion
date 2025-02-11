package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	isLoggedIn := false // Flag to check if the user is logged in

	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == nil {
			isLoggedIn = true // User is logged in
		} else if err == sql.ErrNoRows {
			// Clear the invalid session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})
		} else {
			log.Printf("Database error: %v", err)
			RenderError(w, r, "Database Error", http.StatusInternalServerError, "/")
			return
		}
	}

	// Query to fetch all posts along with the user's name, creation time, like count, and dislike count
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, 
		u.username, p.created_at, 
		COALESCE(l.like_count, 0) AS like_count,
		COALESCE(l.dislike_count, 0) AS dislike_count
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN (
			SELECT post_id, 
			COUNT(CASE WHEN is_like = 1 THEN 1 END) AS like_count,
			COUNT(CASE WHEN is_like = 0 THEN 1 END) AS dislike_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		GROUP BY p.id, p.title, p.content, u.username, p.created_at
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/")
		return
	}
	defer rows.Close()

	// Parse the rows into a slice of Post structs
	var posts []Post
	for rows.Next() {
		var post Post
		var categories sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Username, &post.CreatedAt, &post.LikeCount, &post.DislikeCount)
		if err != nil {
			log.Printf("Error scanning post: %v", err)
			RenderError(w, r, "Error scanning posts", http.StatusInternalServerError, "/")
			return
		}
		if categories.Valid {
			post.Categories = categories.String
		} else {
			post.Categories = ""
		}
		posts = append(posts, post)
	}

	// Render the home template with the posts data
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		RenderError(w, r, "Error parsing template", http.StatusInternalServerError, "/")
		return
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": isLoggedIn,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		RenderError(w, r, "Error rendering page", http.StatusInternalServerError, "/")
		return
	}
}