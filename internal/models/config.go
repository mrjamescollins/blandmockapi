package models

import "time"

// Config represents the entire application configuration
type Config struct {
	Server    ServerConfig      `toml:"server"`
	Endpoints []EndpointConfig  `toml:"endpoints"`
	GraphQL   *GraphQLConfig    `toml:"graphql"`
}

// ServerConfig contains server-level settings
type ServerConfig struct {
	Port         int    `toml:"port"`
	Host         string `toml:"host"`
	ReadTimeout  int    `toml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout"`
}

// EndpointConfig defines a REST endpoint
type EndpointConfig struct {
	Path        string            `toml:"path"`
	Method      string            `toml:"method"`
	Status      int               `toml:"status"`
	Response    string            `toml:"response"`
	Headers     map[string]string `toml:"headers"`
	Delay       int               `toml:"delay"` // milliseconds
	Description string            `toml:"description"`
}

// GraphQLConfig defines GraphQL endpoint configuration
type GraphQLConfig struct {
	Enabled bool                `toml:"enabled"`
	Path    string              `toml:"path"`
	Types   []GraphQLType       `toml:"types"`
	Queries []GraphQLQuery      `toml:"queries"`
	Mutations []GraphQLMutation `toml:"mutations"`
}

// GraphQLType represents a GraphQL type definition
type GraphQLType struct {
	Name        string              `toml:"name"`
	Fields      map[string]string   `toml:"fields"`
	Description string              `toml:"description"`
}

// GraphQLQuery represents a GraphQL query
type GraphQLQuery struct {
	Name        string            `toml:"name"`
	ReturnType  string            `toml:"return_type"`
	Args        map[string]string `toml:"args"`
	Response    string            `toml:"response"`
	Description string            `toml:"description"`
}

// GraphQLMutation represents a GraphQL mutation
type GraphQLMutation struct {
	Name        string            `toml:"name"`
	ReturnType  string            `toml:"return_type"`
	Args        map[string]string `toml:"args"`
	Response    string            `toml:"response"`
	Description string            `toml:"description"`
}

// GetReadTimeout returns the read timeout as a duration
func (s *ServerConfig) GetReadTimeout() time.Duration {
	if s.ReadTimeout <= 0 {
		return 15 * time.Second
	}
	return time.Duration(s.ReadTimeout) * time.Second
}

// GetWriteTimeout returns the write timeout as a duration
func (s *ServerConfig) GetWriteTimeout() time.Duration {
	if s.WriteTimeout <= 0 {
		return 15 * time.Second
	}
	return time.Duration(s.WriteTimeout) * time.Second
}

// GetPort returns the server port with a default
func (s *ServerConfig) GetPort() int {
	if s.Port <= 0 {
		return 8080
	}
	return s.Port
}

// GetHost returns the server host with a default
func (s *ServerConfig) GetHost() string {
	if s.Host == "" {
		return "0.0.0.0"
	}
	return s.Host
}
