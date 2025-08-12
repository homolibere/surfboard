package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Gateway is the main API gateway class
type Gateway struct {
	config    Config
	mux       *http.ServeMux
	proxies   map[string]*Proxy // Map of path to proxy for callback registration
	telemetry *TelemetryManager
}

// NewGateway creates a new Gateway with the given configuration and telemetry manager
func NewGateway(config Config, telemetry *TelemetryManager) *Gateway {
	return &Gateway{
		config:    config,
		mux:       http.NewServeMux(),
		proxies:   make(map[string]*Proxy),
		telemetry: telemetry,
	}
}

// RegisterEndpoints registers all endpoints from the configuration
func (g *Gateway) RegisterEndpoints() {
	for _, endpoint := range g.config.Endpoints {
		LogInfo("Registering endpoint", map[string]interface{}{
			"method":  endpoint.Method,
			"path":    endpoint.Path,
			"backend": endpoint.Backend,
		})
		proxy := NewProxy(endpoint, g.config.Debug, g.telemetry)
		g.proxies[endpoint.Path] = proxy
		g.mux.HandleFunc(endpoint.Path, proxy.Handler())
	}
}

// AddPreBackendCallback adds a callback to be executed before the request is sent to the backend
// for the specified endpoint path
func (g *Gateway) AddPreBackendCallback(path string, callback RequestCallback) {
	if proxy, ok := g.proxies[path]; ok {
		proxy.AddPreBackendCallback(callback)
		LogInfo("Pre-backend callback added", map[string]interface{}{
			"path": path,
		})
	} else {
		LogError("Failed to add pre-backend callback: endpoint not found", nil, map[string]interface{}{
			"path": path,
		})
	}
}

// AddPostBackendCallback adds a callback to be executed after the response is received from the backend
// for the specified endpoint path
func (g *Gateway) AddPostBackendCallback(path string, callback ResponseCallback) {
	if proxy, ok := g.proxies[path]; ok {
		proxy.AddPostBackendCallback(callback)
		LogInfo("Post-backend callback added", map[string]interface{}{
			"path": path,
		})
	} else {
		LogError("Failed to add post-backend callback: endpoint not found", nil, map[string]interface{}{
			"path": path,
		})
	}
}

// RegisterPreBackendCallbacks registers a pre-backend callback for all endpoints
func (g *Gateway) RegisterPreBackendCallbacks(callback RequestCallback) {
	for path, proxy := range g.proxies {
		proxy.AddPreBackendCallback(callback)
		LogInfo("Pre-backend callback registered for endpoint", map[string]interface{}{
			"path": path,
		})
	}
}

// RegisterPostBackendCallbacks registers a post-backend callback for all endpoints
func (g *Gateway) RegisterPostBackendCallbacks(callback ResponseCallback) {
	for path, proxy := range g.proxies {
		proxy.AddPostBackendCallback(callback)
		LogInfo("Post-backend callback registered for endpoint", map[string]interface{}{
			"path": path,
		})
	}
}

// RegisterHealthCheck adds a health check endpoint
func (g *Gateway) RegisterHealthCheck() {
	g.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log the health check request
		LogRequest(r, g.config.Debug)

		// Create a logging response writer
		lrw := NewLoggingResponseWriter(w)

		// Set response headers and write response
		lrw.Header().Set("Content-Type", "application/json")
		lrw.WriteHeader(http.StatusOK)
		err := json.NewEncoder(lrw).Encode(map[string]string{"status": "ok"})
		if err != nil {
			return
		}

		// Calculate duration
		duration := time.Since(startTime)

		// Log the response
		LogResponse(lrw, r, duration.String(), g.config.Debug)

		// Record metrics if telemetry is enabled
		if g.telemetry != nil {
			g.telemetry.RecordRequest(
				r.Context(),
				"/health",
				r.Method,
				lrw.statusCode,
				float64(duration.Milliseconds()),
			)
		}
	})
}

// RegisterMetricsEndpoint adds a metrics endpoint for Prometheus scraping
func (g *Gateway) RegisterMetricsEndpoint() {
	if g.telemetry == nil {
		LogInfo("Metrics endpoint not registered: telemetry is nil", nil)
		return
	}

	LogInfo("Registering metrics endpoint", nil)

	// Get the metrics handler from the telemetry manager
	metricsHandler := g.telemetry.GetMetricsHandler()

	// Register the metrics endpoint
	g.mux.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log the metrics request
		LogRequest(r, g.config.Debug)

		// Create a logging response writer
		lrw := NewLoggingResponseWriter(w)

		// Serve the metrics
		metricsHandler.ServeHTTP(lrw, r)

		// Calculate duration
		duration := time.Since(startTime)

		// Log the response
		LogResponse(lrw, r, duration.String(), g.config.Debug)

		// Record metrics for the metrics endpoint itself
		if g.telemetry != nil {
			g.telemetry.RecordRequest(
				r.Context(),
				"/metrics",
				r.Method,
				lrw.statusCode,
				float64(duration.Milliseconds()),
			)
		}
	}))
}

// Start starts the API gateway server
func (g *Gateway) Start() error {
	addr := fmt.Sprintf(":%d", g.config.Port)
	LogInfo("Starting API gateway", map[string]interface{}{
		"address": addr,
		"port":    g.config.Port,
	})

	if g.config.Debug {
		LogInfo("Debug mode enabled - verbose logging will be shown", nil)

		// Log configuration details
		configData := map[string]interface{}{
			"port":  g.config.Port,
			"debug": g.config.Debug,
		}
		LogInfo("Configuration", configData)

		// Log all registered endpoints
		LogInfo("Registered endpoints", nil)
		for i, endpoint := range g.config.Endpoints {
			endpointInfo := map[string]interface{}{
				"index":           i + 1,
				"method":          endpoint.Method,
				"path":            endpoint.Path,
				"backend":         endpoint.Backend,
				"timeout":         endpoint.Timeout,
				"has_path_params": endpoint.HasPathParams,
			}

			if len(endpoint.Headers) > 0 {
				endpointInfo["headers"] = endpoint.Headers
			}

			if len(endpoint.QueryParams) > 0 {
				endpointInfo["query_params"] = endpoint.QueryParams
			}

			LogInfo("Endpoint details", endpointInfo)
		}
	}

	return http.ListenAndServe(addr, g.mux)
}
