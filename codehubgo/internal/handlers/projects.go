package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"log"
	
	"github.com/gorilla/mux"
	"codehubgo/internal/middleware"
	"codehubgo/internal/models"
	"codehubgo/internal/utils"
)

// ProjectsHomeHandler handles the projects home page
func ProjectsHomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	
	// Get user projects
	userProjects, err := models.GetProjectsByOwner(session.UserID)
	if err != nil {
		log.Printf("Error fetching user projects: %v", err)
		http.Error(w, "Error fetching projects", http.StatusInternalServerError)
		return
	}
	
	// Get public projects
	publicProjects, err := models.GetPublicProjects()
	if err != nil {
		log.Printf("Error fetching public projects: %v", err)
		http.Error(w, "Error fetching projects", http.StatusInternalServerError)
		return
	}
	
	// Check if each project is starred by the current user
	for _, project := range userProjects {
		starred, err := project.IsStarredByUser(session.UserID)
		if err != nil {
			log.Printf("Error checking if project is starred: %v", err)
		} else {
			project.Starred = starred
		}
	}
	
	// Check if each public project is starred by the current user
	for _, project := range publicProjects {
		starred, err := project.IsStarredByUser(session.UserID)
		if err != nil {
			log.Printf("Error checking if public project is starred: %v", err)
		} else {
			project.Starred = starred
		}
	}
	
	// Count statistics
	var totalProjects = len(userProjects)
	var publicCount = 0
	var privateCount = 0
	var starredCount = 0
	
	for _, project := range userProjects {
		if project.Public {
			publicCount++
		} else {
			privateCount++
		}
		if project.Starred {
			starredCount++
		}
	}
	
	// Render projects template
	utils.RenderTemplate(w, "projects.html", map[string]interface{}{
		"Title":           "Projects",
		"ActivePage":      "projects",
		"User":            session,
		"Projects":        userProjects,
		"PublicProjects":  publicProjects,
		"TotalProjects":   totalProjects,
		"PublicCount":     publicCount,
		"PrivateProjects": privateCount,
		"StarredProjects": starredCount,
	})
}

// ProjectHandler handles viewing and editing a project
func ProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	
	// Get project ID from URL
	vars := mux.Vars(r)
	projID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	// Get project
	project, err := models.GetProjectByID(projID)
	if err != nil {
		http.Error(w, "Error fetching project", http.StatusInternalServerError)
		return
	}
	
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	
	// Check if user is the owner
	isOwner := project.Owner == session.UserID
	
	// Check if project is published or if user is the owner
	if !project.Enabled && !isOwner {
		http.Error(w, "You don't have permission to view this project", http.StatusForbidden)
		return
	}
	
	// Get owner information
	owner, err := models.GetUserByID(project.Owner)
	if err != nil {
		http.Error(w, "Error fetching owner information", http.StatusInternalServerError)
		return
	}
	
	// Get environment variables
	envVariables, err := project.GetEnvironmentVariables()
	if err != nil {
		http.Error(w, "Error parsing environment variables", http.StatusInternalServerError)
		return
	}
	
	// Check if user has starred this project
	isStarred, err := project.IsStarredByUser(session.UserID)
	if err != nil {
		http.Error(w, "Error checking star status", http.StatusInternalServerError)
		return
	}
	
	// Get star count
	starCount, err := project.GetStarCount()
	if err != nil {
		http.Error(w, "Error fetching star count", http.StatusInternalServerError)
		return
	}
	
	// Convert markdown description to HTML
	projectDescriptionHTML := utils.MarkdownToHTML(project.Description)
	
	// Handle form submission for editing
	if r.Method == http.MethodPost && isOwner {
		name := r.FormValue("name")
		githubLink := r.FormValue("github_link")
		description := r.FormValue("description")
		
		// Update environment variables
		updatedEnv := make(map[string]string)
		for key, _ := range r.Form {
			if len(key) > 9 && key[:9] == "variable_" {
				varName := key[9:]
				updatedEnv[varName] = r.FormValue(key)
			}
		}
		
		// Update project
		project.Name = name
		project.GithubLink = githubLink
		project.Description = description
		
		// Convert environment to JSON
		envJSON, err := json.Marshal(updatedEnv)
		if err != nil {
			http.Error(w, "Error updating project", http.StatusInternalServerError)
			return
		}
		project.Environment = string(envJSON)
		
		// Save project
		err = project.Update()
		if err != nil {
			http.Error(w, "Error updating project", http.StatusInternalServerError)
			return
		}
		
		// Redirect to refresh page
		http.Redirect(w, r, "/project/"+vars["id"], http.StatusSeeOther)
		return
	}
	
	// Render project template
	utils.RenderTemplate(w, "project.html", map[string]interface{}{
		"Title":                 "Project: " + project.Name,
		"ActivePage":            "projects",
		"User":                  session,
		"Project":               project,
		"IsOwner":               isOwner,
		"EnvVariables":          envVariables,
		"ProjectDescriptionHTML": projectDescriptionHTML,
		"IsStarred":             isStarred,
		"StarCount":             starCount,
		"OwnerInfo":             owner,
	})
}

