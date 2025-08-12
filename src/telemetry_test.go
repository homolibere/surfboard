package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewTelemetryManager tests the creation of a new TelemetryManager
func TestNewTelemetryManager(t *testing.T) {
	// Test with telemetry disabled
	config := TelemetryConfig{
		Enabled: false,
	}

	tm, err := NewTelemetryManager(config)
	if err != nil {
		t.Fatalf("Failed to create TelemetryManager with disabled config: %v", err)
	}

	if tm == nil {
		t.Fatal("TelemetryManager should not be nil even when disabled")
	}

	// Test with telemetry enabled but invalid URL (should fail)
	configInvalid := TelemetryConfig{
		Enabled:       true,
		MetricsURL:    "invalid://url",
		ServiceName:   "test-service",
		ExportTimeout: 1000,
	}

	_, err = NewTelemetryManager(configInvalid)
	if err == nil {
		t.Fatal("Expected error when creating TelemetryManager with invalid URL")
	}
}

// TestTelemetryRecordRequest tests the RecordRequest method
func TestTelemetryRecordRequest(t *testing.T) {
	// Create a TelemetryManager with disabled telemetry (for safety in tests)
	config := TelemetryConfig{
		Enabled: false,
	}

	tm, err := NewTelemetryManager(config)
	if err != nil {
		t.Fatalf("Failed to create TelemetryManager: %v", err)
	}

	// Test that RecordRequest doesn't panic when telemetry is disabled
	tm.RecordRequest(
		context.Background(),
		"/test",
		"GET",
		200,
		100.0,
	)

	// No assertion needed - if it doesn't panic, the test passes
}

// TestTelemetryShutdown tests the Shutdown method
func TestTelemetryShutdown(t *testing.T) {
	// Create a TelemetryManager with disabled telemetry
	config := TelemetryConfig{
		Enabled: false,
	}

	tm, err := NewTelemetryManager(config)
	if err != nil {
		t.Fatalf("Failed to create TelemetryManager: %v", err)
	}

	// Test that Shutdown doesn't panic when telemetry is disabled
	err = tm.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}

// TestTelemetryIntegration tests the integration of telemetry with the gateway
func TestTelemetryIntegration(t *testing.T) {
	// Create a mock backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	}))
	defer backendServer.Close()

	// Create a test configuration with telemetry disabled (for test safety)
	config := Config{
		Endpoints: []Endpoint{
			{
				Path:          "/test-telemetry",
				Method:        "GET",
				Backend:       backendServer.URL,
				Timeout:       1000,
				Headers:       map[string]string{},
				QueryParams:   map[string]string{},
				HasPathParams: false,
			},
		},
		Port: 8080,
		Telemetry: TelemetryConfig{
			Enabled:       false,
			ServiceName:   "test-service",
			MetricsURL:    "http://localhost:4318/v1/metrics",
			ExportTimeout: 1000,
		},
	}

	// Create a telemetry manager
	telemetry, err := NewTelemetryManager(config.Telemetry)
	if err != nil {
		t.Fatalf("Failed to create TelemetryManager: %v", err)
	}

	// Create a new gateway with the telemetry manager
	gateway := NewGateway(config, telemetry)

	// Register endpoints
	gateway.RegisterEndpoints()

	// Create a test request
	req, err := http.NewRequest("GET", "/test-telemetry", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request using the gateway's mux
	gateway.mux.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test health check with telemetry
	gateway.RegisterHealthCheck()

	// Create a test request for health check
	reqHealth, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create health request: %v", err)
	}

	// Create a response recorder
	rrHealth := httptest.NewRecorder()

	// Serve the request using the gateway's mux
	gateway.mux.ServeHTTP(rrHealth, reqHealth)

	// Check the response status code
	if status := rrHealth.Code; status != http.StatusOK {
		t.Errorf("health handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Shutdown telemetry
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = telemetry.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown telemetry: %v", err)
	}
}

// TestTelemetryWithMockMetrics tests the telemetry with mock metrics
func TestTelemetryWithMockMetrics(t *testing.T) {
	// This test would ideally use a mock meter provider to verify metrics are recorded
	// However, OpenTelemetry doesn't provide an easy way to mock metrics in tests
	// So we'll just test that the code doesn't panic when recording metrics

	// Create a TelemetryManager with disabled telemetry
	config := TelemetryConfig{
		Enabled: false,
	}

	tm, err := NewTelemetryManager(config)
	if err != nil {
		t.Fatalf("Failed to create TelemetryManager: %v", err)
	}

	// Record metrics for different status codes
	ctx := context.Background()

	// Success case
	tm.RecordRequest(ctx, "/test", "GET", 200, 100.0)

	// Client error case
	tm.RecordRequest(ctx, "/test", "GET", 404, 50.0)

	// Server error case
	tm.RecordRequest(ctx, "/test", "GET", 500, 200.0)

	// No assertion needed - if it doesn't panic, the test passes
}
