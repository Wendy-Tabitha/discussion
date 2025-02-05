package handlers

// User represents a user in the system
type User struct {
	ID       string 
	Email    string
	Username string
	Password string
}

// Post represents a post in the forum
type Post struct {
	ID       int
	UserID   string 
	Title    string
	Content  string
	Categories string // Changed to string to hold multiple categories
	Username string 
	CreatedAt string
	LikeCount   int // Number of likes
	DislikeCount int
}

type Like struct {
	ID     int
	UserID string // User who liked/disliked the post
	PostID int    // Post that was liked/disliked
	IsLike bool   // true for like, false for dislike
}

// Comment represents a comment on a post
type Comment struct {
	ID      int
	PostID  int
	UserID  string 
	Content string
}

// Session represents a user session
type Session struct {
	SessionID string 
	UserID    string
}