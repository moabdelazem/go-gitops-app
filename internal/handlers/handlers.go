// Package handlers provides HTTP request handlers for the GitOps application.
//
// This package contains all the HTTP handler functions that process incoming
// requests. Each handler is responsible for a specific endpoint and follows
// a consistent pattern of logging, metric tracking, and response formatting.
//
// The handlers in this package are designed to work with Gorilla Mux router
// and utilize structured logging for observability.
package handlers

import (
	"math"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/moabdelazem/go-gitops-app/pkg/logger"
	"github.com/moabdelazem/go-gitops-app/pkg/metrics"
	"github.com/moabdelazem/go-gitops-app/pkg/response"
)

const (
	// AppVersion represents the current version of the application.
	// This is included in API responses for client version tracking.
	AppVersion = "v1.0.0"

	// Default stress test configuration
	defaultStressDuration = 2 * time.Second
	maxStressDuration     = 30 * time.Second
)

// validate is the singleton validator instance used across all handlers.
var validate = validator.New()

// StressRequest represents the validated parameters for a stress test.
// Validation tags ensure all values are within acceptable bounds.
type StressRequest struct {
	// DurationSeconds is the stress test duration in seconds (1-30).
	DurationSeconds int `validate:"min=1,max=30"`

	// Workers is the number of concurrent CPU workers (1 to 2x CPU cores).
	Workers int `validate:"min=1"`
}

// StressResponse represents the response from a stress test.
type StressResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duration string `json:"duration"`
	Workers  int    `json:"workers"`
}

// HomeHandler handles requests to the root endpoint.
// It returns a welcome message along with the current application version.
//
// Endpoint: GET /
// Response: JSON with status, message, and version fields.
//
// This handler tracks the request in Prometheus metrics and logs the
// request at debug level.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Track request metrics for monitoring
	metrics.TrackRequest(r.URL.Path, r.Method)

	logger.Debug().
		Str("path", r.URL.Path).
		Str("method", r.Method).
		Str("remote_addr", r.RemoteAddr).
		Msg("Processing home request")

	resp := response.New(
		"success",
		"Welcome to the Resilient GitOps Platform!",
		AppVersion,
	)

	response.SendJSON(w, http.StatusOK, resp)
}

// HealthHandler handles health check requests for Kubernetes probes.
// It returns a simple "OK" response to indicate the service is healthy.
//
// Endpoint: GET /health
// Response: Plain text "OK" with 200 status code.
//
// This handler intentionally does not track metrics to avoid noise from
// frequent health check requests by Kubernetes liveness/readiness probes.
//
// Usage:
//   - Kubernetes Liveness Probe: Ensures the container is running
//   - Kubernetes Readiness Probe: Ensures the service is ready to accept traffic
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Health checks are not logged at info level to reduce noise
	logger.Debug().
		Str("path", r.URL.Path).
		Msg("Health check requested")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// StressHandler simulates high CPU load to trigger Horizontal Pod Autoscaler (HPA).