// CreateProjectHandler handles creating a new project
func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	
	// Get available eggs
	eggs, err := utils.GetAvailableEggs()
	if err != nil {
		http.Error(w, "Error fetching available eggs", http.StatusInternalServerError)
		return
	}
	
	// Handle form submission
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		githubLink := r.FormValue("github_link")
		eggID, _ := strconv.Atoi(r.FormValue("egg"))
		
		// Get environment variables
		environment := make(map[string]string)
		for key, _ := range r.Form {
			if len(key) > 9 && key[:9] == "variable_" {
				varName := key[9:]
				environment[varName] = r.FormValue(key)
			}
		}
		
		// Validate input
		if name == "" || githubLink == "" {
			utils.RenderTemplate(w, "create_project.html", map[string]interface{}{
				"Username":      session.Email,
				"UserID":        session.UserID,
				"Error":         "Both fields are required!",
				"AvailableEggs": eggs,
			})
			return
		}
		
		// Create project
		err := models.CreateProject(name, githubLink, session.UserID, eggID, environment)
		if err != nil {
			http.Error(w, "Error creating project", http.StatusInternalServerError)
			return
		}
		
		// Redirect to projects page
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
		return
	}
	
	// Render create project template
	utils.RenderTemplate(w, "create_project.html", map[string]interface{}{
		"Username":      session.Email,
		"UserID":        session.UserID,
		"AvailableEggs": eggs,
	})
}

// ToggleStarHandler handles toggling a project star
func ToggleStarHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get project ID from URL
	vars := mux.Vars(r)
	projID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	// Get project
	project, err := models.GetProjectByID(projID)
	if err != nil {
		http.Error(w, "Error fetching project", http.StatusInternalServerError)
		return
	}
	
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	
	// Toggle star
	isStarred, err := project.ToggleStar(session.UserID)
	if err != nil {
		http.Error(w, "Error toggling star", http.StatusInternalServerError)
		return
	}
	
	// Get updated star count
	starCount, err := project.GetStarCount()
	if err != nil {
		http.Error(w, "Error fetching star count", http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_starred": isStarred,
		"star_count": starCount,
	})
}

// DeleteProjectHandler handles deleting a project
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get project ID from URL
	vars := mux.Vars(r)
	projID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	// Get project
	project, err := models.GetProjectByID(projID)
	if err != nil {
		http.Error(w, "Error fetching project", http.StatusInternalServerError)
		return
	}
	
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	
	// Check if user is the owner
	if project.Owner != session.UserID {
		http.Error(w, "You don't have permission to delete this project", http.StatusForbidden)
		return
	}
	
	// Delete project
	err = project.Delete()
	if err != nil {
		http.Error(w, "Error deleting project", http.StatusInternalServerError)
		return
	}
	
	// Redirect to projects page
	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

// TogglePublishHandler handles toggling a project's publish status
func TogglePublishHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get project ID from URL
	vars := mux.Vars(r)
	projID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	// Get project
	project, err := models.GetProjectByID(projID)
	if err != nil {
		http.Error(w, "Error fetching project", http.StatusInternalServerError)
		return
	}
	
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	
	// Check if user is the owner
	if project.Owner != session.UserID {
		http.Error(w, "You don't have permission to publish this project", http.StatusForbidden)
		return
	}
	
	// Toggle publish status
	err = project.TogglePublish()
	if err != nil {
		http.Error(w, "Error toggling publish status", http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": project.Enabled,
	})
}

// UserProfileHandler handles viewing a user's profile
func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	
	// Get user ID from URL
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Get user
	user, err := models.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}
	
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Get user's projects
	projects, err := models.GetProjectsByOwner(userID)
	if err != nil {
		http.Error(w, "Error fetching projects", http.StatusInternalServerError)
		return
	}
	
	// Check if viewing own profile
	isOwnProfile := session.UserID == userID
	
	// Render user profile template
	utils.RenderTemplate(w, "user_profile.html", map[string]interface{}{
		"Username":     session.Email,
		"UserID":       session.UserID,
		"ProfileUser":  user,
		"Projects":     projects,
		"IsOwnProfile": isOwnProfile,
	})
}

// ToggleEmailVisibilityHandler handles toggling a user's email visibility
func ToggleEmailVisibilityHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get user
	user, err := models.GetUserByID(session.UserID)
	if err != nil {
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}
	
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Toggle email visibility
	err = user.ToggleEmailVisibility()
	if err != nil {
		http.Error(w, "Error toggling email visibility", http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"show_email": user.ShowEmail,
	})
}

// SearchHandler handles searching for projects
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	
	// Get search query
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
		return
	}
	
	// Get all projects
	allProjects, err := models.GetPublicProjects()
	if err != nil {
		http.Error(w, "Error fetching projects", http.StatusInternalServerError)
		return
	}
	
	// Search projects
	searchResults := utils.SearchProjects(allProjects, query)
	
	// Render search results template
	utils.RenderTemplate(w, "search_results.html", map[string]interface{}{
		"Username":      session.Email,
		"UserID":        session.UserID,
		"Query":         query,
		"SearchResults": searchResults,
	})
}

// GetEggVariablesHandler handles fetching egg variables
func GetEggVariablesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user session
	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get selected egg ID
	eggID := r.FormValue("egg")
	if eggID == "" {
		http.Error(w, "No egg selected", http.StatusBadRequest)
		return
	}
	
	// Fetch egg variables
	eggVariables, err := utils.GetEggVariables(eggID)
	if err != nil {
		http.Error(w, "Error fetching egg variables", http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"egg_variables": eggVariables,
	})
}
