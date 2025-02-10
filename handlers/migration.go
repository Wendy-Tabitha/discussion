package handlers

import (
	"log"
)

func RunMigrations() error {
	// Add parent_id column to comments table if it doesn't exist
	_, err := db.Exec(`
		SELECT parent_id FROM comments LIMIT 1;
	`)
	if err != nil {
		log.Println("Adding parent_id column to comments table...")
		_, err = db.Exec(`
			ALTER TABLE comments ADD COLUMN parent_id INTEGER DEFAULT NULL REFERENCES comments(id) ON DELETE CASCADE;
		`)
		if err != nil {
			return err
		}
	}

	// Create comment_likes table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comment_likes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comment_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			is_like BOOLEAN NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(comment_id, user_id)
		)
	`)
	if err != nil {
		log.Println("Error creating comment_likes table:", err)
		return err
	}

	return nil
}
