package web

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a type for context keys
type contextKey string

const (
	// contextKeyParams is the key for URL parameters in context
	contextKeyParams contextKey = "params"
)

// Params represents URL parameters
type Params map[string]string

// GetParams extracts parameters from the request context
func GetParams(r *http.Request) Params {
	if params, ok := r.Context().Value(contextKeyParams).(Params); ok {
		return params
	}
	return make(Params)
}

// HandlerFunc is an enhanced handler function with error handling
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Route represents a single route
type Route struct {
	Method      string
	Pattern     string
	Handler     HandlerFunc
	Middlewares []Middleware
}

// Router is an enhanced HTTP router
type Router struct {
	routes      []Route
	middlewares []Middleware
	notFound    http.HandlerFunc
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes:   make([]Route, 0),
		notFound: http.NotFound,
	}
}

// Use adds a middleware to the router
func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// NotFound sets the handler for 404 responses
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.notFound = handler
}

// Handle registers a new route
func (r *Router) Handle(method, pattern string, handler HandlerFunc, middlewares ...Middleware) {
	r.routes = append(r.routes, Route{
		Method:      method,
		Pattern:     pattern,
		Handler:     handler,
		Middlewares: middlewares,
	})
}

// GET registers a GET route
func (r *Router) GET(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	r.Handle(http.MethodGet, pattern, handler, middlewares...)
}

// POST registers a POST route
func (r *Router) POST(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	r.Handle(http.MethodPost, pattern, handler, middlewares...)
}

// DELETE registers a DELETE route
func (r *Router) DELETE(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	r.Handle(http.MethodDelete, pattern, handler, middlewares...)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Find matching route
	for _, route := range r.routes {
		if route.Method != req.Method {
			continue
		}
		
		params, matched := r.matchPattern(route.Pattern, req.URL.Path)
		if !matched {
			continue
		}
		
		// Add params to context
		ctx := context.WithValue(req.Context(), contextKeyParams, params)
		req = req.WithContext(ctx)
		
		// Build handler chain
		var handler http.Handler = r.wrapHandler(route.Handler)
		
		// Apply route-specific middlewares
		for i := len(route.Middlewares) - 1; i >= 0; i-- {
			handler = route.Middlewares[i](handler)
		}
		
		// Apply global middlewares
		for i := len(r.middlewares) - 1; i >= 0; i-- {
			handler = r.middlewares[i](handler)
		}
		
		handler.ServeHTTP(w, req)
		return
	}
	
	// No route found
	r.notFound(w, req)
}

// matchPattern matches a URL path against a pattern and extracts parameters
func (r *Router) matchPattern(pattern, path string) (Params, bool) {
	params := make(Params)
	
	// Split pattern and path
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	
	// Different lengths mean no match (unless pattern ends with *)
	if len(patternParts) != len(pathParts) {
		// Check for wildcard at the end
		if len(patternParts) > 0 && patternParts[len(patternParts)-1] == "*" {
			if len(pathParts) < len(patternParts)-1 {
				return nil, false
			}
			patternParts = patternParts[:len(patternParts)-1]
		} else {
			return nil, false
		}
	}
	
	// Match each part
	for i := 0; i < len(patternParts); i++ {
		patternPart := patternParts[i]
		pathPart := pathParts[i]
		
		if strings.HasPrefix(patternPart, ":") {
			// Parameter
			paramName := patternPart[1:]
			params[paramName] = pathPart
		} else if patternPart != pathPart {
			// Not a match
			return nil, false
		}
	}
	
	return params, true
}

// wrapHandler wraps a HandlerFunc to handle errors
func (r *Router) wrapHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := handler(w, req); err != nil {
			// Handle error - could be customized
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Group creates a sub-router with a common prefix
type Group struct {
	router      *Router
	prefix      string
	middlewares []Middleware
}

// Group creates a new route group
func (r *Router) Group(prefix string, middlewares ...Middleware) *Group {
	return &Group{
		router:      r,
		prefix:      prefix,
		middlewares: middlewares,
	}
}

// Handle registers a route in the group
func (g *Group) Handle(method, pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := g.prefix + pattern
	allMiddlewares := append(g.middlewares, middlewares...)
	g.router.Handle(method, fullPattern, handler, allMiddlewares...)
}

// GET registers a GET route in the group
func (g *Group) GET(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	g.Handle(http.MethodGet, pattern, handler, middlewares...)
}

// POST registers a POST route in the group
func (g *Group) POST(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	g.Handle(http.MethodPost, pattern, handler, middlewares...)
}

// DELETE registers a DELETE route in the group
func (g *Group) DELETE(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	g.Handle(http.MethodDelete, pattern, handler, middlewares...)
}