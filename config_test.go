package main

import (
	"reflect"
	"testing"
)

// TestEndpointExtractPathParams tests the ExtractPathParams method of the Endpoint struct
func TestEndpointExtractPathParams(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       Endpoint
		requestPath    string
		expectedParams map[string]string
	}{
		{
			name: "No path parameters",
			endpoint: Endpoint{
				Path:          "/api/users",
				HasPathParams: false,
			},
			requestPath:    "/api/users",
			expectedParams: map[string]string{},
		},
		{
			name: "Single path parameter",
			endpoint: Endpoint{
				Path:          "/api/users/:id",
				HasPathParams: true,
			},
			requestPath:    "/api/users/123",
			expectedParams: map[string]string{"id": "123"},
		},
		{
			name: "Multiple path parameters",
			endpoint: Endpoint{
				Path:          "/api/users/:id/posts/:postId",
				HasPathParams: true,
			},
			requestPath:    "/api/users/123/posts/456",
			expectedParams: map[string]string{"id": "123", "postId": "456"},
		},
		{
			name: "Different segment count",
			endpoint: Endpoint{
				Path:          "/api/users/:id",
				HasPathParams: true,
			},
			requestPath:    "/api/users/123/extra",
			expectedParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.endpoint.ExtractPathParams(tt.requestPath)
			if !reflect.DeepEqual(params, tt.expectedParams) {
				t.Errorf("Endpoint.ExtractPathParams() = %v, want %v", params, tt.expectedParams)
			}
		})
	}
}
