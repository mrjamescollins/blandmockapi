// +build !lambda

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jimbo/blandmockapi/internal/config"
	"github.com/jimbo/blandmockapi/internal/graphql"
	"github.com/jimbo/blandmockapi/internal/router"
)

var (
	configPath = flag.String("config", "./examples", "Path to configuration file or directory")
	lambda     = flag.Bool("lambda", false, "Run in AWS Lambda mode")
)

func main() {
	flag.Parse()

	// Check if running in Lambda mode
	if *lambda || os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		runLambda()
		return
	}

	// Standard server mode
	runServer()
}

func runServer() {
	log.Println("Starting Bland Mock API...")

	// Load configuration
	loader := config.New()
	if err := loader.LoadFromPath(*configPath); err != nil {
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
		log.Printf("GraphQL endpoint enabled with %d types, %d queries, %d mutations",
			len(cfg.GraphQL.Types), len(cfg.GraphQL.Queries), len(cfg.GraphQL.Mutations))
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.GetHost(), cfg.Server.GetPort())
	srv := &http.Server{
		Addr:         addr,
		Handler:      rt.Handler(),
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func runLambda() {
	log.Fatal("Lambda mode requires building with -tags lambda. See README for details.")
}
