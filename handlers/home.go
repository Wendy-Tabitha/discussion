package handlers

import (
	"html/template"
	"net/http"
	"database/sql"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(w, r)

	// Query to fetch all posts along with user info, categories, like counts, and comments
	rows, err := db.Query(`
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			GROUP_CONCAT(DISTINCT pc.category) as categories, 
			u.username, 
			p.created_at,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislike_count
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		LEFT JOIN post_categories pc ON p.id = pc.post_id 
		GROUP BY p.id 
		ORDER BY p.created_at DESC`)
	if err != nil {
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categories sql.NullString
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&categories,
			&post.Username,
			&post.CreatedAt,
			&post.LikeCount,
			&post.DislikeCount,
		); err != nil {
			RenderError(w, r, "Error scanning posts", http.StatusInternalServerError)
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
			RenderError(w, r, "Error fetching comments", http.StatusInternalServerError)
			return
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	// Render the index page with posts
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		RenderError(w, r, "Error parsing file", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": userID != "",
	})
}