// It spawns multiple worker goroutines to stress multiple CPU cores simultaneously,
// allowing effective testing of auto-scaling behavior in multi-core environments.
//
// Endpoint: GET /stress
//
// Query Parameters:
//   - duration: How long to run the stress test (e.g., "5s", "10s"). Default: 2s, Max: 30s
//   - workers: Number of concurrent CPU workers. Default: number of CPU cores
//
// Examples:
//   - GET /stress                     (2s duration, all cores)
//   - GET /stress?duration=5s         (5s duration, all cores)
//   - GET /stress?workers=2           (2s duration, 2 cores)
//   - GET /stress?duration=10s&workers=4
//
// Response: JSON with status, duration, and worker count.
//
// ! WARNING: This endpoint is intended for testing purposes and the nature of this experimental api
// ! Real applications have something like this
func StressHandler(w http.ResponseWriter, r *http.Request) {
	// Track request metrics for monitoring
	metrics.TrackRequest(r.URL.Path, r.Method)

	// Parse and validate request parameters
	req, err := parseAndValidateStressRequest(r)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("path", r.URL.Path).
			Msg("Invalid stress request parameters")

		response.SendJSON(w, http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	duration := time.Duration(req.DurationSeconds) * time.Second

	logger.Warn().
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Dur("duration", duration).
		Int("workers", req.Workers).
		Msg("Multi-core stress test initiated - CPU spike incoming")

	// Execute stress test across multiple goroutines
	start := time.Now()
	runMultiCoreStress(duration, req.Workers)
	elapsed := time.Since(start)

	logger.Info().
		Dur("duration", elapsed).
		Int("workers", req.Workers).
		Msg("Stress test completed")

	// Build detailed response
	resp := StressResponse{
		Status:   "stress_complete",
		Message:  "CPU load simulation finished",
		Duration: elapsed.String(),
		Workers:  req.Workers,
	}

	response.SendJSON(w, http.StatusOK, resp)
}

// parseAndValidateStressRequest extracts and validates stress test parameters
// from the HTTP request query string using go-playground/validator.
//
// Returns a validated StressRequest or an error if validation fails.
func parseAndValidateStressRequest(r *http.Request) (*StressRequest, error) {
	numCPU := runtime.NumCPU()
	maxWorkers := numCPU * 2

	// Parse duration (default: 2s)
	durationSeconds := int(defaultStressDuration.Seconds())
	if durationStr := r.URL.Query().Get("duration"); durationStr != "" {
		if d, err := time.ParseDuration(durationStr); err == nil {
			durationSeconds = int(d.Seconds())
		}
	}

	// Parse workers (default: number of CPU cores)
	workers := numCPU
	if workersStr := r.URL.Query().Get("workers"); workersStr != "" {
		if w, err := strconv.Atoi(workersStr); err == nil {
			workers = w
		}
	}

	// Create request struct for validation
	req := &StressRequest{
		DurationSeconds: durationSeconds,
		Workers:         workers,
	}

	// Validate using struct tags
	if err := validate.Struct(req); err != nil {
		// Translate validation errors to user-friendly messages
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return nil, formatValidationError(validationErrors, maxWorkers)
		}
		return nil, err
	}

	// Apply max bounds (validator doesn't support dynamic max)
	if req.DurationSeconds > int(maxStressDuration.Seconds()) {
		req.DurationSeconds = int(maxStressDuration.Seconds())
	}
	if req.Workers > maxWorkers {
		req.Workers = maxWorkers
	}

	return req, nil
}

// formatValidationError converts validator.ValidationErrors to a user-friendly error.
func formatValidationError(errors validator.ValidationErrors, maxWorkers int) error {
	for _, err := range errors {
		switch err.Field() {
		case "DurationSeconds":
			return &ValidationError{
				Field:   "duration",
				Message: "duration must be between 1s and 30s",
			}
		case "Workers":
			return &ValidationError{
				Field:   "workers",
				Message: "workers must be between 1 and " + strconv.Itoa(maxWorkers),
			}
		}
	}
	return &ValidationError{Field: "unknown", Message: "validation failed"}
}

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// runMultiCoreStress executes CPU-intensive work across multiple goroutines.
// Each worker performs continuous math operations to consume CPU cycles.
func runMultiCoreStress(duration time.Duration, workers int) {
	var wg sync.WaitGroup

	// Launch worker goroutines
	for i := range workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			stressWorker(duration, workerID)
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
}

// stressWorker performs CPU-intensive calculations for the specified duration.
// It runs a tight loop of math operations to maximize CPU utilization.
// The workerID is used for logging to identify individual workers.
func stressWorker(duration time.Duration, workerID int) {
	logger.Debug().
		Int("worker_id", workerID).
		Dur("target_duration", duration).
		Msg("Stress worker started")

	start := time.Now()

	// Use multiple math operations to maximize CPU usage
	var result float64
	for time.Since(start) < duration {
		// Mix of operations to prevent compiler optimization
		result = math.Sqrt(float64(time.Now().UnixNano()))
		result = math.Sin(result) * math.Cos(result)
		result = math.Log(math.Abs(result) + 1)
		_ = result
	}

	elapsed := time.Since(start)
	logger.Debug().
		Int("worker_id", workerID).
		Dur("elapsed", elapsed).
		Msg("Stress worker finished")
}
