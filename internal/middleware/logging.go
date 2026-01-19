// Package middleware provides HTTP middleware components for the application.
//
// Middleware functions wrap HTTP handlers to provide cross-cutting concerns
// such as logging, authentication, rate limiting, and request tracking.
// These middleware components are designed to work with Gorilla Mux router.
//
// Example usage:
//
//	router := mux.NewRouter()
//	router.Use(middleware.Logging)
package middleware

import (
	"net/http"
	"time"

	"github.com/moabdelazem/go-gitops-app/pkg/logger"
	"github.com/moabdelazem/go-gitops-app/pkg/metrics"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
// This is necessary because the standard ResponseWriter doesn't expose
// the status code after WriteHeader is called.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new responseWriter with a default status of 200.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code before delegating to the underlying writer.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging is a middleware that logs HTTP requests with structured fields.
// It captures the request method, path, status code, and duration.
//
// The middleware logs at different levels based on the HTTP status code:
//   - 2xx, 3xx: Info level
//   - 4xx: Warn level (client errors)
//   - 5xx: Error level (server errors)
//
// This middleware also records request duration in Prometheus metrics.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		wrapped := newResponseWriter(w)

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Calculate request duration
		duration := time.Since(start)
		durationSeconds := duration.Seconds()

		// Record duration in metrics
		metrics.ObserveRequestDuration(r.URL.Path, r.Method, durationSeconds)

		// Build the log event with common fields
		logEvent := logger.Info()
		if wrapped.statusCode >= 400 && wrapped.statusCode < 500 {
			logEvent = logger.Warn()
		} else if wrapped.statusCode >= 500 {
			logEvent = logger.Error()
		}

		// Log the request with structured fields
		logEvent.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("HTTP request completed")
	})
}

// Recovery is a middleware that recovers from panics and returns a 500 error.
// This prevents the server from crashing due to unhandled panics in handlers.
//
// When a panic is recovered, the error is logged with stack trace information
// and a generic error response is sent to the client.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error().
					Interface("panic", err).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("Recovered from panic")

				http.Error(w, `{"status":"error","message":"Internal server error"}`, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
