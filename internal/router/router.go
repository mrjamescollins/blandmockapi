package router

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jimbo/blandmockapi/internal/models"
)

// Router manages HTTP routing for the mock API
type Router struct {
	mux       *http.ServeMux
	endpoints []models.EndpointConfig
	// Map of path -> method -> endpoint for multi-method support
	pathMethods  map[string]map[string]models.EndpointConfig
	graphqlPath  string
	hasGraphQL   bool
}

// New creates a new router
func New() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		endpoints:   []models.EndpointConfig{},
		pathMethods: make(map[string]map[string]models.EndpointConfig),
	}
}

// RegisterEndpoints registers all configured endpoints
func (rt *Router) RegisterEndpoints(endpoints []models.EndpointConfig) error {
	for _, endpoint := range endpoints {
		if err := rt.RegisterEndpoint(endpoint); err != nil {
			return err
		}
	}
	return nil
}

// RegisterEndpoint registers a single endpoint
func (rt *Router) RegisterEndpoint(endpoint models.EndpointConfig) error {
	// Validate endpoint
	if endpoint.Path == "" {
		return fmt.Errorf("endpoint path cannot be empty")
	}
	if endpoint.Method == "" {
		endpoint.Method = "GET"
	}

	// Normalize method to uppercase
	endpoint.Method = strings.ToUpper(endpoint.Method)

	// Check if this path is already registered
	if _, exists := rt.pathMethods[endpoint.Path]; !exists {
		// First time seeing this path - register it with the mux
		rt.pathMethods[endpoint.Path] = make(map[string]models.EndpointConfig)
		rt.mux.HandleFunc(endpoint.Path, rt.multiMethodHandler(endpoint.Path))
	}

	// Store the endpoint config for this method
	rt.pathMethods[endpoint.Path][endpoint.Method] = endpoint
	rt.endpoints = append(rt.endpoints, endpoint)

	log.Printf("Registered endpoint: %s %s -> %d", endpoint.Method, endpoint.Path, endpoint.Status)
	return nil
}

// multiMethodHandler creates a handler that routes based on HTTP method
func (rt *Router) multiMethodHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		methodMap, exists := rt.pathMethods[path]
		if !exists {
			NotFoundHandler()(w, r)
			return
		}

		endpoint, methodExists := methodMap[r.Method]
		if !methodExists {
			// Method not allowed - list allowed methods
			allowed := make([]string, 0, len(methodMap))
			for method := range methodMap {
				allowed = append(allowed, method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(fmt.Sprintf(`{"error":"method not allowed","allowed":%q,"received":"%s"}`, allowed, r.Method)))
			return
		}

		// Call the handler for this specific endpoint
		Handler(endpoint)(w, r)
	}
}

// RegisterHealthCheck registers a health check endpoint
func (rt *Router) RegisterHealthCheck() {
	rt.mux.HandleFunc("/health", HealthHandler())
	log.Printf("Registered health check endpoint: GET /health")
}

// RegisterGraphQL registers a GraphQL endpoint handler
func (rt *Router) RegisterGraphQL(path string, handler http.HandlerFunc) {
	if path == "" {
		path = "/graphql"
	}
	rt.graphqlPath = path
	rt.hasGraphQL = true
	rt.mux.HandleFunc(path, handler)
	log.Printf("Registered GraphQL endpoint: POST %s", path)
}

// Handler returns the underlying HTTP handler
func (rt *Router) Handler() http.Handler {
	// Wrap the mux with a custom handler that provides 404 responses
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if any pattern matches
		pattern := rt.findMatchingPattern(r)
		if pattern != "" {
			rt.mux.ServeHTTP(w, r)
		} else {
			NotFoundHandler()(w, r)
		}
	})
}

// findMatchingPattern checks if a request matches any registered pattern
func (rt *Router) findMatchingPattern(r *http.Request) string {
	// Check health endpoint
	if r.URL.Path == "/health" {
		return "/health"
	}

	// Check GraphQL endpoint
	if rt.hasGraphQL && r.URL.Path == rt.graphqlPath {
		return rt.graphqlPath
	}

	// Check registered endpoints
	for _, ep := range rt.endpoints {
		if matchesPattern(ep.Path, r.URL.Path) {
			return ep.Path
		}
	}

	return ""
}

// matchesPattern checks if a URL path matches a pattern
// Simple implementation - could be enhanced with path parameters
func matchesPattern(pattern, path string) bool {
	// Exact match
	if pattern == path {
		return true
	}

	// Trailing slash handling
	if strings.TrimSuffix(pattern, "/") == strings.TrimSuffix(path, "/") {
		return true
	}

	// Basic wildcard support for patterns ending with /
	if strings.HasSuffix(pattern, "/") && strings.HasPrefix(path, pattern) {
		return true
	}

	return false
}

// GetEndpoints returns all registered endpoints for debugging
func (rt *Router) GetEndpoints() []models.EndpointConfig {
	return rt.endpoints
}
