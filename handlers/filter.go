package handlers

import (
	"html/template"
	"net/http"
	"strings"
)

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		categories := r.URL.Query()["category"]

		// Create a query to filter posts by multiple categories
		query := "SELECT p.id, p.title, p.content, GROUP_CONCAT(pc.category) as categories, u.username FROM posts p JOIN users u ON p.user_id = u.id LEFT JOIN post_categories pc ON p.id = pc.post_id WHERE pc.category IN (?" + strings.Repeat(",?", len(categories)-1) + ") GROUP BY p.id"
		args := make([]interface{}, len(categories))
		for i, category := range categories {
			args[i] = category
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			RenderError(w, "Error fetching posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Categories, &post.Username); err != nil {
				RenderError(w, "Error scanning posts", http.StatusInternalServerError)
				return
			}
			posts = append(posts, post)
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			RenderError(w, "Error parsing file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"Posts": posts,
		})
		return
	}

	RenderError(w, "Invalid request method", http.StatusMethodNotAllowed)
}
