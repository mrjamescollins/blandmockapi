// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

const (
	baseURL    = "http://localhost:8080"
	serverWait = 2 * time.Second
)

var serverCmd *exec.Cmd

// TestMain sets up and tears down the test server
func TestMain(m *testing.M) {
	// Start the server
	serverCmd = exec.Command("go", "run", "../../cmd/server", "-config", "../../examples")
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for server to be ready
	time.Sleep(serverWait)

	if !isServerReady() {
		shutdownServer()
		fmt.Println("Server failed to become ready")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Shutdown server gracefully
	shutdownServer()

	os.Exit(code)
}

func shutdownServer() {
	if serverCmd == nil || serverCmd.Process == nil {
		return
	}

	// Send interrupt signal to process group to ensure all child processes receive it
	// For go run, this ensures both the wrapper and actual server process get the signal
	pgid, err := syscall.Getpgid(serverCmd.Process.Pid)
	if err == nil {
		// Send signal to process group
		syscall.Kill(-pgid, syscall.SIGINT)
	} else {
		// Fallback to sending to just the process
		serverCmd.Process.Signal(os.Interrupt)
	}

	// Wait for server to shutdown gracefully with timeout
	done := make(chan error, 1)
	go func() {
		done <- serverCmd.Wait()
	}()

	select {
	case err := <-done:
		// Server exited - ignore exit errors from interrupt signal
		if err != nil {
			// Check if it's just an interrupt signal exit
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Exit code -1 or signal interrupt is expected
				if exitErr.ProcessState.ExitCode() == -1 {
					// This is normal for interrupted process
					return
				}
			}
		}
	case <-time.After(5 * time.Second):
		// Timeout - force kill the process group
		if pgid, err := syscall.Getpgid(serverCmd.Process.Pid); err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			serverCmd.Process.Kill()
		}
		serverCmd.Wait() // Reap the process
	}
}

func isServerReady() bool {
	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", result["status"])
	}

	if result["service"] != "blandmockapi" {
		t.Errorf("Expected service 'blandmockapi', got '%s'", result["service"])
	}
}

func TestRESTGetEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to GET /api/users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	users, ok := result["users"].([]interface{})
	if !ok {
		t.Fatal("Expected 'users' array in response")
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestRESTPostEndpoint(t *testing.T) {
	payload := map[string]string{
		"name":  "Test User",
		"email": "test@example.com",
	}

	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/api/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to POST /api/users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["message"] != "User created successfully" {
		t.Errorf("Unexpected message: %v", result["message"])
	}
}

func TestMethodNotAllowed(t *testing.T) {
	req, _ := http.NewRequest("PATCH", baseURL+"/api/users", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 405 {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestNotFoundEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/nonexistent/path")
	if err != nil {
		t.Fatalf("Failed to GET nonexistent path: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestGraphQLQuery(t *testing.T) {
	query := map[string]string{
		"query": "{ users { id name email } }",
	}

	jsonData, _ := json.Marshal(query)

	resp, err := http.Post(baseURL+"/graphql", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to POST GraphQL query: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'data' in GraphQL response")
	}

	users, ok := data["users"].([]interface{})
	if !ok {
		t.Fatal("Expected 'users' array in GraphQL response")
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users from GraphQL, got %d", len(users))
	}
}

func TestGraphQLMutation(t *testing.T) {
	mutation := map[string]string{
		"query": `mutation { createUser(name: "Jane", email: "jane@example.com") { id name email } }`,
	}

	jsonData, _ := json.Marshal(mutation)

	resp, err := http.Post(baseURL+"/graphql", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to POST GraphQL mutation: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'data' in GraphQL response")
	}

	user, ok := data["createUser"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'createUser' in GraphQL response")
	}

	if user["name"] == nil {
		t.Error("Expected 'name' in created user")
	}
}

func TestSlowEndpoint(t *testing.T) {
	start := time.Now()

	resp, err := http.Get(baseURL + "/api/slow")
	if err != nil {
		t.Fatalf("Failed to GET /api/slow: %v", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if duration < 3*time.Second {
		t.Errorf("Expected delay of at least 3 seconds, got %v", duration)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestErrorEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/error")
	if err != nil {
		t.Fatalf("Failed to GET /api/error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["error"] != "Internal Server Error" {
		t.Errorf("Unexpected error message: %v", result["error"])
	}
}

func TestCustomHeaders(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to GET /api/users: %v", err)
	}
	defer resp.Body.Close()

	apiVersion := resp.Header.Get("X-API-Version")
	if apiVersion != "1.0" {
		t.Errorf("Expected X-API-Version '1.0', got '%s'", apiVersion)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}
