package utils

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	
	"github.com/russross/blackfriday/v2"
)

var (
	templates     map[string]*template.Template
	templateMutex sync.RWMutex
)

// Initialize templates
func init() {
	templates = make(map[string]*template.Template)
}

// LoadTemplates loads all templates from the templates directory
func LoadTemplates(templatesDir string) error {
	templateMutex.Lock()
	defer templateMutex.Unlock()
	
	log.Printf("Loading templates from directory: %s", templatesDir)
	
	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Printf("Templates directory does not exist: %s", templatesDir)
		return err
	}
	
	// Define template functions
	funcMap := template.FuncMap{
		"markdownToHTML": MarkdownToHTML,
	}
	
	// Find all template files
	layoutPath := filepath.Join(templatesDir, "layouts")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		log.Printf("Layouts directory does not exist: %s", layoutPath)
		return err
	}
	
	layoutFiles, err := filepath.Glob(filepath.Join(layoutPath, "*.html"))
	if err != nil {
		log.Printf("Error finding layout files: %v", err)
		return err
	}
	log.Printf("Found %d layout files: %v", len(layoutFiles), layoutFiles)
	
	includePath := filepath.Join(templatesDir, "includes")
	if _, err := os.Stat(includePath); os.IsNotExist(err) {
		log.Printf("Includes directory does not exist: %s", includePath)
		return err
	}
	
	includeFiles, err := filepath.Glob(filepath.Join(includePath, "*.html"))
	if err != nil {
		log.Printf("Error finding include files: %v", err)
		return err
	}
	log.Printf("Found %d include files: %v", len(includeFiles), includeFiles)
	
	// Get all page templates
	pageFiles, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		log.Printf("Error finding page files: %v", err)
		return err
	}
	log.Printf("Found %d page files: %v", len(pageFiles), pageFiles)
	
	// For each page, create a template set with the layouts and includes
	for _, page := range pageFiles {
		fileName := filepath.Base(page)
		log.Printf("Processing template: %s", fileName)
		
		// Combine all files needed for this template
		allFiles := append([]string{page}, layoutFiles...)
		allFiles = append(allFiles, includeFiles...)
		
		// Create a new template set with the functions
		tmpl, err := template.New(fileName).Funcs(funcMap).ParseFiles(allFiles...)
		if err != nil {
			log.Printf("Error parsing template files for %s: %v", fileName, err)
			return err
		}
		
		// Store the template
		templates[fileName] = tmpl
		log.Printf("Template %s loaded successfully", fileName)
	}
	
	log.Printf("All templates loaded successfully. Total templates: %d", len(templates))
	return nil
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templateMutex.RLock()
	t, ok := templates[tmpl]
	templateMutex.RUnlock()
	
	if !ok {
		log.Printf("Template not found: %s", tmpl)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Rendering template: %s", tmpl)
	err := t.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmpl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MarkdownToHTML converts markdown to HTML
func MarkdownToHTML(md string) template.HTML {
	output := blackfriday.Run([]byte(md))
	return template.HTML(output)
}
