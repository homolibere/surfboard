package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProxyHandlerDirectly tests the Handler method of the Proxy class directly
func TestProxyHandlerDirectly(t *testing.T) {
	// Create a mock backend server
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that headers were forwarded correctly
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected X-Test-Header to be 'test-value', got '%s'", r.Header.Get("X-Test-Header"))
		}

		// Verify that query parameters were forwarded correctly
		if r.URL.Query().Get("param1") != "value1" {
			t.Errorf("Expected query param 'param1' to be 'value1', got '%s'", r.URL.Query().Get("param1"))
		}

		// Send a response
		_, err := fmt.Fprintln(w, "Hello from mock backend")
		if err != nil {
			t.Fatalf("Failed to create mock backend: %v", err)
		}
	}))
	defer mockBackend.Close()

	// Create a test endpoint with the mock backend URL
	endpoint := Endpoint{
		Path:          "/test",
		Method:        "GET",
		Backend:       mockBackend.URL,
		Timeout:       1000,
		Headers:       map[string]string{"X-Test-Header": "test-value"},
		QueryParams:   map[string]string{"param1": "value1"},
		HasPathParams: false,
	}

	// Create a new proxy
	proxy := NewProxy(endpoint, false, nil)

	// Get the handler
	handler := proxy.Handler()

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expectedBody := "Hello from mock backend\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

// TestProxyHandlerInvalidMethod tests the Handler method with an invalid HTTP method
func TestProxyHandlerInvalidMethod(t *testing.T) {
	// Create a test endpoint that only accepts GET requests
	endpoint := Endpoint{
		Path:          "/test",
		Method:        "GET",
		Backend:       "https://example.com",
		Timeout:       1000,
		Headers:       map[string]string{},
		QueryParams:   map[string]string{},
		HasPathParams: false,
	}

	// Create a new proxy
	proxy := NewProxy(endpoint, false, nil)

	// Get the handler
	handler := proxy.Handler()

	// Create a test request with POST method (should be rejected)
	req, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

// TestProxyHandlerInvalidBackendURL tests the Handler method with an invalid backend URL
func TestProxyHandlerInvalidBackendURL(t *testing.T) {
	// Create a test endpoint with an invalid backend URL
	endpoint := Endpoint{
		Path:          "/test",
		Method:        "GET",
		Backend:       "://invalid-url", // Invalid URL
		Timeout:       1000,
		Headers:       map[string]string{},
		QueryParams:   map[string]string{},
		HasPathParams: false,
	}

	// Create a new proxy
	proxy := NewProxy(endpoint, false, nil)

	// Get the handler
	handler := proxy.Handler()

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// TestProxyHandlerWithPathParams tests the Handler method with path parameters
func TestProxyHandlerWithPathParams(t *testing.T) {
	// Create a mock backend server that verifies path parameters
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that path parameters were correctly extracted and used in the backend URL
		if !strings.HasSuffix(r.URL.Path, "/123") {
			t.Errorf("Expected path to end with '/123', got '%s'", r.URL.Path)
		}

		// Verify that path parameters were also added as query parameters
		if r.URL.Query().Get("id") != "123" {
			t.Errorf("Expected query param 'id' to be '123', got '%s'", r.URL.Query().Get("id"))
		}

		// Send a response with the path parameter
		// Extract just the ID from the path, regardless of the full path structure
		pathParts := strings.Split(r.URL.Path, "/")
		id := pathParts[len(pathParts)-1] // Get the last part of the path
		_, err := fmt.Fprintf(w, "User ID: %s", id)
		if err != nil {
			t.Errorf("Error on logging to console")
		}
	}))
	defer mockBackend.Close()

	// Create a test endpoint with path parameters
	endpoint := Endpoint{
		Path:          "/users/:id",
		Method:        "GET",
		Backend:       mockBackend.URL + "/api/users/:id",
		Timeout:       1000,
		Headers:       map[string]string{},
		QueryParams:   map[string]string{},
		HasPathParams: true,
	}

	// Create a new proxy
	proxy := NewProxy(endpoint, false, nil)

	// Get the handler
	handler := proxy.Handler()

	// Create a test request with a path parameter
	req, err := http.NewRequest("GET", "/users/123", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expectedBody := "User ID: 123"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

// TestProxyHandlerWithPreBackendCallback tests the Handler method with a pre-backend callback
func TestProxyHandlerWithPreBackendCallback(t *testing.T) {
	// Create a mock backend server that verifies the pre-backend callback was executed
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the custom header was added by the pre-backend callback
		if r.Header.Get("X-Pre-Callback") != "executed" {
			t.Errorf("Expected X-Pre-Callback header to be 'executed', got '%s'", r.Header.Get("X-Pre-Callback"))
		}

		// Send a response
		_, err := fmt.Fprintln(w, "Pre-backend callback test successful")
		if err != nil {
			t.Errorf("Error on logging to console")
		}
	}))
	defer mockBackend.Close()

	// Create a test endpoint
	endpoint := Endpoint{
		Path:          "/test-pre-callback",
		Method:        "GET",
		Backend:       mockBackend.URL,
		Timeout:       1000,
		Headers:       map[string]string{},
		QueryParams:   map[string]string{},
		HasPathParams: false,
	}

	// Create a new proxy
	proxy := NewProxy(endpoint, false, nil)

	// Add a pre-backend callback that adds a custom header
	callbackExecuted := false
	proxy.AddPreBackendCallback(func(req *http.Request) *http.Request {
		req.Header.Set("X-Pre-Callback", "executed")
		callbackExecuted = true
		return req
	})

	// Get the handler
	handler := proxy.Handler()

	// Create a test request
	req, err := http.NewRequest("GET", "/test-pre-callback", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Verify that the callback was executed
	if !callbackExecuted {
		t.Errorf("Pre-backend callback was not executed")
	}

	// Check the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expectedBody := "Pre-backend callback test successful\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

// TestProxyHandlerWithPostBackendCallback tests the Handler method with a post-backend callback
func TestProxyHandlerWithPostBackendCallback(t *testing.T) {
	// Create a mock backend server
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a response
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintln(w, `{"message": "Original response"}`)
		if err != nil {
			t.Errorf("Error on logging to console")
		}
	}))
	defer mockBackend.Close()

	// Create a test endpoint
	endpoint := Endpoint{
		Path:          "/test-post-callback",
		Method:        "GET",
		Backend:       mockBackend.URL,
		Timeout:       1000,
		Headers:       map[string]string{},
		QueryParams:   map[string]string{},
		HasPathParams: false,
	}

	// Create a new proxy
	proxy := NewProxy(endpoint, false, nil)

	// Add a post-backend callback that just marks it was executed
	// We're not trying to modify the response since that's difficult to test
	callbackExecuted := false
	proxy.AddPostBackendCallback(func(resp *http.Response, req *http.Request) *http.Response {
		// Mark the callback as executed
		callbackExecuted = true
		return resp
	})

	// Get the handler
	handler := proxy.Handler()

	// Create a test request
	req, err := http.NewRequest("GET", "/test-post-callback", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Verify that the callback was executed
	if !callbackExecuted {
		t.Errorf("Post-backend callback was not executed")
	}

	// Check the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body - we expect the original response since we're not modifying it
	expectedBody := `{"message": "Original response"}
`
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}
