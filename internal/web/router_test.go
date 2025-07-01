package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_MatchPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		path        string
		shouldMatch bool
		params      Params
	}{
		{
			name:        "Exact match",
			pattern:     "/api/torrents",
			path:        "/api/torrents",
			shouldMatch: true,
			params:      Params{},
		},
		{
			name:        "Parameter match",
			pattern:     "/api/torrents/:id",
			path:        "/api/torrents/123",
			shouldMatch: true,
			params:      Params{"id": "123"},
		},
		{
			name:        "Multiple parameters",
			pattern:     "/api/torrents/:id/files/:file",
			path:        "/api/torrents/123/files/readme.txt",
			shouldMatch: true,
			params:      Params{"id": "123", "file": "readme.txt"},
		},
		{
			name:        "No match - different path",
			pattern:     "/api/torrents",
			path:        "/api/files",
			shouldMatch: false,
			params:      nil,
		},
		{
			name:        "No match - shorter path",
			pattern:     "/api/torrents/:id",
			path:        "/api/torrents",
			shouldMatch: false,
			params:      nil,
		},
		{
			name:        "Wildcard match",
			pattern:     "/static/*",
			path:        "/static/js/app.js",
			shouldMatch: true,
			params:      Params{},
		},
		{
			name:        "Root path",
			pattern:     "/",
			path:        "/",
			shouldMatch: true,
			params:      Params{},
		},
		{
			name:        "Trailing slash normalization",
			pattern:     "/api/torrents",
			path:        "/api/torrents/",
			shouldMatch: true,
			params:      Params{},
		},
	}
	
	router := NewRouter()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, matched := router.matchPattern(tt.pattern, tt.path)
			
			if matched != tt.shouldMatch {
				t.Errorf("expected match=%v, got %v", tt.shouldMatch, matched)
			}
			
			if tt.shouldMatch && tt.params != nil {
				for key, expectedValue := range tt.params {
					if value, ok := params[key]; !ok || value != expectedValue {
						t.Errorf("expected param %s=%s, got %s", key, expectedValue, value)
					}
				}
			}
		})
	}
}

func TestRouter_Routing(t *testing.T) {
	router := NewRouter()
	
	// Test data
	var calledRoute string
	
	// Register routes
	router.GET("/", func(_ http.ResponseWriter, _ *http.Request) error {
		calledRoute = "home"
		return nil
	})
	
	router.GET("/api/torrents", func(w http.ResponseWriter, r *http.Request) error {
		calledRoute = "list"
		return nil
	})
	
	router.POST("/api/torrents", func(w http.ResponseWriter, r *http.Request) error {
		calledRoute = "create"
		return nil
	})
	
	router.GET("/api/torrents/:id", func(w http.ResponseWriter, r *http.Request) error {
		params := GetParams(r)
		calledRoute = "get:" + params["id"]
		return nil
	})
	
	router.DELETE("/api/torrents/:id", func(w http.ResponseWriter, r *http.Request) error {
		params := GetParams(r)
		calledRoute = "delete:" + params["id"]
		return nil
	})
	
	tests := []struct {
		method       string
		path         string
		expectedCall string
		expectedCode int
	}{
		{"GET", "/", "home", http.StatusOK},
		{"GET", "/api/torrents", "list", http.StatusOK},
		{"POST", "/api/torrents", "create", http.StatusOK},
		{"GET", "/api/torrents/123", "get:123", http.StatusOK},
		{"DELETE", "/api/torrents/456", "delete:456", http.StatusOK},
		{"GET", "/not/found", "", http.StatusNotFound},
		{"PUT", "/api/torrents", "", http.StatusNotFound},
	}
	
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			calledRoute = ""
			
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			
			router.ServeHTTP(rec, req)
			
			if calledRoute != tt.expectedCall {
				t.Errorf("expected route %s, got %s", tt.expectedCall, calledRoute)
			}
			
			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rec.Code)
			}
		})
	}
}

func TestRouter_Middleware(t *testing.T) {
	router := NewRouter()
	
	// Track middleware execution order
	var executionOrder []string
	
	// Global middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "global1")
			next.ServeHTTP(w, r)
		})
	})
	
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "global2")
			next.ServeHTTP(w, r)
		})
	})
	
	// Route with middleware
	router.GET("/test", func(w http.ResponseWriter, r *http.Request) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "route")
			next.ServeHTTP(w, r)
		})
	})
	
	// Execute request
	executionOrder = []string{}
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	// Check execution order
	expectedOrder := []string{"global1", "global2", "route", "handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("expected %d calls, got %d", len(expectedOrder), len(executionOrder))
	}
	
	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("execution order[%d]: expected %s, got %s", i, expected, executionOrder[i])
		}
	}
}

func TestRouter_Group(t *testing.T) {
	router := NewRouter()
	
	var calledRoute string
	
	// Create API group
	api := router.Group("/api")
	api.GET("/torrents", func(w http.ResponseWriter, r *http.Request) error {
		calledRoute = "api-torrents"
		return nil
	})
	api.GET("/status", func(w http.ResponseWriter, r *http.Request) error {
		calledRoute = "api-status"
		return nil
	})
	
	// Create admin group with middleware
	var adminCheck bool
	admin := router.Group("/admin", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			adminCheck = true
			next.ServeHTTP(w, r)
		})
	})
	admin.GET("/users", func(w http.ResponseWriter, r *http.Request) error {
		calledRoute = "admin-users"
		return nil
	})
	
	tests := []struct {
		path          string
		expectedCall  string
		expectAdmin   bool
	}{
		{"/api/torrents", "api-torrents", false},
		{"/api/status", "api-status", false},
		{"/admin/users", "admin-users", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			calledRoute = ""
			adminCheck = false
			
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			
			router.ServeHTTP(rec, req)
			
			if calledRoute != tt.expectedCall {
				t.Errorf("expected route %s, got %s", tt.expectedCall, calledRoute)
			}
			
			if adminCheck != tt.expectAdmin {
				t.Errorf("expected admin check %v, got %v", tt.expectAdmin, adminCheck)
			}
		})
	}
}