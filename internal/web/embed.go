package web

import (
	"embed"
	"io/fs"
)

// Embed static files and templates
//
//go:embed all:templates/* all:static/*
var embedFS embed.FS

// GetStaticFS returns the embedded static files
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(embedFS, "static")
}

// GetTemplatesFS returns the embedded template files
func GetTemplatesFS() (fs.FS, error) {
	return fs.Sub(embedFS, "templates")
}