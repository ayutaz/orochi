package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// handleAPIDocs serves the Swagger UI page.
func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	// Serve the swagger.html file
	apiDocsPath := filepath.Join("api", "swagger.html")
	
	// Check if file exists
	if _, err := os.Stat(apiDocsPath); os.IsNotExist(err) {
		http.Error(w, "API documentation not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, apiDocsPath)
}

// handleAPIDocsStatic serves static files for API documentation.
func (s *Server) handleAPIDocsStatic(w http.ResponseWriter, r *http.Request) {
	// Get the requested path
	path := GetParams(r)["*"]
	if path == "" {
		path = "openapi.yaml"
	}

	// Remove any leading slash
	path = strings.TrimPrefix(path, "/")

	// Construct the full path
	fullPath := filepath.Join("api", path)

	// Serve only specific files for security
	allowedFiles := map[string]bool{
		"openapi.yaml": true,
		"swagger.html": true,
	}

	if !allowedFiles[path] {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set content type for YAML files
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		w.Header().Set("Content-Type", "application/x-yaml")
	}

	http.ServeFile(w, r, fullPath)
}