package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"codehubgo/internal/database"
)

// Project represents a project in the system
type Project struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`    // Used in template (was Name)
	Name        string    `json:"name"`     // Keep for backward compatibility
	Description string    `json:"description"`
	GithubLink  string    `json:"github_link"`
	CreateTime  time.Time `json:"create_time"`
	CreatedAt   string    `json:"created_at"` // Formatted date for display
	Owner       int       `json:"owner"`
	OwnerName   string    `json:"owner_name"`
	Environment string    `json:"environment"`
	EggID       int       `json:"egg_id"`
	Enabled     bool      `json:"enabled"`
	Public      bool      `json:"public"`      // Same as Enabled but named for template
	Starred     bool      `json:"starred"`     // Whether current user has starred it
	Stars       int       `json:"stars"`       // Star count
	UpdatedAt   string    `json:"updated_at"` // Formatted date for display
}

// GetProjectByID retrieves a project by ID
func GetProjectByID(id int) (*Project, error) {
	row := database.ExecuteQueryRow(`
		SELECT p.id, p.name, p.description, p.github_link, p.create_time, 
		       p.owner, u.name as owner_name, p.environment, p.egg_id, p.enabled
		FROM projects p
		LEFT JOIN users u ON p.owner = u.id
		WHERE p.id = ?
	`, id)

	var project Project
	var createTimeStr string
	var enabled int

	err := row.Scan(
		&project.ID, &project.Name, &project.Description, &project.GithubLink,
		&createTimeStr, &project.Owner, &project.OwnerName, &project.Environment,
		&project.EggID, &enabled,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse create time - try multiple formats
	createTime, err := parseTime(createTimeStr)
	if err != nil {
		log.Printf("Error parsing time '%s': %v", createTimeStr, err)
		// Use current time as fallback
		createTime = time.Now()
	}

	project.CreateTime = createTime
	project.CreatedAt = createTime.Format("Jan 2, 2006")

	// Set enabled flag
	project.Enabled = enabled == 1
	project.Public = project.Enabled

	// Set title to be the same as name for template compatibility
	project.Title = project.Name

	// Get star count
	project.Stars, _ = project.GetStarCount()

	return &project, nil
}

// GetProjectsByOwner retrieves all projects owned by a user
func GetProjectsByOwner(ownerID int) ([]*Project, error) {
	rows, err := database.ExecuteQuery(`
		SELECT p.id, p.name, p.description, p.github_link, p.create_time, 
		       p.owner, u.name as owner_name, p.environment, p.egg_id, p.enabled
		FROM projects p
		LEFT JOIN users u ON p.owner = u.id
		WHERE p.owner = ?
		ORDER BY p.create_time DESC
	`, ownerID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project

	for rows.Next() {
		var project Project
		var createTimeStr string
		var enabled int

		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.GithubLink,
			&createTimeStr, &project.Owner, &project.OwnerName, &project.Environment,
			&project.EggID, &enabled,
		)

		if err != nil {
			return nil, err
		}

		// Parse create time - try multiple formats
		createTime, err := parseTime(createTimeStr)
		if err != nil {
			log.Printf("Error parsing time '%s': %v", createTimeStr, err)
			// Use current time as fallback
			createTime = time.Now()
		}

		project.CreateTime = createTime
		project.CreatedAt = createTime.Format("Jan 2, 2006")

		// Set enabled flag
		project.Enabled = enabled == 1
		project.Public = project.Enabled

		// Set title to be the same as name for template compatibility
		project.Title = project.Name

		// Get star count
		project.Stars, _ = project.GetStarCount()

		projects = append(projects, &project)
	}

	return projects, nil
}

// GetPublicProjects returns all public projects
func GetPublicProjects() ([]*Project, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.github_link, p.create_time, 
		       p.owner, u.name as owner_name, p.environment, p.egg_id, p.enabled,
		       (SELECT COUNT(*) FROM project_stars WHERE project_id = p.id) as stars
		FROM projects p
		LEFT JOIN users u ON p.owner = u.id
		WHERE p.enabled = 1
		ORDER BY p.create_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var project Project
		var createTimeStr string
		var enabled int
		var stars int
		
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.GithubLink,
			&createTimeStr, &project.Owner, &project.OwnerName, &project.Environment,
			&project.EggID, &enabled, &stars)
		
		if err != nil {
			return nil, err
		}
		
		// Parse create time - try multiple formats
		createTime, err := parseTime(createTimeStr)
		if err != nil {
			log.Printf("Error parsing time '%s': %v", createTimeStr, err)
			// Use current time as fallback
			createTime = time.Now()
		}
		
		project.CreateTime = createTime
		project.CreatedAt = createTime.Format("Jan 2, 2006")
		
		// Set enabled flag
		project.Enabled = enabled == 1
		project.Public = project.Enabled
		
		// Set title to be the same as name for template compatibility
		project.Title = project.Name
		
		// Set star count
		project.Stars = stars
		
		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// CreateProject creates a new project
func CreateProject(name, githubLink string, ownerID, eggID int, environment map[string]string) error {
	// Convert environment to JSON string
	envJSON, err := json.Marshal(environment)
	if err != nil {
		return err
	}

	_, err = database.ExecuteStatement(`
		INSERT INTO projects (create_time, name, owner, github_link, environment, egg_id, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, time.Now(), name, ownerID, githubLink, string(envJSON), eggID, 0)

	return err
}

