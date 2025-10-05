// +build lambda

package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/jimbo/blandmockapi/internal/config"
	"github.com/jimbo/blandmockapi/internal/graphql"
	"github.com/jimbo/blandmockapi/internal/router"
)

func main() {
	runLambda()
}

func runLambda() {
	log.Println("Initializing Lambda handler...")

	// Get config path from environment or use default
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config"
	}

	// Load configuration
	loader := config.New()
	if err := loader.LoadFromPath(configPath); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := loader.GetConfig()
	log.Printf("Loaded configuration with %d endpoints", len(cfg.Endpoints))

	// Create router
	rt := router.New()

	// Register health check
	rt.RegisterHealthCheck()

	// Register REST endpoints
	if err := rt.RegisterEndpoints(cfg.Endpoints); err != nil {
		log.Fatalf("Failed to register endpoints: %v", err)
	}

	// Register GraphQL endpoint if enabled
	if cfg.GraphQL != nil && cfg.GraphQL.Enabled {
		gqlHandler, err := graphql.New(cfg.GraphQL)
		if err != nil {
			log.Fatalf("Failed to create GraphQL handler: %v", err)
		}

		path := cfg.GraphQL.Path
		if path == "" {
			path = "/graphql"
		}
		rt.RegisterGraphQL(path, gqlHandler.ServeHTTP)
		log.Printf("GraphQL endpoint enabled")
	}

	// Create Lambda handler using httpadapter
	log.Println("Starting Lambda handler...")
	lambda.Start(httpadapter.New(rt.Handler()).ProxyWithContext)
}
