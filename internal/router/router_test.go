package router

import (
	"net/http/httptest"
	"testing"

	"github.com/jimbo/blandmockapi/internal/models"
)

func TestNew(t *testing.T) {
	router := New()

	if router == nil {
		t.Fatal("New() returned nil")
	}

	if router.mux == nil {
		t.Error("Router mux is nil")
	}

	if router.pathMethods == nil {
		t.Error("Router pathMethods is nil")
	}
}

func TestRegisterEndpoint(t *testing.T) {
	router := New()

	endpoint := models.EndpointConfig{
		Path:     "/test",
		Method:   "GET",
		Status:   200,
		Response: `{"test": true}`,
	}

	err := router.RegisterEndpoint(endpoint)
	if err != nil {
		t.Fatalf("RegisterEndpoint failed: %v", err)
	}

	if len(router.endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(router.endpoints))
	}

	if router.endpoints[0].Path != "/test" {
		t.Errorf("Expected path /test, got %s", router.endpoints[0].Path)
	}
}

func TestRegisterEndpoint_EmptyPath(t *testing.T) {
	router := New()

	endpoint := models.EndpointConfig{
		Method:   "GET",
		Status:   200,
		Response: "{}",
	}

	err := router.RegisterEndpoint(endpoint)
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
}

func TestRegisterEndpoint_DefaultMethod(t *testing.T) {
	router := New()

	endpoint := models.EndpointConfig{
		Path:     "/test",
		Status:   200,
		Response: "{}",
	}

	err := router.RegisterEndpoint(endpoint)
	if err != nil {
		t.Fatalf("RegisterEndpoint failed: %v", err)
	}

	if router.endpoints[0].Method != "GET" {
		t.Errorf("Expected default method GET, got %s", router.endpoints[0].Method)
	}
}

func TestRegisterMultipleMethodsSamePath(t *testing.T) {
	router := New()

	endpoints := []models.EndpointConfig{
		{Path: "/api/users", Method: "GET", Status: 200, Response: `{"users": []}`},
		{Path: "/api/users", Method: "POST", Status: 201, Response: `{"id": 1}`},
		{Path: "/api/users", Method: "PUT", Status: 200, Response: `{"updated": true}`},
	}

	for _, ep := range endpoints {
		if err := router.RegisterEndpoint(ep); err != nil {
			t.Fatalf("Failed to register %s %s: %v", ep.Method, ep.Path, err)
		}
	}

	if len(router.endpoints) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(router.endpoints))
	}

	// Verify all methods are registered for the path
	pathMethods := router.pathMethods["/api/users"]
	if len(pathMethods) != 3 {
		t.Errorf("Expected 3 methods for /api/users, got %d", len(pathMethods))
	}

	if _, exists := pathMethods["GET"]; !exists {
		t.Error("GET method not registered")
	}
	if _, exists := pathMethods["POST"]; !exists {
		t.Error("POST method not registered")
	}
	if _, exists := pathMethods["PUT"]; !exists {
		t.Error("PUT method not registered")
	}
}

func TestRouterHandler_Success(t *testing.T) {
	router := New()

	endpoint := models.EndpointConfig{
		Path:     "/test",
		Method:   "GET",
		Status:   200,
		Response: `{"message": "success"}`,
		Headers: map[string]string{
			"X-Custom": "value",
		},
	}

	router.RegisterEndpoint(endpoint)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("X-Custom") != "value" {
		t.Errorf("Expected X-Custom header 'value', got '%s'", w.Header().Get("X-Custom"))
	}

	expectedBody := `{"message": "success"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestRouterHandler_NotFound(t *testing.T) {
	router := New()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.Handler().ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestRouterHandler_MethodNotAllowed(t *testing.T) {
	router := New()

	endpoint := models.EndpointConfig{
		Path:     "/test",
		Method:   "GET",
		Status:   200,
		Response: "{}",
	}

	router.RegisterEndpoint(endpoint)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	router.Handler().ServeHTTP(w, req)

	if w.Code != 405 {
		t.Errorf("Expected status 405, got %d", w.Code)
	}

	if w.Header().Get("Allow") == "" {
		t.Error("Expected Allow header to be set")
	}
}

func TestRegisterHealthCheck(t *testing.T) {
	router := New()
	router.RegisterHealthCheck()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json")
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"/api/users", "/api/users", true},
		{"/api/users", "/api/users/", true},
		{"/api/users/", "/api/users", true},
		{"/api/", "/api/users", true},
		{"/api/", "/api/users/123", true},
		{"/api/users", "/api/products", false},
		{"/api/users", "/api", false},
	}

	for _, tt := range tests {
		got := matchesPattern(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}

func TestGetEndpoints(t *testing.T) {
	router := New()

	endpoints := []models.EndpointConfig{
		{Path: "/api/v1", Method: "GET", Status: 200, Response: "{}"},
		{Path: "/api/v2", Method: "POST", Status: 201, Response: "{}"},
	}

	for _, ep := range endpoints {
		router.RegisterEndpoint(ep)
	}

	retrieved := router.GetEndpoints()

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(retrieved))
	}
}
