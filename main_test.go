package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test LoggingResponseWriter
func TestLoggingResponseWriter(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "Status OK",
			statusCode: http.StatusOK,
			body:       "Test response body",
		},
		{
			name:       "Status Not Found",
			statusCode: http.StatusNotFound,
			body:       "Not Found",
		},
		{
			name:       "Status Internal Server Error",
			statusCode: http.StatusInternalServerError,
			body:       "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			rr := httptest.NewRecorder()

			// Create a logging response writer
			lrw := NewLoggingResponseWriter(rr)

			// Set the status code
			lrw.WriteHeader(tt.statusCode)

			// Write the body
			_, _ = lrw.Write([]byte(tt.body))

			// Check if the status code was captured correctly
			if lrw.statusCode != tt.statusCode {
				t.Errorf("LoggingResponseWriter.statusCode = %v, want %v", lrw.statusCode, tt.statusCode)
			}

			// Check if the status code was written to the underlying ResponseWriter
			if rr.Code != tt.statusCode {
				t.Errorf("ResponseRecorder.Code = %v, want %v", rr.Code, tt.statusCode)
			}

			// Check if the body was captured correctly
			if lrw.GetBody() != tt.body {
				t.Errorf("LoggingResponseWriter.GetBody() = %v, want %v", lrw.GetBody(), tt.body)
			}

			// Check if the body was written to the underlying ResponseWriter
			if rr.Body.String() != tt.body {
				t.Errorf("ResponseRecorder.Body.String() = %v, want %v", rr.Body.String(), tt.body)
			}
		})
	}
}

// Test health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a gateway with empty config
	gateway := NewGateway(Config{}, nil)

	// Register the health check endpoint
	gateway.RegisterHealthCheck()

	// Serve the request using the gateway's mux
	gateway.mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("handler returned unexpected body: got %v want %v", response["status"], "ok")
	}
}
