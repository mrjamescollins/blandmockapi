package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	loader := New()

	if loader == nil {
		t.Fatal("New() returned nil")
	}

	cfg := loader.GetConfig()
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host 0.0.0.0, got %s", cfg.Server.Host)
	}
}

func TestLoadFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	configContent := `
[server]
port = 9000
host = "localhost"

[[endpoints]]
path = "/test"
method = "GET"
status = 200
response = '{"test": true}'
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	loader := New()
	if err := loader.LoadFile(configPath); err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	cfg := loader.GetConfig()

	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", cfg.Server.Host)
	}

	if len(cfg.Endpoints) != 1 {
		t.Fatalf("Expected 1 endpoint, got %d", len(cfg.Endpoints))
	}

	if cfg.Endpoints[0].Path != "/test" {
		t.Errorf("Expected path /test, got %s", cfg.Endpoints[0].Path)
	}
}

func TestLoadDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple config files
	config1 := `
[[endpoints]]
path = "/endpoint1"
method = "GET"
status = 200
response = '{"id": 1}'
`

	config2 := `
[[endpoints]]
path = "/endpoint2"
method = "POST"
status = 201
response = '{"id": 2}'
`

	if err := os.WriteFile(filepath.Join(tmpDir, "config1.toml"), []byte(config1), 0644); err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "config2.toml"), []byte(config2), 0644); err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	// Create a non-toml file that should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	loader := New()
	if err := loader.LoadDirectory(tmpDir); err != nil {
		t.Fatalf("LoadDirectory failed: %v", err)
	}

	cfg := loader.GetConfig()

	if len(cfg.Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(cfg.Endpoints))
	}
}

func TestLoadFromPath_File(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	configContent := `
[server]
port = 3000
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	loader := New()
	if err := loader.LoadFromPath(configPath); err != nil {
		t.Fatalf("LoadFromPath failed: %v", err)
	}

	cfg := loader.GetConfig()
	if cfg.Server.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Server.Port)
	}
}

func TestLoadFromPath_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
[[endpoints]]
path = "/test"
method = "GET"
status = 200
response = '{}'
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	loader := New()
	if err := loader.LoadFromPath(tmpDir); err != nil {
		t.Fatalf("LoadFromPath failed: %v", err)
	}

	cfg := loader.GetConfig()
	if len(cfg.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(cfg.Endpoints))
	}
}

func TestMergeConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// First config
	config1 := `
[server]
port = 8080

[[endpoints]]
path = "/api/v1"
method = "GET"
status = 200
response = '{}'
`

	// Second config overrides port
	config2 := `
[server]
port = 9090

[[endpoints]]
path = "/api/v2"
method = "POST"
status = 201
response = '{}'
`

	if err := os.WriteFile(filepath.Join(tmpDir, "01-base.toml"), []byte(config1), 0644); err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "02-override.toml"), []byte(config2), 0644); err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	loader := New()
	if err := loader.LoadDirectory(tmpDir); err != nil {
		t.Fatalf("LoadDirectory failed: %v", err)
	}

	cfg := loader.GetConfig()

	// Last loaded port should override
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090 (from override), got %d", cfg.Server.Port)
	}

	// Both endpoints should be present
	if len(cfg.Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(cfg.Endpoints))
	}
}

func TestLoadInvalidPath(t *testing.T) {
	loader := New()
	err := loader.LoadFromPath("/nonexistent/path/config.toml")

	if err == nil {
		t.Error("Expected error for nonexistent path, got nil")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")

	invalidContent := `
[server
port = "not a number"
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	loader := New()
	err := loader.LoadFile(configPath)

	if err == nil {
		t.Error("Expected error for invalid TOML, got nil")
	}
}
