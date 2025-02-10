package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
	"time"
)

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
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
			RenderError(w, r, "Database Error", http.StatusInternalServerError)
			return
		}
	}

	if r.Method != http.MethodGet {
		RenderError(w, r, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	categories := r.URL.Query()["category"]
	if len(categories) == 0 || (len(categories) == 1 && (categories[0] == "all" || categories[0] == "")) {
		// No category filter or "all" category - fetch all posts
		query := `
			SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, 
			u.username, p.created_at 
			FROM posts p 
			JOIN users u ON p.user_id = u.id 
			LEFT JOIN post_categories pc ON p.id = pc.post_id 
			GROUP BY p.id 
			ORDER BY p.created_at DESC`
		
		rows, err := db.Query(query)
		if err != nil {
			RenderError(w, r, "Error fetching posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			var categories sql.NullString
			if err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Username, &post.CreatedAt); err != nil {
				RenderError(w, r, "Error scanning posts", http.StatusInternalServerError)
				return
			}
			if categories.Valid {
				post.Categories = categories.String
			}
			posts = append(posts, post)
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			RenderError(w, r, "Error parsing file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"Posts":            posts,
			"IsLoggedIn":       userID != "",
			"SelectedCategory": "all",
		})
		return
	}

	// Filter by specific categories
	query := `
		SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, 
		u.username, p.created_at 
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		LEFT JOIN post_categories pc ON p.id = pc.post_id 
		WHERE pc.category IN (?` + strings.Repeat(",?", len(categories)-1) + `) 
		GROUP BY p.id 
		ORDER BY p.created_at DESC`

	args := make([]interface{}, len(categories))
	for i, category := range categories {
		args[i] = category
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categories sql.NullString
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Username, &post.CreatedAt); err != nil {
			RenderError(w, r, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		if categories.Valid {
			post.Categories = categories.String
		}
		posts = append(posts, post)
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		RenderError(w, r, "Error parsing file", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Posts":            posts,
		"IsLoggedIn":       userID != "",
		"SelectedCategory": strings.Join(categories, ","),
	})
}
