package router

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jimbo/blandmockapi/internal/models"
)

// Handler creates an HTTP handler for a configured endpoint
func Handler(endpoint models.EndpointConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Apply configured delay if specified
		if endpoint.Delay > 0 {
			time.Sleep(time.Duration(endpoint.Delay) * time.Millisecond)
		}

		// Set configured headers
		for key, value := range endpoint.Headers {
			w.Header().Set(key, value)
		}

		// Set default Content-Type if not specified
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		// Set status code
		status := endpoint.Status
		if status == 0 {
			status = 200
		}
		w.WriteHeader(status)

		// Process and write response
		response := processResponse(endpoint.Response, r)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}

// processResponse handles response templating with request data
func processResponse(response string, r *http.Request) string {
	// Replace common variables
	response = strings.ReplaceAll(response, "{{path}}", r.URL.Path)
	response = strings.ReplaceAll(response, "{{method}}", r.Method)

	// Replace query parameters
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			response = strings.ReplaceAll(response, fmt.Sprintf("{{query.%s}}", key), values[0])
		}
	}

	// Replace path parameters (simple implementation)
	// For more complex routing, could integrate a router library

	// Try to parse and include request body if it's JSON
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		if body, err := io.ReadAll(r.Body); err == nil {
			var jsonBody interface{}
			if err := json.Unmarshal(body, &jsonBody); err == nil {
				if bodyJSON, err := json.Marshal(jsonBody); err == nil {
					response = strings.ReplaceAll(response, "{{body}}", string(bodyJSON))
				}
			}
		}
	}

	return response
}

// HealthHandler returns a basic health check handler
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy","service":"blandmockapi"}`)); err != nil {
			log.Printf("Failed to write health response: %v", err)
		}
	}
}

// NotFoundHandler returns a custom 404 handler
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[404] %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := fmt.Sprintf(`{"error":"endpoint not found","path":"%s","method":"%s"}`, r.URL.Path, r.Method)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Printf("Failed to write 404 response: %v", err)
		}
	}
}
