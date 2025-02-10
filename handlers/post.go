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
			log.Printf("Error retrieving session: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/")
			return
		}
	}

	if r.Method == http.MethodPost {
		if userID == "" {
			RenderError(w, r, "Please log in to create posts", http.StatusUnauthorized, "/login")
			return
		}

		// Handle post creation
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["category"]

		// Validate inputs
		if title == "" || content == "" {
			RenderError(w, r, "Title and content cannot be empty", http.StatusBadRequest, "/post")
			return
		}

		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			RenderError(w, r, "Server Error", http.StatusInternalServerError, "/post")
			return
		}
		defer tx.Rollback()

		// Insert the new post into the database
		result, err := tx.Exec("INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)",
			userID, title, content, time.Now())
		if err != nil {
			log.Printf("Error inserting post: %v", err)
			RenderError(w, r, "Error creating post", http.StatusInternalServerError, "/post")
			return
		}

		postID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting last insert ID: %v", err)
			RenderError(w, r, "Error creating post", http.StatusInternalServerError, "/post")
			return
		}

		// Insert categories into the database
		for _, category := range categories {
			if category == "" {
				continue
			}
			_, err = tx.Exec("INSERT INTO post_categories (post_id, category) VALUES (?, ?)", postID, category)
			if err != nil {
				log.Printf("Error inserting category %s for post %d: %v", category, postID, err)
				RenderError(w, r, "Error saving categories", http.StatusInternalServerError, "/post")
				return
			}
		}

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			RenderError(w, r, "Error saving post", http.StatusInternalServerError, "/post")
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		// For GET requests, fetch and display posts
		query := `
			SELECT 
				p.id, 
				p.title, 
				p.content, 
				GROUP_CONCAT(DISTINCT pc.category) as categories, 
				u.username, 
				p.created_at,
				COUNT(DISTINCT CASE WHEN l.is_like = 1 THEN l.id END) as like_count,
				COUNT(DISTINCT CASE WHEN l.is_like = 0 THEN l.id END) as dislike_count,
				COUNT(DISTINCT c.id) as comment_count
			FROM posts p 
			JOIN users u ON p.user_id = u.id 
			LEFT JOIN post_categories pc ON p.id = pc.post_id 
			LEFT JOIN likes l ON p.id = l.post_id
			LEFT JOIN comments c ON p.id = c.post_id
			GROUP BY p.id 
			ORDER BY p.created_at DESC`

		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Error fetching posts: %v", err)
			RenderError(w, r, "Error loading posts", http.StatusInternalServerError, "/")
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			var categories sql.NullString
			var commentCount sql.NullInt64
			err := rows.Scan(
				&post.ID,
				&post.Title,
				&post.Content,
				&categories,
				&post.Username,
				&post.CreatedAt,
				&post.LikeCount,
				&post.DislikeCount,
				&commentCount,
			)
			if err != nil {
				log.Printf("Error scanning post row: %v", err)
				RenderError(w, r, "Error processing posts", http.StatusInternalServerError, "/")
				return
			}

			if categories.Valid {
				post.Categories = categories.String
			}
			if commentCount.Valid {
				post.CommentCount = commentCount.String
			}

			posts = append(posts, post)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Error iterating posts: %v", err)
			RenderError(w, r, "Error processing posts", http.StatusInternalServerError, "/")
			return
		}

		// Get available categories for the post form
		var availableCategories []string
		categoryRows, err := db.Query("SELECT DISTINCT category FROM post_categories ORDER BY category")
		if err != nil {
			log.Printf("Error fetching categories: %v", err)
			RenderError(w, r, "Error loading categories", http.StatusInternalServerError, "/")
			return
		}
		defer categoryRows.Close()

		for categoryRows.Next() {
			var category string
			if err := categoryRows.Scan(&category); err != nil {
				log.Printf("Error scanning category: %v", err)
				RenderError(w, r, "Error processing categories", http.StatusInternalServerError, "/")
				return
			}
			availableCategories = append(availableCategories, category)
		}

		// Render the template
		tmpl, err := template.ParseFiles("templates/post.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			RenderError(w, r, "Error loading page", http.StatusInternalServerError, "/")
			return
		}

		data := map[string]interface{}{
			"Posts":       posts,
			"IsLoggedIn":  userID != "",
			"Categories":  availableCategories,
			"CurrentUser": userID,
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing template: %v", err)
			RenderError(w, r, "Error rendering page", http.StatusInternalServerError, "/")
			return
		}
		return
	}

	RenderError(w, r, "Method Not Allowed", http.StatusMethodNotAllowed, "/")
}