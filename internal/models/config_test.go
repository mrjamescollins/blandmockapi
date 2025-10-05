package models

import (
	"testing"
	"time"
)

func TestServerConfig_GetReadTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{"default zero", 0, 15 * time.Second},
		{"negative", -5, 15 * time.Second},
		{"custom value", 30, 30 * time.Second},
		{"small value", 5, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{ReadTimeout: tt.timeout}
			got := cfg.GetReadTimeout()

			if got != tt.expected {
				t.Errorf("GetReadTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestServerConfig_GetWriteTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{"default zero", 0, 15 * time.Second},
		{"negative", -10, 15 * time.Second},
		{"custom value", 45, 45 * time.Second},
		{"small value", 1, 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{WriteTimeout: tt.timeout}
			got := cfg.GetWriteTimeout()

			if got != tt.expected {
				t.Errorf("GetWriteTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestServerConfig_GetPort(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected int
	}{
		{"default zero", 0, 8080},
		{"negative", -1, 8080},
		{"custom port", 3000, 3000},
		{"high port", 9999, 9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{Port: tt.port}
			got := cfg.GetPort()

			if got != tt.expected {
				t.Errorf("GetPort() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestServerConfig_GetHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{"default empty", "", "0.0.0.0"},
		{"localhost", "localhost", "localhost"},
		{"specific IP", "192.168.1.1", "192.168.1.1"},
		{"all interfaces", "0.0.0.0", "0.0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{Host: tt.host}
			got := cfg.GetHost()

			if got != tt.expected {
				t.Errorf("GetHost() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEndpointConfig(t *testing.T) {
	endpoint := EndpointConfig{
		Path:        "/api/users",
		Method:      "GET",
		Status:      200,
		Response:    `{"users": []}`,
		Headers:     map[string]string{"X-Custom": "value"},
		Delay:       100,
		Description: "Get all users",
	}

	if endpoint.Path != "/api/users" {
		t.Errorf("Expected path /api/users, got %s", endpoint.Path)
	}

	if endpoint.Method != "GET" {
		t.Errorf("Expected method GET, got %s", endpoint.Method)
	}

	if endpoint.Status != 200 {
		t.Errorf("Expected status 200, got %d", endpoint.Status)
	}

	if endpoint.Delay != 100 {
		t.Errorf("Expected delay 100, got %d", endpoint.Delay)
	}

	if endpoint.Headers["X-Custom"] != "value" {
		t.Errorf("Expected header X-Custom=value, got %s", endpoint.Headers["X-Custom"])
	}
}

func TestGraphQLConfig(t *testing.T) {
	gqlCfg := GraphQLConfig{
		Enabled: true,
		Path:    "/graphql",
		Types: []GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
				Description: "A user type",
			},
		},
		Queries: []GraphQLQuery{
			{
				Name:       "user",
				ReturnType: "User",
				Args: map[string]string{
					"id": "Int!",
				},
				Response:    `{"id": 1, "name": "Test"}`,
				Description: "Get user by ID",
			},
		},
	}

	if !gqlCfg.Enabled {
		t.Error("Expected GraphQL to be enabled")
	}

	if gqlCfg.Path != "/graphql" {
		t.Errorf("Expected path /graphql, got %s", gqlCfg.Path)
	}

	if len(gqlCfg.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(gqlCfg.Types))
	}

	if gqlCfg.Types[0].Name != "User" {
		t.Errorf("Expected type name User, got %s", gqlCfg.Types[0].Name)
	}

	if len(gqlCfg.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(gqlCfg.Queries))
	}
}

func TestConfig_FullStructure(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Port:         9000,
			Host:         "localhost",
			ReadTimeout:  20,
			WriteTimeout: 25,
		},
		Endpoints: []EndpointConfig{
			{
				Path:     "/test",
				Method:   "GET",
				Status:   200,
				Response: "{}",
			},
		},
		GraphQL: &GraphQLConfig{
			Enabled: true,
			Path:    "/gql",
		},
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Server.Port)
	}

	if len(cfg.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(cfg.Endpoints))
	}

	if cfg.GraphQL == nil {
		t.Fatal("Expected GraphQL config, got nil")
	}

	if !cfg.GraphQL.Enabled {
		t.Error("Expected GraphQL to be enabled")
	}
}
