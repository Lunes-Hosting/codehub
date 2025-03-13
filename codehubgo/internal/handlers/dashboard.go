package handlers

import (
	"net/http"
	
	"codehubgo/internal/middleware"
	"codehubgo/internal/utils"
)

// DashboardHandler handles the dashboard page
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	
	// Render dashboard template
	utils.RenderTemplate(w, "index.html", map[string]interface{}{
		"Username": session.Email,
		"UserID":   session.UserID,
	})
}
