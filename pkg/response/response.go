// Package response provides standardized HTTP response types and utilities.
//
// This package defines a consistent response structure for API endpoints,
// ensuring all responses follow a predictable JSON format. It includes
// helper functions for encoding and sending responses.
//
// Example usage:
//
//	resp := response.New("success", "Operation completed", "v1.0.0")
//	response.SendJSON(w, http.StatusOK, resp)
package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standardized API response structure.
// All API endpoints should return responses in this format to ensure
// consistency across the application.
type Response struct {
	// Status indicates the outcome of the request (e.g., "success", "error").
	Status string `json:"status"`

	// Message provides human-readable information about the response.
	Message string `json:"message"`

	// Version indicates the API version that generated this response.
	// This field is optional and may be empty for certain responses.
	Version string `json:"version,omitempty"`
}

// New creates a new Response with the specified status, message, and version.
// This is the preferred constructor for creating Response instances.
func New(status, message, version string) Response {
	return Response{
		Status:  status,
		Message: message,
		Version: version,
	}
}

// Error creates a new error Response with the specified message.
// This is a convenience constructor for error responses.
func Error(message string) Response {
	return Response{
		Status:  "error",
		Message: message,
	}
}

// Success creates a new success Response with the specified message.
// This is a convenience constructor for success responses.
func Success(message string) Response {
	return Response{
		Status:  "success",
		Message: message,
	}
}

// SendJSON encodes the provided data as JSON and writes it to the ResponseWriter.
// It sets the appropriate Content-Type header and HTTP status code.
//
// Parameters:
//   - w: The HTTP ResponseWriter to write the response to.
//   - statusCode: The HTTP status code to set (e.g., http.StatusOK).
//   - data: The data to encode as JSON. This can be any type that is JSON-serializable.
//
// If JSON encoding fails, a 500 Internal Server Error is returned.
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error - this is a critical failure
		http.Error(w, `{"status":"error","message":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}
