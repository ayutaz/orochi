package web

import (
	"io"
	"net/http"
	"strings"

	"github.com/ayutaz/orochi/internal/logger"
)

// handleAPIDocs serves the Swagger UI page.
func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	apiFS, err := GetAPIDocsFS()
	if err != nil {
		s.logger.Error("failed to get API docs filesystem", logger.Err(err))
		http.Error(w, "API documentation not available", http.StatusInternalServerError)
		return
	}

	// Open swagger.html from embedded filesystem
	file, err := apiFS.Open("swagger.html")
	if err != nil {
		s.logger.Error("failed to open swagger.html", logger.Err(err))
		http.Error(w, "API documentation not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Copy file contents to response
	if _, err := io.Copy(w, file); err != nil {
		s.logger.Error("failed to serve swagger.html", logger.Err(err))
	}
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

	// Serve only specific files for security
	allowedFiles := map[string]bool{
		"openapi.yaml": true,
		"swagger.html": true,
	}

	if !allowedFiles[path] {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	apiFS, err := GetAPIDocsFS()
	if err != nil {
		s.logger.Error("failed to get API docs filesystem", logger.Err(err))
		http.Error(w, "API documentation not available", http.StatusInternalServerError)
		return
	}

	// Open file from embedded filesystem
	file, err := apiFS.Open(path)
	if err != nil {
		s.logger.Error("failed to open API docs file", logger.String("path", path), logger.Err(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set content type
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
	} else if strings.HasSuffix(path, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	// Copy file contents to response
	if _, err := io.Copy(w, file); err != nil {
		s.logger.Error("failed to serve API docs file", logger.String("path", path), logger.Err(err))
	}
}
