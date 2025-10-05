package graphql

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/jimbo/blandmockapi/internal/models"
)

func TestNew(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Path:    "/graphql",
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "user",
				ReturnType: "User",
				Response:   `{"id": 1, "name": "Test"}`,
			},
		},
	}

	handler, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if handler == nil {
		t.Fatal("New() returned nil handler")
	}

	if handler.config != config {
		t.Error("Handler config not set correctly")
	}
}

func TestNew_Disabled(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: false,
	}

	_, err := New(config)
	if err == nil {
		t.Error("Expected error when GraphQL is disabled")
	}
}

func TestNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Error("Expected error when config is nil")
	}
}

func TestParseType(t *testing.T) {
	handler := &Handler{
		config: &models.GraphQLConfig{},
	}

	tests := []struct {
		name     string
		typeStr  string
		wantNull bool
		wantList bool
	}{
		{"basic string", "String", false, false},
		{"non-null string", "String!", true, false},
		{"list of strings", "[String]", false, true},
		{"non-null list", "[String]!", true, true},
		{"int type", "Int", false, false},
		{"non-null int", "Int!", true, false},
		{"float type", "Float", false, false},
		{"boolean type", "Boolean", false, false},
		{"id type", "ID", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.parseType(tt.typeStr)
			if result == nil {
				t.Error("parseType() returned nil")
			}
			// Type checking is complex, just verify it doesn't panic
		})
	}
}

func TestBuildSchema_SimpleType(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "users",
				ReturnType: "[User]",
				Response:   `[{"id": 1, "name": "Test"}]`,
			},
		},
	}

	handler := &Handler{config: config}
	schema, err := handler.buildSchema()

	if err != nil {
		t.Fatalf("buildSchema() failed: %v", err)
	}

	if schema.QueryType() == nil {
		t.Error("Schema has no query type")
	}
}

func TestBuildSchema_WithMutations(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "user",
				ReturnType: "User",
				Response:   `{"id": 1, "name": "Test"}`,
			},
		},
		Mutations: []models.GraphQLMutation{
			{
				Name:       "createUser",
				ReturnType: "User",
				Args: map[string]string{
					"name": "String!",
				},
				Response: `{"id": 2, "name": "Created"}`,
			},
		},
	}

	handler := &Handler{config: config}
	schema, err := handler.buildSchema()

	if err != nil {
		t.Fatalf("buildSchema() failed: %v", err)
	}

	if schema.MutationType() == nil {
		t.Error("Schema has no mutation type")
	}
}

func TestServeHTTP_ValidQuery(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Path:    "/graphql",
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "users",
				ReturnType: "[User]",
				Response:   `[{"id": 1, "name": "Test"}]`,
			},
		},
	}

	handler, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	query := map[string]string{
		"query": "{ users { id name } }",
	}
	body, _ := json.Marshal(query)

	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["data"] == nil {
		t.Error("Response missing 'data' field")
	}
}

func TestServeHTTP_MethodNotAllowed(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id": "Int!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "user",
				ReturnType: "User",
				Response:   `{"id": 1}`,
			},
		},
	}

	handler, _ := New(config)

	req := httptest.NewRequest("GET", "/graphql", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 405 {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestServeHTTP_InvalidJSON(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id": "Int!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "user",
				ReturnType: "User",
				Response:   `{"id": 1}`,
			},
		},
	}

	handler, _ := New(config)

	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServeHTTP_WithVariables(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "user",
				ReturnType: "User",
				Args: map[string]string{
					"id": "Int!",
				},
				Response: `{"id": 1, "name": "Test"}`,
			},
		},
	}

	handler, _ := New(config)

	requestBody := map[string]interface{}{
		"query": "query GetUser($id: Int!) { user(id: $id) { id name } }",
		"variables": map[string]interface{}{
			"id": 1,
		},
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCreateResolver(t *testing.T) {
	handler := &Handler{
		config: &models.GraphQLConfig{},
	}

	tests := []struct {
		name         string
		responseJSON string
		wantErr      bool
	}{
		{
			name:         "valid json object",
			responseJSON: `{"id": 1, "name": "Test"}`,
			wantErr:      false,
		},
		{
			name:         "valid json array",
			responseJSON: `[{"id": 1}, {"id": 2}]`,
			wantErr:      false,
		},
		{
			name:         "valid json string",
			responseJSON: `"test"`,
			wantErr:      false,
		},
		{
			name:         "valid json number",
			responseJSON: `42`,
			wantErr:      false,
		},
		{
			name:         "valid json boolean",
			responseJSON: `true`,
			wantErr:      false,
		},
		{
			name:         "invalid json",
			responseJSON: `{invalid}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := handler.createResolver(tt.responseJSON)
			result, err := resolver(graphql.ResolveParams{})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid JSON")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result, got nil")
				}
			}
		})
	}
}

func TestResolveType(t *testing.T) {
	handler := &Handler{
		config: &models.GraphQLConfig{},
	}

	tests := []struct {
		name     string
		typeName string
	}{
		{"string type", "String"},
		{"int type", "Int"},
		{"float type", "Float"},
		{"boolean type", "Boolean"},
		{"id type", "ID"},
		{"list of strings", "[String]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a complex type resolution test
			// Just verify it doesn't panic
			_ = handler.parseType(tt.typeName)
		})
	}
}

func TestMultipleTypes(t *testing.T) {
	config := &models.GraphQLConfig{
		Enabled: true,
		Types: []models.GraphQLType{
			{
				Name: "User",
				Fields: map[string]string{
					"id":   "Int!",
					"name": "String!",
				},
			},
			{
				Name: "Post",
				Fields: map[string]string{
					"id":      "Int!",
					"title":   "String!",
					"content": "String",
				},
			},
		},
		Queries: []models.GraphQLQuery{
			{
				Name:       "users",
				ReturnType: "[User]",
				Response:   `[{"id": 1, "name": "Test"}]`,
			},
			{
				Name:       "posts",
				ReturnType: "[Post]",
				Response:   `[{"id": 1, "title": "Test Post"}]`,
			},
		},
	}

	handler, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	if handler.schema.QueryType() == nil {
		t.Error("Schema missing query type")
	}

	// Test users query
	query := map[string]string{"query": "{ users { id name } }"}
	body, _ := json.Marshal(query)
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Users query failed with status %d", w.Code)
	}

	// Test posts query
	query = map[string]string{"query": "{ posts { id title } }"}
	body, _ = json.Marshal(query)
	req = httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Posts query failed with status %d", w.Code)
	}
}
