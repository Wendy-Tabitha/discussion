package handlers

import (
    "html/template"
    "net/http"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        // Handle new post creation
        title := r.FormValue("title")
        content := r.FormValue("content")
        category := r.FormValue("category")

        // Get user ID from session (you should implement session management)
        userID := 1 // Replace with actual user ID from session

        // Insert the new post into the database
        _, err := db.Exec("INSERT INTO posts (user_id, title, content, category) VALUES (?, ?, ?, ?)", userID, title, content, category)
        if err != nil {
            http.Error(w, "Error creating post", http.StatusInternalServerError)
            return
        }
    }

    // Query to fetch all posts
    rows, err := db.Query("SELECT id, user_id, title, content, category FROM posts")
    if err != nil {
        http.Error(w, "Error fetching posts", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var posts []Post
    for rows.Next() {
        var post Post
        if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Category); err != nil {
            http.Error(w, "Error scanning posts", http.StatusInternalServerError)
            return
        }

        // Fetch comments for the current post
        commentRows, err := db.Query("SELECT id, content, user_id FROM comments WHERE post_id = ?", post.ID)
        if err != nil {
            http.Error(w, "Error fetching comments", http.StatusInternalServerError)
            return
        }

        var comments []Comment
        for commentRows.Next() {
            var comment Comment
            if err := commentRows.Scan(&comment.ID, &comment.Content, &comment.UserID); err != nil {
                http.Error(w, "Error scanning comments", http.StatusInternalServerError)
                return
            }
            comments = append(comments, comment)
        }

        // Add comments to the post
        post.Comments = comments

        posts = append(posts, post)
    }

    // Render the post page with posts and comments
    tmpl, err := template.ParseFiles("templates/post.html")
    if err != nil {
        http.Error(w, "Error parsing file", http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, map[string]interface{}{
        "Posts": posts,
    })
}
