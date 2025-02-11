package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == sql.ErrNoRows {
			// Clear the invalid session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})
		} else if err != nil {
			log.Printf("Database error: %v", err)
			RenderError(w, r, "Database Error", http.StatusInternalServerError, "/post")
			return
		}
	}

	// Handle POST request (create a new post)
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["category"] // Get multiple categories

		// Validate input
		if title == "" || content == "" || len(categories) == 0 {
			RenderError(w, r, "Title, content, and at least one category are required", http.StatusBadRequest, "/post")
			return
		}

		// Insert the new post into the database
		result, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
		if err != nil {
			log.Printf("Error creating post: %v", err)
			RenderError(w, r, "Error creating post", http.StatusInternalServerError, "/post")
			return
		}

		// Get the ID of the newly created post
		postID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error retrieving post ID: %v", err)
			RenderError(w, r, "Error retrieving post ID", http.StatusInternalServerError, "/post")
			return
		}

		// Insert categories into the database
		for _, category := range categories {
			_, err = db.Exec("INSERT INTO post_categories (post_id, category) VALUES (?, ?)", postID, category)
			if err != nil {
				log.Printf("Error inserting category: %v", err)
				RenderError(w, r, "Error inserting categories", http.StatusInternalServerError, "/post")
				return
			}
		}

		// Redirect to the posts page after successful creation
		http.Redirect(w, r, "/post", http.StatusSeeOther)
		return
	}

	// Handle GET request (fetch and display posts)
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
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/post")
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
			RenderError(w, r, "Error scanning posts", http.StatusInternalServerError, "/post")
			return
		}
		if categories.Valid {
			post.Categories = categories.String
		} else {
			post.Categories = ""
		}
		posts = append(posts, post)
	}

	// Render the template with the posts data
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		RenderError(w, r, "Error parsing template", http.StatusInternalServerError, "/post")
		return
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": userID != "",
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		RenderError(w, r, "Error rendering page", http.StatusInternalServerError, "/post")
		return
	}
}