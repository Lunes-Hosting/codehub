package models

// Session represents a user session
type Session struct {
	UserID int
	Email  string
	Role   string
}
