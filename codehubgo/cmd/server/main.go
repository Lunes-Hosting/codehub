package main

import (
	"log"
	"net/http"
	"path/filepath"
	
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	
	"codehubgo/internal/config"
	"codehubgo/internal/database"
	"codehubgo/internal/handlers"
	"codehubgo/internal/middleware"
	"codehubgo/internal/utils"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}
	
	// Initialize database connection
	_, err = database.InitDB()
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Continuing without database connection for UI testing purposes.")
		// Not fatal, continue with application startup for UI testing
	} else {
		defer database.CloseDB()
		log.Println("Database connection successful!")
	}
	
	// Load templates
	templatesDir := filepath.Join("web", "templates")
	err = utils.LoadTemplates(templatesDir)
	if err != nil {
		log.Fatalf("Error loading templates: %v", err)
	}
	
	// Create router
	r := mux.NewRouter()
	
	// Static files
	staticDir := filepath.Join("web", "static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	
	// Auth routes
	r.HandleFunc("/auth/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/auth/logout", handlers.LogoutHandler).Methods("GET")
	r.HandleFunc("/auth/register", handlers.RegisterHandler).Methods("GET")
	
	// Dashboard routes
	r.HandleFunc("/", middleware.AuthRequired(handlers.DashboardHandler)).Methods("GET")
	
	// Project routes
	r.HandleFunc("/projects", middleware.AuthRequired(handlers.ProjectsHomeHandler)).Methods("GET")
	r.HandleFunc("/project/{id:[0-9]+}", middleware.AuthRequired(handlers.ProjectHandler)).Methods("GET", "POST")
	r.HandleFunc("/project/create", middleware.AuthRequired(handlers.CreateProjectHandler)).Methods("GET", "POST")
	r.HandleFunc("/project/{id:[0-9]+}/delete", middleware.AuthRequired(handlers.DeleteProjectHandler)).Methods("POST")
	r.HandleFunc("/project/{id:[0-9]+}/toggle-publish", middleware.AuthRequired(handlers.TogglePublishHandler)).Methods("POST")
	r.HandleFunc("/project/{id:[0-9]+}/toggle-star", middleware.AuthRequired(handlers.ToggleStarHandler)).Methods("POST")
	
	// User routes
	r.HandleFunc("/user/{id:[0-9]+}", middleware.AuthRequired(handlers.UserProfileHandler)).Methods("GET")
	r.HandleFunc("/user/toggle-email-visibility", middleware.AuthRequired(handlers.ToggleEmailVisibilityHandler)).Methods("POST")
	
	// Search routes
	r.HandleFunc("/search", middleware.AuthRequired(handlers.SearchHandler)).Methods("GET")
	
	// API routes
	r.HandleFunc("/api/egg-variables", middleware.AuthRequired(handlers.GetEggVariablesHandler)).Methods("POST")
	
	// Start server
	log.Printf("Server starting on port %s...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, r))
}
