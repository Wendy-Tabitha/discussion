package handlers

import (
	"database/sql"
	"html/template"
	"log"
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
			log.Println("Error retrieving session:", err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodPost {
		// Handle post creation
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["category"]

		// Validate inputs
		if title == "" || content == "" {
			RenderError(w, r, "Title and content cannot be empty", http.StatusBadRequest)
			return
		}

		// Insert the new post into the database
		result, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
		if err != nil {
			log.Println("Error inserting post:", err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}

		postID, err := result.LastInsertId()
		if err != nil {
			log.Println("Error getting last insert ID:", err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}

		// Insert categories into the database
		for _, category := range categories {
			_, err = db.Exec("INSERT INTO post_categories (post_id, category) VALUES (?, ?)", postID, category)
			if err != nil {
				log.Printf("Error inserting category %s for post %d: %v", category, postID, err)
				// Continue with other categories even if one fails
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// For GET requests, fetch and display posts
	rows, err := db.Query(`
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			GROUP_CONCAT(pc.category) as categories, 
			u.username, 
			p.created_at,
			COALESCE(SUM(CASE WHEN l.is_like = 1 THEN 1 ELSE 0 END), 0) AS like_count,
			COALESCE(SUM(CASE WHEN l.is_like = 0 THEN 1 ELSE 0 END), 0) AS dislike_count
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		LEFT JOIN post_categories pc ON p.id = pc.post_id 
		LEFT JOIN likes l ON p.id = l.post_id 
		GROUP BY p.id, p.title, p.content, u.username, p.created_at 
		ORDER BY p.created_at DESC`)
	if err != nil {
		log.Println("Error fetching posts:", err)
		RenderError(w, r, "database_error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categories sql.NullString
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Username, &post.CreatedAt, &post.LikeCount, &post.DislikeCount); err != nil {
			log.Println("Error scanning posts:", err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}
		if categories.Valid {
			post.Categories = categories.String
		} else {
			post.Categories = ""
		}

		// Fetch comments for this post
		comments, err := GetCommentsForPost(post.ID)
		if err != nil {
			log.Printf("Error fetching comments for post %d: %v", post.ID, err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		RenderError(w, r, "template_error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": userID != "",
	})
	if err != nil {
		log.Println("Error executing template:", err)
		RenderError(w, r, "template_error", http.StatusInternalServerError)
	}
}