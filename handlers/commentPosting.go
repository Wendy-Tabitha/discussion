package handlers

import (
    "net/http"
    "strconv"
)

func CommentHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        // Get the post ID from the URL (assuming you use /comment/{postID} format)
        postIDStr := r.URL.Path[len("/comment/"):]
        postID, err := strconv.Atoi(postIDStr)
        if err != nil {
            http.Error(w, "Invalid post ID", http.StatusBadRequest)
            return
        }

        content := r.FormValue("content")
        userID := 1 // Replace with actual user ID from session

        // Insert the new comment into the database
        _, err = db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
        if err != nil {
            http.Error(w, "Error posting comment", http.StatusInternalServerError)
            return
        }

        // Redirect back to the post page
        http.Redirect(w, r, "/post/"+postIDStr, http.StatusFound)
    }
}
