package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	isLoggedIn := false

	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == nil {
			isLoggedIn = true
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
			log.Printf("Database error in FilterHandler: %v", err)
			RenderError(w, r, "Database Error", http.StatusInternalServerError, "/")
			return
		}
	}

	if r.Method != http.MethodGet {
		RenderError(w, r, "Method Not Allowed", http.StatusMethodNotAllowed, "/")
		return
	}

	categories := r.URL.Query()["category"]
	var posts []Post
	var query string
	var args []interface{}

	if len(categories) == 0 || (len(categories) == 1 && (categories[0] == "all" || categories[0] == "")) {
		// No category filter or "all" category - fetch all posts
		query = `
			SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, 
			u.username, p.created_at, COUNT(DISTINCT c.id) as comment_count,
			COUNT(DISTINCT l.id) as like_count
			FROM posts p 
			JOIN users u ON p.user_id = u.id 
			LEFT JOIN post_categories pc ON p.id = pc.post_id 
			LEFT JOIN comments c ON p.id = c.post_id
			LEFT JOIN likes l ON p.id = l.post_id
			GROUP BY p.id 
			ORDER BY p.created_at DESC`
	} else {
		// Filter by specific categories
		query = `
			SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, 
			u.username, p.created_at, COUNT(DISTINCT c.id) as comment_count,
			COUNT(DISTINCT l.id) as like_count
			FROM posts p 
			JOIN users u ON p.user_id = u.id 
			LEFT JOIN post_categories pc ON p.id = pc.post_id 
			LEFT JOIN comments c ON p.id = c.post_id
			LEFT JOIN likes l ON p.id = l.post_id
			WHERE pc.category IN (?` + strings.Repeat(",?", len(categories)-1) + `) 
			GROUP BY p.id 
			ORDER BY p.created_at DESC`

		args = make([]interface{}, len(categories))
		for i, category := range categories {
			args[i] = category
		}
	}

	var rows *sql.Rows
	if len(args) > 0 {
		rows, err = db.Query(query, args...)
	} else {
		rows, err = db.Query(query)
	}

	if err != nil {
		log.Printf("Error querying posts: %v", err)
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError, "/")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var categories, commentCount, likeCount sql.NullString
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&categories,
			&post.Username,
			&post.CreatedAt,
			&commentCount,
			&likeCount,
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
		if likeCount.Valid {
			post.LikeCount = likeCount.String
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating posts: %v", err)
		RenderError(w, r, "Error processing posts", http.StatusInternalServerError, "/")
		return
	}

	// Get all available categories for the filter dropdown
	var availableCategories []string
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
		availableCategories = append(availableCategories, category)
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		RenderError(w, r, "Error loading page", http.StatusInternalServerError, "/")
		return
	}

	selectedCategory := "all"
	if len(categories) == 1 {
		selectedCategory = categories[0]
	}

	data := map[string]interface{}{
		"Posts":              posts,
		"IsLoggedIn":         isLoggedIn,
		"Categories":         availableCategories,
		"SelectedCategory":   selectedCategory,
		"CurrentUser":        userID,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		RenderError(w, r, "Error rendering page", http.StatusInternalServerError, "/")
		return
	}
}
