package router

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jimbo/blandmockapi/internal/models"
)

func TestHandler_BasicResponse(t *testing.T) {
	endpoint := models.EndpointConfig{
		Path:     "/test",
		Method:   "GET",
		Status:   200,
		Response: `{"message": "hello"}`,
	}

	handler := Handler(endpoint)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := `{"message": "hello"}`
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}
}

func TestHandler_CustomHeaders(t *testing.T) {
	endpoint := models.EndpointConfig{
		Path:   "/test",
		Method: "GET",
		Status: 200,
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
			"X-API-Version":   "1.0",
		},
		Response: "{}",
	}

	handler := Handler(endpoint)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", w.Header().Get("X-Custom-Header"))
	}

	if w.Header().Get("X-API-Version") != "1.0" {
		t.Errorf("Expected X-API-Version '1.0', got '%s'", w.Header().Get("X-API-Version"))
	}
}

func TestHandler_DefaultContentType(t *testing.T) {
	endpoint := models.EndpointConfig{
		Path:     "/test",
		Method:   "GET",
		Status:   200,
		Response: "{}",
	}

	handler := Handler(endpoint)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected default Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestHandler_CustomContentType(t *testing.T) {
	endpoint := models.EndpointConfig{
		Path:   "/test",
		Method: "GET",
		Status: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Response: "plain text",
	}

	handler := Handler(endpoint)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestHandler_WithDelay(t *testing.T) {
	endpoint := models.EndpointConfig{
		Path:     "/slow",
		Method:   "GET",
		Status:   200,
		Delay:    100, // 100ms
		Response: "{}",
	}

	handler := Handler(endpoint)

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	handler(w, req)
	duration := time.Since(start)

	if duration < 100*time.Millisecond {
		t.Errorf("Expected delay of at least 100ms, got %v", duration)
	}
}

func TestHandler_StatusCodes(t *testing.T) {
	tests := []struct {
		status   int
		expected int
	}{
		{200, 200},
		{201, 201},
		{204, 204},
		{400, 400},
		{401, 401},
		{404, 404},
		{500, 500},
		{0, 200}, // Default to 200
	}

	for _, tt := range tests {
		endpoint := models.EndpointConfig{
			Path:     "/test",
			Method:   "GET",
			Status:   tt.status,
			Response: "{}",
		}

		handler := Handler(endpoint)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != tt.expected {
			t.Errorf("For status %d, expected %d, got %d", tt.status, tt.expected, w.Code)
		}
	}
}

func TestProcessResponse_PathVariable(t *testing.T) {
	response := `{"path": "{{path}}"}`

	req := httptest.NewRequest("GET", "/api/users/123", nil)
	result := processResponse(response, req)

	expected := `{"path": "/api/users/123"}`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestProcessResponse_MethodVariable(t *testing.T) {
	response := `{"method": "{{method}}"}`

	req := httptest.NewRequest("POST", "/api/test", nil)
	result := processResponse(response, req)

	expected := `{"method": "POST"}`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestProcessResponse_QueryParameter(t *testing.T) {
	response := `{"name": "{{query.name}}", "age": "{{query.age}}"}`

	req := httptest.NewRequest("GET", "/api/test?name=Alice&age=30", nil)
	result := processResponse(response, req)

	expected := `{"name": "Alice", "age": "30"}`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestProcessResponse_RequestBody(t *testing.T) {
	response := `{"received": {{body}}}`

	body := `{"name":"Bob"}`
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	result := processResponse(response, req)

	expected := `{"received": {"name":"Bob"}}`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestHealthHandler(t *testing.T) {
	handler := HealthHandler()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}

	expectedBody := `{"status":"healthy","service":"blandmockapi"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestNotFoundHandler(t *testing.T) {
	handler := NotFoundHandler()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
}
