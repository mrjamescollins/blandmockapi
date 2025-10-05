package graphql

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/jimbo/blandmockapi/internal/models"
)

// Handler manages GraphQL requests based on TOML configuration
type Handler struct {
	schema graphql.Schema
	config *models.GraphQLConfig
}

// New creates a new GraphQL handler from configuration
func New(config *models.GraphQLConfig) (*Handler, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("GraphQL is not enabled")
	}

	h := &Handler{
		config: config,
	}

	// Build the GraphQL schema from configuration
	schema, err := h.buildSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to build GraphQL schema: %w", err)
	}

	h.schema = schema
	return h, nil
}

// buildSchema constructs a GraphQL schema from TOML configuration
func (h *Handler) buildSchema() (graphql.Schema, error) {
	// Create custom types
	types := make(map[string]*graphql.Object)
	for _, typeDef := range h.config.Types {
		fields := graphql.Fields{}
		for fieldName, fieldType := range typeDef.Fields {
			fields[fieldName] = &graphql.Field{
				Type:        h.parseType(fieldType),
				Description: fmt.Sprintf("Field %s of type %s", fieldName, fieldType),
			}
		}

		types[typeDef.Name] = graphql.NewObject(graphql.ObjectConfig{
			Name:        typeDef.Name,
			Description: typeDef.Description,
			Fields:      fields,
		})
	}

	// Build query fields
	queryFields := graphql.Fields{}
	for _, query := range h.config.Queries {
		returnType := h.resolveType(query.ReturnType, types)
		if returnType == nil {
			log.Printf("Warning: Unknown return type '%s' for query '%s', using String", query.ReturnType, query.Name)
			returnType = graphql.String
		}

		args := graphql.FieldConfigArgument{}
		for argName, argType := range query.Args {
			args[argName] = &graphql.ArgumentConfig{
				Type: h.parseType(argType),
			}
		}

		queryFields[query.Name] = &graphql.Field{
			Type:        returnType,
			Description: query.Description,
			Args:        args,
			Resolve:     h.createResolver(query.Response),
		}
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: queryFields,
	})

	// Build mutation fields
	var rootMutation *graphql.Object
	if len(h.config.Mutations) > 0 {
		mutationFields := graphql.Fields{}
		for _, mutation := range h.config.Mutations {
			returnType := h.resolveType(mutation.ReturnType, types)
			if returnType == nil {
				log.Printf("Warning: Unknown return type '%s' for mutation '%s', using String", mutation.ReturnType, mutation.Name)
				returnType = graphql.String
			}

			args := graphql.FieldConfigArgument{}
			for argName, argType := range mutation.Args {
				args[argName] = &graphql.ArgumentConfig{
					Type: h.parseType(argType),
				}
			}

			mutationFields[mutation.Name] = &graphql.Field{
				Type:        returnType,
				Description: mutation.Description,
				Args:        args,
				Resolve:     h.createResolver(mutation.Response),
			}
		}

		rootMutation = graphql.NewObject(graphql.ObjectConfig{
			Name:   "RootMutation",
			Fields: mutationFields,
		})
	}

	// Create schema
	schemaConfig := graphql.SchemaConfig{
		Query: rootQuery,
	}
	if rootMutation != nil {
		schemaConfig.Mutation = rootMutation
	}

	return graphql.NewSchema(schemaConfig)
}

// parseType converts a string type to a GraphQL type
func (h *Handler) parseType(typeStr string) graphql.Output {
	// Handle non-null types
	isNonNull := false
	if len(typeStr) > 0 && typeStr[len(typeStr)-1] == '!' {
		isNonNull = true
		typeStr = typeStr[:len(typeStr)-1]
	}

	// Handle list types
	isList := false
	if len(typeStr) > 2 && typeStr[0] == '[' && typeStr[len(typeStr)-1] == ']' {
		isList = true
		typeStr = typeStr[1 : len(typeStr)-1]
	}

	var baseType graphql.Output

	// Map to GraphQL scalar types
	switch typeStr {
	case "String":
		baseType = graphql.String
	case "Int":
		baseType = graphql.Int
	case "Float":
		baseType = graphql.Float
	case "Boolean":
		baseType = graphql.Boolean
	case "ID":
		baseType = graphql.ID
	default:
		// Assume it's a custom type (will be resolved later)
		baseType = graphql.String
	}

	if isList {
		baseType = graphql.NewList(baseType)
	}

	if isNonNull {
		return graphql.NewNonNull(baseType)
	}

	return baseType
}

// resolveType resolves a type name to a GraphQL type (including custom types)
func (h *Handler) resolveType(typeName string, types map[string]*graphql.Object) graphql.Output {
	// Check for custom types first
	if customType, ok := types[typeName]; ok {
		return customType
	}

	// Check for list types
	if len(typeName) > 2 && typeName[0] == '[' && typeName[len(typeName)-1] == ']' {
		innerType := h.resolveType(typeName[1:len(typeName)-1], types)
		if innerType != nil {
			return graphql.NewList(innerType)
		}
	}

	// Fall back to parsing as a basic type
	return h.parseType(typeName)
}

// createResolver creates a resolver function that returns the configured response
func (h *Handler) createResolver(responseJSON string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Parse the JSON response
		var result interface{}
		if err := json.Unmarshal([]byte(responseJSON), &result); err != nil {
			return nil, fmt.Errorf("invalid response JSON: %w", err)
		}
		return result, nil
	}
}

// ServeHTTP handles GraphQL HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"error": "GraphQL endpoint only accepts POST requests",
		}); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	// Parse the request body
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("invalid request body: %v", err),
		}); encErr != nil {
			log.Printf("Failed to encode error response: %v", encErr)
		}
		return
	}

	// Execute the GraphQL query
	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  params.Query,
		VariableValues: params.Variables,
		OperationName:  params.OperationName,
	})

	// Log any errors
	if len(result.Errors) > 0 {
		log.Printf("GraphQL errors: %v", result.Errors)
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode GraphQL response: %v", err)
	}
}
