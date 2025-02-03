package handlers

type User struct {
	ID       int
	Email    string
	Username string
	Password string
}

type Post struct {
	ID       int
	UserID   int
	Title    string
	Content  string
	Category string
	Comments []Comment // Add this field to store comments for each post
}

type Comment struct {
	ID      int
	PostID  int
	UserID  int
	Content string
}
