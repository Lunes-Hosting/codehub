package middleware

import (
	"context"
	"net/http"
	"net/url"
	
	"github.com/gorilla/sessions"
	
	"codehubgo/internal/config"
	"codehubgo/internal/models"
)

// Key for storing session in context
type sessionKey struct{}

// Store is the session store
var Store = sessions.NewCookieStore([]byte(config.SecretKey))

// Initialize session store
func init() {
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
}

// AuthRequired is middleware that checks if the user is authenticated
func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session")
		
		// Check if user is authenticated
		if session.Values["user_id"] == nil {
			// Store the full URL they were trying to access
			nextPage := r.URL.String()
			if IsSafeURL(nextPage, r) {
				session.Values["next"] = nextPage
				session.Save(r, w)
			}
			
			http.SetCookie(w, &http.Cookie{
				Name:  "flash",
				Value: url.QueryEscape("Please log in to access this page"),
				Path:  "/",
			})
			
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		
		// Create session object and add to context
		userSession := &models.Session{
			UserID: session.Values["user_id"].(int),
			Email:  session.Values["email"].(string),
			Role:   session.Values["role"].(string),
		}
		
		ctx := context.WithValue(r.Context(), sessionKey{}, userSession)
		next(w, r.WithContext(ctx))
	}
}

// GetSession gets the user session from the request context
func GetSession(r *http.Request) *models.Session {
	session, ok := r.Context().Value(sessionKey{}).(*models.Session)
	if !ok {
		return nil
	}
	return session
}

// IsSafeURL checks if the URL is safe to redirect to
func IsSafeURL(target string, r *http.Request) bool {
	refURL, err := url.Parse(r.Host)
	if err != nil {
		return false
	}
	
	testURL, err := url.Parse(target)
	if err != nil {
		return false
	}
	
	// Only allow http and https URLs
	if testURL.Scheme != "" && testURL.Scheme != "http" && testURL.Scheme != "https" {
		return false
	}
	
	// Only allow URLs on the same host
	if testURL.Host != "" && testURL.Host != refURL.Host {
		return false
	}
	
	return true
}
