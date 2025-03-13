package models

import (
	"database/sql"
	"time"
	
	"codehubgo/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID              int       `json:"id"`
	Email           string    `json:"email"`
	Password        string    `json:"-"` // Don't expose password
	Name            string    `json:"name"`
	Role            string    `json:"role"`
	CreatedAt       time.Time `json:"created_at"`
	ShowEmail       bool      `json:"show_email"`
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*User, error) {
	row := database.ExecuteQueryRow("SELECT id, email, password, name, role, created_at, show_email FROM users WHERE id = ?", id)
	
	var user User
	var createdAtStr string
	
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Role, &createdAtStr, &user.ShowEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Parse created_at
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
	row := database.ExecuteQueryRow("SELECT id, email, password, name, role, created_at, show_email FROM users WHERE email = ?", email)
	
	var user User
	var createdAtStr string
	
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Role, &createdAtStr, &user.ShowEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Parse created_at
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	
	return &user, nil
}

// VerifyPassword checks if the provided password matches the stored hash
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// ToggleEmailVisibility toggles the show_email field for a user
func (u *User) ToggleEmailVisibility() error {
	u.ShowEmail = !u.ShowEmail
	
	_, err := database.ExecuteStatement(
		"UPDATE users SET show_email = ? WHERE id = ?",
		u.ShowEmail, u.ID,
	)
	
	return err
}

// GetUserFromSession creates a User object from session data
func GetUserFromSession(session *Session) (*User, error) {
	if session == nil || session.UserID == 0 {
		return nil, nil
	}
	
	return GetUserByID(session.UserID)
}
