// Package main is the entry point for the Resilient GitOps Platform application.
//
// This application provides a REST API with Prometheus metrics integration
// for monitoring, health checks for Kubernetes probes, and a stress test
// endpoint for demonstrating Horizontal Pod Autoscaler (HPA) behavior.
//
// Configuration:
//   - PORT: HTTP server port (default: 8080)
//   - LOG_LEVEL: Logging verbosity - debug, info, warn, error (default: info)
//
// Endpoints:
//   - GET /         : Main application endpoint with welcome message
//   - GET /health   : Health check endpoint for Kubernetes probes
//   - GET /stress   : CPU stress test endpoint for HPA demonstration
//   - GET /metrics  : Prometheus metrics endpoint
//
// Example:
//
//	LOG_LEVEL=debug PORT=8080 go run ./cmd
package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/moabdelazem/go-gitops-app/internal/handlers"
	"github.com/moabdelazem/go-gitops-app/internal/middleware"
	"github.com/moabdelazem/go-gitops-app/pkg/logger"
	"github.com/moabdelazem/go-gitops-app/pkg/metrics"
)

func main() {
	// Load environment variables from .env file if present.
	// This is a no-op in production where environment variables are set directly.
	// Errors are ignored as the .env file is optional.
	_ = godotenv.Load()

	// Initialize the structured logger first to enable logging throughout startup
	logger.Init()

	// Register Prometheus metrics collectors
	metrics.Register()

	// Create and configure the Gorilla Mux router
	router := setupRouter()

	// Determine the server port from environment or use default
	port := getPort()

	// Start the HTTP server
	startServer(router, port)
}

// setupRouter creates and configures the Gorilla Mux router with all routes
// and middleware. This function centralizes route configuration for clarity.
//
// Routes are organized into two groups:
//   - Application routes: Business logic endpoints with logging middleware
//   - Infrastructure routes: Metrics and health endpoints
func setupRouter() *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware in order:
	// 1. Recovery: Catches panics and prevents server crashes
	// 2. Logging: Logs all requests with structured fields
	router.Use(middleware.Recovery)
	router.Use(middleware.Logging)

	// Register application routes
	// These endpoints serve the main application functionality
	router.HandleFunc("/", handlers.HomeHandler).Methods(http.MethodGet)
	router.HandleFunc("/health", handlers.HealthHandler).Methods(http.MethodGet)
	router.HandleFunc("/stress", handlers.StressHandler).Methods(http.MethodGet)

	// Register infrastructure routes
	// Prometheus metrics endpoint for observability
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	logger.Info().
		Int("route_count", 4).
		Msg("Router configured successfully")

	return router
}

// getPort retrieves the server port from the PORT environment variable.
// If not set, it returns the default port "8080".
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// startServer starts the HTTP server on the specified port.
// It logs startup information and handles fatal errors during server startup.
//
// The server binds to all network interfaces (0.0.0.0) on the specified port.
func startServer(router *mux.Router, port string) {
	logger.Info().
		Str("port", port).
		Str("version", handlers.AppVersion).
		Msg("Starting Resilient GitOps Platform")

	addr := ":" + port

	logger.Info().
		Str("addr", addr).
		Msg("Server listening")

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal().
			Err(err).
			Msg("Server failed to start")
	}
}
