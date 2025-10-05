package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/jimbo/blandmockapi/internal/models"
)

// Loader handles loading and merging configuration files
type Loader struct {
	config models.Config
}

// New creates a new configuration loader
func New() *Loader {
	return &Loader{
		config: models.Config{
			Server: models.ServerConfig{
				Port:         8080,
				Host:         "0.0.0.0",
				ReadTimeout:  15,
				WriteTimeout: 15,
			},
			Endpoints: []models.EndpointConfig{},
		},
	}
}

// LoadFile loads a single TOML configuration file
func (l *Loader) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg models.Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Merge the loaded config into the main config
	l.mergeConfig(cfg)
	return nil
}

// LoadDirectory loads all .toml files from a directory
func (l *Loader) LoadDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".toml" {
			path := filepath.Join(dir, entry.Name())
			if err := l.LoadFile(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadFromPath loads configuration from a file or directory
func (l *Loader) LoadFromPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if info.IsDir() {
		return l.LoadDirectory(path)
	}

	return l.LoadFile(path)
}

// mergeConfig merges a loaded config into the main config
func (l *Loader) mergeConfig(cfg models.Config) {
	// Override server config if provided
	if cfg.Server.Port > 0 {
		l.config.Server.Port = cfg.Server.Port
	}
	if cfg.Server.Host != "" {
		l.config.Server.Host = cfg.Server.Host
	}
	if cfg.Server.ReadTimeout > 0 {
		l.config.Server.ReadTimeout = cfg.Server.ReadTimeout
	}
	if cfg.Server.WriteTimeout > 0 {
		l.config.Server.WriteTimeout = cfg.Server.WriteTimeout
	}

	// Append endpoints
	l.config.Endpoints = append(l.config.Endpoints, cfg.Endpoints...)

	// Override GraphQL config if provided
	if cfg.GraphQL != nil {
		if l.config.GraphQL == nil {
			l.config.GraphQL = cfg.GraphQL
		} else {
			// Merge GraphQL configuration
			if cfg.GraphQL.Enabled {
				l.config.GraphQL.Enabled = true
			}
			if cfg.GraphQL.Path != "" {
				l.config.GraphQL.Path = cfg.GraphQL.Path
			}
			l.config.GraphQL.Types = append(l.config.GraphQL.Types, cfg.GraphQL.Types...)
			l.config.GraphQL.Queries = append(l.config.GraphQL.Queries, cfg.GraphQL.Queries...)
			l.config.GraphQL.Mutations = append(l.config.GraphQL.Mutations, cfg.GraphQL.Mutations...)
		}
	}
}

// GetConfig returns the loaded configuration
func (l *Loader) GetConfig() models.Config {
	return l.config
}
