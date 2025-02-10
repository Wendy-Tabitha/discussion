package main

import (
	"log"
	"net/http"

	"forum/handlers"
)

func main() {
	// Serve static files from the "static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Define routes
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/post", handlers.PostHandler)
	http.HandleFunc("/like", handlers.LikeHandler)
	http.HandleFunc("/filter", handlers.FilterHandler)
	http.HandleFunc("/comment", handlers.CommentHandler)
	http.HandleFunc("/comment/like", handlers.CommentLikeHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Initialize the database
	handlers.InitDB()

	// Run migrations
	if err := handlers.RunMigrations(); err != nil {
		log.Fatal("Error running migrations:", err)
	}

	// Start the server
	log.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
