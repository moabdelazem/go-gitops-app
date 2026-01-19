// Package metrics provides Prometheus instrumentation for the application.
//
// This package defines and registers Prometheus metrics for monitoring
// HTTP requests and other application-specific telemetry. It provides
// a clean interface for recording metrics throughout the application.
//
// Example usage:
//
//	metrics.Register()
//	metrics.TrackRequest("/api/users", "GET")
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// httpRequestsTotal tracks the total number of HTTP requests processed,
// labeled by path and method. This counter is essential for monitoring
// request volume and traffic patterns.
var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed",
	},
	[]string{"path", "method"},
)

// httpRequestDuration tracks the duration of HTTP requests in seconds,
// labeled by path and method. This histogram helps identify slow endpoints
// and monitor latency distribution.
var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"path", "method"},
)

// Register registers all application metrics with the default Prometheus registry.
// This function should be called once during application startup, typically
// in the main function before starting the HTTP server.
//
// Panics if metrics are already registered (duplicate registration).
func Register() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// TrackRequest increments the request counter for the specified path and method.
// This function should be called for each incoming HTTP request.
//
// Parameters:
//   - path: The request URL path (e.g., "/api/users").
//   - method: The HTTP method (e.g., "GET", "POST").
func TrackRequest(path, method string) {
	httpRequestsTotal.WithLabelValues(path, method).Inc()
}

// ObserveRequestDuration records the duration of an HTTP request.
// This function should be called after the request has been processed.
//
// Parameters:
//   - path: The request URL path.
//   - method: The HTTP method.
//   - durationSeconds: The request processing time in seconds.
func ObserveRequestDuration(path, method string, durationSeconds float64) {
	httpRequestDuration.WithLabelValues(path, method).Observe(durationSeconds)
}
