package handlers

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        username TEXT,
        password TEXT
    );

    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        title TEXT,
        content TEXT,
        category TEXT,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER,
        user_id INTEGER,
        content TEXT,
        FOREIGN KEY(post_id) REFERENCES posts(id),
        FOREIGN KEY(user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        post_id INTEGER,
        is_like BOOLEAN,
        FOREIGN KEY(user_id) REFERENCES users(id),
        FOREIGN KEY(post_id) REFERENCES posts(id)
    );

     CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        session_id TEXT NOT NULL,
        user_id INTEGER, 
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
    `
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}


// Insert a like into the database
func insertLike(userID, postID int, isLike bool) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	query := `INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)`
	_, err = db.Exec(query, userID, postID, isLike)
	if err != nil {
		return err
	}
	return nil
}

func insertSession(sessionID string, userID int) error {
	query := `INSERT INTO sessions (session_id, user_id) VALUES (?,?)`

	_, err := db.Exec(query, sessionID, userID)
	if err != nil {
		return err
	}
	return nil
}