// UpdateProject updates an existing project
func (p *Project) Update() error {
	_, err := database.ExecuteStatement(`
		UPDATE projects
		SET name = ?, github_link = ?, description = ?, environment = ?
		WHERE id = ?
	`, p.Name, p.GithubLink, p.Description, p.Environment, p.ID)

	return err
}

// TogglePublish toggles the enabled status of a project
func (p *Project) TogglePublish() error {
	p.Enabled = !p.Enabled

	_, err := database.ExecuteStatement(
		"UPDATE projects SET enabled = ? WHERE id = ?",
		boolToInt(p.Enabled), p.ID,
	)

	return err
}

// Delete deletes a project
func (p *Project) Delete() error {
	// First delete all stars for this project
	_, err := database.ExecuteStatement("DELETE FROM project_stars WHERE project_id = ?", p.ID)
	if err != nil {
		return err
	}

	// Then delete the project itself
	_, err = database.ExecuteStatement("DELETE FROM projects WHERE id = ?", p.ID)
	return err
}

// IsStarredByUser checks if a project is starred by a user
func (p *Project) IsStarredByUser(userID int) (bool, error) {
	row := database.ExecuteQueryRow(
		"SELECT COUNT(*) FROM project_stars WHERE project_id = ? AND user_id = ?",
		p.ID, userID,
	)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetStarCount returns the number of stars for a project
func (p *Project) GetStarCount() (int, error) {
	row := database.ExecuteQueryRow(
		"SELECT COUNT(*) FROM project_stars WHERE project_id = ?",
		p.ID,
	)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ToggleStar toggles the star status for a project by a user
func (p *Project) ToggleStar(userID int) (bool, error) {
	// Check if the user has already starred this project
	isStarred, err := p.IsStarredByUser(userID)
	if err != nil {
		return false, err
	}

	if isStarred {
		// User has already starred this project, so remove the star
		_, err = database.ExecuteStatement(
			"DELETE FROM project_stars WHERE project_id = ? AND user_id = ?",
			p.ID, userID,
		)
		return false, err
	} else {
		// User has not starred this project, so add a star
		_, err = database.ExecuteStatement(
			"INSERT INTO project_stars (project_id, user_id) VALUES (?, ?)",
			p.ID, userID,
		)
		return true, err
	}
}

// GetEnvironmentVariables returns the environment variables for a project
func (p *Project) GetEnvironmentVariables() (map[string]string, error) {
	if p.Environment == "" {
		return make(map[string]string), nil
	}

	var env map[string]string
	err := json.Unmarshal([]byte(p.Environment), &env)
	if err != nil {
		return make(map[string]string), err
	}

	return env, nil
}

// Helper function to parse time with multiple formats
func parseTime(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	var err error
	var t time.Time

	for _, format := range formats {
		t, err = time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	// Return the last error
	return time.Time{}, err
}

// Helper function to convert bool to int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
