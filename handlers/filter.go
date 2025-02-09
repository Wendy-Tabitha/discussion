package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"
)

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	isLoggedIn := false // Flag to check if the user is logged in

	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == nil {
			isLoggedIn = true // User is logged in
		} else if err == sql.ErrNoRows {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})
		} else {
			RenderError(w, r, "Database Error", http.StatusInternalServerError, "/")
			return
		}
	}

	category := r.URL.Query().Get("category") // Get the category from the query parameters

	// Query to fetch posts based on the selected category
	var query string
	if category == "all" || category == "" {
		query = `SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, u.username, 
		p.created_at FROM posts p JOIN users u ON p.user_id = u.id LEFT JOIN post_categories pc ON p.id = pc.post_id 
		GROUP BY p.id ORDER BY p.created_at DESC`
	} else {
		query = `SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, u.username, 
		p.created_at FROM posts p JOIN users u ON p.user_id = u.id LEFT JOIN post_categories pc ON p.id = pc.post_id 
		WHERE pc.category = ? GROUP BY p.id ORDER BY p.created_at DESC`
	}

	var rows *sql.Rows
	if category == "all" || category == "" {
		rows, err = db.Query(query)
	} else {
		rows, err = db.Query(query, category)
	}
	if err != nil {
		// Redirect based on login status
		if isLoggedIn {
			RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/post")
		} else {
			RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/")
		}
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categories sql.NullString // Use sql.NullString to handle NULL values
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Username, &post.CreatedAt); err != nil {
			// Redirect based on login status
			if isLoggedIn {
				RenderError(w, r, "Error scanning posts", http.StatusInternalServerError, "/post")
			} else {
				RenderError(w, r, "Error scanning posts", http.StatusInternalServerError, "/")
			}
			return
		}
		if categories.Valid {
			post.Categories = categories.String // Assign the string value if valid
		} else {
			post.Categories = "" // Set to empty string if NULL
		}
		posts = append(posts, post)
	}

	// Render the home template with the filtered posts
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		// Redirect based on login status
		if isLoggedIn {
			RenderError(w, r, "Error parsing file", http.StatusInternalServerError, "/post")
		} else {
			RenderError(w, r, "Error parsing file", http.StatusInternalServerError, "/")
		}
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Posts":            posts,
		"IsLoggedIn":       isLoggedIn,
		"SelectedCategory": category,
	})
}
