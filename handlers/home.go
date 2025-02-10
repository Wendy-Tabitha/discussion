package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(w, r)

	// Query to fetch all posts along with user info, categories, like counts, and comments
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			GROUP_CONCAT(DISTINCT pc.category) as categories, 
			u.username, 
			p.created_at,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislike_count,
			COUNT(DISTINCT c.id) as comment_count
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		LEFT JOIN post_categories pc ON p.id = pc.post_id 
		LEFT JOIN comments c ON p.id = c.post_id
		GROUP BY p.id 
		ORDER BY p.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/")
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

		// Fetch comments for this post
		comments, err := GetCommentsForPost(post.ID)
		if err != nil {
			log.Printf("Error fetching comments for post %d: %v", post.ID, err)
			RenderError(w, r, "Error fetching comments", http.StatusInternalServerError, "/")
			return
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating posts: %v", err)
		RenderError(w, r, "Error processing posts", http.StatusInternalServerError, "/")
		return
	}

	// Get categories for filter dropdown
	var categories []string
	categoryRows, err := db.Query("SELECT DISTINCT category FROM post_categories ORDER BY category")
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		RenderError(w, r, "Error fetching categories", http.StatusInternalServerError, "/")
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
		categories = append(categories, category)
	}

	// Render the home page
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		RenderError(w, r, "Error loading page", http.StatusInternalServerError, "/")
		return
	}

	data := map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": userID != "",
		"Categories": categories,
		"CurrentUser": userID,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		RenderError(w, r, "Error rendering page", http.StatusInternalServerError, "/")
		return
	}
}