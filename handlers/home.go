package handlers

import (
	"html/template"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Query to fetch all posts along with the user's name and creation time, ordered by created_at descending
	rows, err := db.Query(`SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, u.username, 
	p.created_at FROM posts p JOIN users u ON p.user_id = u.id LEFT JOIN post_categories pc ON p.id = pc.post_id GROUP BY p.id ORDER BY p.created_at DESC`)
	if err != nil {
		RenderError(w,r, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Categories, &post.Username, &post.CreatedAt); err != nil {
			RenderError(w,r, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}
	// Render the index page with posts
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		RenderError(w,r, "Error parsing file", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Posts": posts,
	})
}
