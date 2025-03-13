package utils

import (
	"regexp"
	"sort"
	"strings"
	
	"codehubgo/internal/models"
)

// SearchResult represents a search result with a project and its score
type SearchResult struct {
	Project *models.Project
	Score   float64
}

// preprocessText prepares text for searching by:
// - Converting to lowercase
// - Removing special characters
// - Removing extra whitespace
func preprocessText(text string) string {
	if text == "" {
		return ""
	}
	
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Remove quotes and special characters
	re := regexp.MustCompile(`['"\[\](){}<>]`)
	text = re.ReplaceAllString(text, "")
	
	// Replace multiple spaces with a single space
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

// calculateSimilarity calculates the similarity between two strings
// Returns a score between 0 and 1, where 1 is a perfect match
func calculateSimilarity(query, text string) float64 {
	if query == "" || text == "" {
		return 0
	}
	
	// Preprocess both query and text
	query = preprocessText(query)
	text = preprocessText(text)
	
	// If either is empty after preprocessing, return 0
	if query == "" || text == "" {
		return 0
	}
	
	// Simple contains check gives a base score
	if strings.Contains(text, query) {
		return 0.8
	}
	
	// Check for partial matches
	queryWords := strings.Fields(query)
	textWords := strings.Fields(text)
	
	matches := 0
	for _, qw := range queryWords {
		for _, tw := range textWords {
			if strings.Contains(tw, qw) || strings.Contains(qw, tw) {
				matches++
				break
			}
		}
	}
	
	if len(queryWords) > 0 {
		return float64(matches) / float64(len(queryWords))
	}
	
	return 0
}

// SearchProjects searches projects by query using fuzzy matching
// Returns a list of SearchResult sorted by score in descending order
func SearchProjects(projects []*models.Project, query string) []SearchResult {
	if query == "" || len(projects) == 0 {
		return nil
	}
	
	// Preprocess query
	processedQuery := preprocessText(query)
	
	// If query is empty after preprocessing, return all projects
	if processedQuery == "" {
		results := make([]SearchResult, len(projects))
		for i, project := range projects {
			results[i] = SearchResult{Project: project, Score: 1.0}
		}
		return results
	}
	
	minScore := 0.3 // Minimum similarity score to include in results
	results := []SearchResult{}
	
	for _, project := range projects {
		// Calculate similarity scores for different fields
		nameScore := calculateSimilarity(processedQuery, project.Name) * 1.5 // Weight name higher
		ownerNameScore := calculateSimilarity(processedQuery, project.OwnerName)
		descriptionScore := calculateSimilarity(processedQuery, project.Description)
		githubScore := calculateSimilarity(processedQuery, project.GithubLink)
		
		// Get the maximum score
		maxScore := nameScore
		if ownerNameScore > maxScore {
			maxScore = ownerNameScore
		}
		if descriptionScore > maxScore {
			maxScore = descriptionScore
		}
		if githubScore > maxScore {
			maxScore = githubScore
		}
		
		// If any field has a score above the threshold, include the project
		if maxScore >= minScore {
			results = append(results, SearchResult{Project: project, Score: maxScore})
		}
	}
	
	// Sort by score in descending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}
