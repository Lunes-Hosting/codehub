package handlers

import (
	"net/http"
	"net/url"
	
	"golang.org/x/crypto/bcrypt"
	
	"codehubgo/internal/config"
	"codehubgo/internal/middleware"
	"codehubgo/internal/models"
	"codehubgo/internal/utils"
)

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := middleware.Store.Get(r, "session")
	
	// Check if user is already logged in
	if session.Values["user_id"] != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	// Check for flash messages
	flashCookie, err := r.Cookie("flash")
	var flashMessage string
	if err == nil {
		flashMessage, _ = url.QueryUnescape(flashCookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:   "flash",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}
	
	// Handle form submission
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		
		user, err := models.GetUserByEmail(email)
		if err != nil {
			utils.RenderTemplate(w, "login.html", map[string]interface{}{
				"Error": "An error occurred during login. Please try again.",
			})
			return
		}
		
		if user != nil && user.VerifyPassword(password) {
			// Set session values
			session.Values["user_id"] = user.ID
			session.Values["email"] = user.Email
			session.Values["role"] = user.Role
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			// Redirect to the stored next page if it exists and is safe
			nextPage, ok := session.Values["next"].(string)
			if ok && middleware.IsSafeURL(nextPage, r) {
				delete(session.Values, "next")
				session.Save(r, w)
				http.Redirect(w, r, nextPage, http.StatusSeeOther)
				return
			}
			
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		
		utils.RenderTemplate(w, "login.html", map[string]interface{}{
			"Error": "Invalid username or password",
		})
		return
	}
	
	// Render login form
	utils.RenderTemplate(w, "login.html", map[string]interface{}{
		"FlashMessage": flashMessage,
	})
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	
	// Clear session
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	session.Save(r, w)
	
	// Set flash message
	http.SetCookie(w, &http.Cookie{
		Name:  "flash",
		Value: url.QueryEscape("You have been logged out"),
		Path:  "/",
	})
	
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// RegisterHandler redirects to the dashboard registration page
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.DashboardURL+"/register", http.StatusSeeOther)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
