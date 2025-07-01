package web

import (
	"embed"
	"io/fs"
)

// Embed the built React app
//
//go:embed dist/*
var distFS embed.FS

// Embed API documentation files
//
//go:embed all:api/*
var apiDocsFS embed.FS

// GetDistFS returns the embedded dist filesystem.
func GetDistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

// GetStaticFS returns the embedded static filesystem.
func GetStaticFS() (fs.FS, error) {
	return GetDistFS()
}

// GetTemplatesFS returns the embedded templates filesystem.
func GetTemplatesFS() (fs.FS, error) {
	return GetDistFS()
}

// GetAPIDocsFS returns the embedded API documentation filesystem.
func GetAPIDocsFS() (fs.FS, error) {
	return fs.Sub(apiDocsFS, "api")
}
