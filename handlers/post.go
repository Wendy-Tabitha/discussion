package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
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
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["category"] // Get multiple categories

		// Insert the new post into the database
		result, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
		if err != nil {
			RenderError(w, r, "Error creating post", http.StatusInternalServerError, "/post")
			return
		}

		postID, err := result.LastInsertId()
		if err != nil {
			RenderError(w, r, "Error retrieving post ID", http.StatusInternalServerError, "/post")
			return
		}

		// Insert categories into the database
		for _, category := range categories {
			_, err = db.Exec("INSERT INTO post_categories (post_id, category) VALUES (?, ?)", postID, category)
			if err != nil {
				RenderError(w, r, "Error inserting categories", http.StatusInternalServerError, "/post")
				return
			}
		}

		http.Redirect(w, r, "/post", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, u.username, 
	p.created_at, COALESCE(SUM(CASE WHEN l.is_like = 1 THEN 1 ELSE 0 END), 0) AS like_count,
	COALESCE(SUM(CASE WHEN l.is_like = 0 THEN 1 ELSE 0 END), 0) AS dislike_count
    FROM posts p JOIN users u ON p.user_id = u.id LEFT JOIN post_categories pc ON p.id = pc.post_id  LEFT JOIN 
	likes l ON p.id = l.post_id GROUP BY p.id, p.title, p.content, u.username, p.created_at ORDER BY p.created_at DESC`)
	if err != nil {
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/post")
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categories sql.NullString // Use sql.NullString to handle NULL values
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Username, &post.CreatedAt, &post.LikeCount, &post.DislikeCount,); err != nil {
			RenderError(w, r, "Error scanning posts", http.StatusInternalServerError, "/post")
			return
		}
		if categories.Valid {
			post.Categories = categories.String // Assign the string value if valid
		} else {
			post.Categories = "" // Set to empty string if NULL
		}
		posts = append(posts, post)
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		RenderError(w, r, "Error parsing file", http.StatusInternalServerError, "/post")
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": userID != "",
	})
}