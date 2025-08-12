package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// RequestCallback is a function that can modify a request before it's sent to the backend
// or after it's received from the backend
type RequestCallback func(req *http.Request) *http.Request

// ResponseCallback is a function that can modify a response before it's sent back to the client
type ResponseCallback func(resp *http.Response, req *http.Request) *http.Response

// Proxy handles the proxying of requests to backend services
type Proxy struct {
	endpoint             Endpoint
	debug                bool
	preBackendCallbacks  []RequestCallback
	postBackendCallbacks []ResponseCallback
	telemetry            *TelemetryManager
}

// NewProxy creates a new Proxy for the given endpoint
func NewProxy(endpoint Endpoint, debug bool, telemetry *TelemetryManager) *Proxy {
	return &Proxy{
		endpoint:             endpoint,
		debug:                debug,
		preBackendCallbacks:  []RequestCallback{},
		postBackendCallbacks: []ResponseCallback{},
		telemetry:            telemetry,
	}
}

// AddPreBackendCallback adds a callback to be executed before the request is sent to the backend
func (p *Proxy) AddPreBackendCallback(callback RequestCallback) {
	p.preBackendCallbacks = append(p.preBackendCallbacks, callback)
}

// AddPostBackendCallback adds a callback to be executed after the response is received from the backend
func (p *Proxy) AddPostBackendCallback(callback ResponseCallback) {
	p.postBackendCallbacks = append(p.postBackendCallbacks, callback)
}

// Handler returns an http.HandlerFunc that handles the proxying of requests
func (p *Proxy) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log incoming request
		LogRequest(r, p.debug)

		// Check if the request method matches the configured method
		if p.endpoint.Method != "" && r.Method != p.endpoint.Method {
			LogError("Method not allowed", nil, map[string]interface{}{
				"method":          r.Method,
				"expected_method": p.endpoint.Method,
				"path":            r.URL.Path,
			})
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the backend URL
		backendURL, err := url.Parse(p.endpoint.Backend)
		if err != nil {
			LogError("Invalid backend URL", err, map[string]interface{}{
				"backend_url": p.endpoint.Backend,
				"path":        r.URL.Path,
			})
			http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
			return
		}

		// Create a reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(backendURL)

		// Set up the director function to modify the request
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Set the Host header to the backend host
			req.Host = backendURL.Host

			// Handle path parameters if needed
			if p.endpoint.HasPathParams {
				// Extract path parameters from the request URL
				pathParams := p.endpoint.ExtractPathParams(r.URL.Path)

				// Replace path parameters in the backend URL
				backendPath := req.URL.Path
				for paramName, paramValue := range pathParams {
					backendPath = strings.Replace(backendPath, ":"+paramName, paramValue, -1)

					// Also add as query parameter for backends that might need it
					q := req.URL.Query()
					q.Set(paramName, paramValue)
					req.URL.RawQuery = q.Encode()
				}
				req.URL.Path = backendPath

				LogInfo("Path parameters extracted", map[string]interface{}{
					"path_params":  pathParams,
					"path":         r.URL.Path,
					"backend_path": backendPath,
				})
			}

			// Add custom headers
			for key, value := range p.endpoint.Headers {
				req.Header.Set(key, value)
			}

			// Add custom query parameters
			q := req.URL.Query()
			for key, value := range p.endpoint.QueryParams {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Execute pre-backend callbacks
			for _, callback := range p.preBackendCallbacks {
				req = callback(req)
			}

			if p.debug {
				LogInfo("Pre-backend callbacks executed", map[string]interface{}{
					"path":   req.URL.Path,
					"method": req.Method,
				})
			}
		}

		// Set timeout for the request
		if p.endpoint.Timeout > 0 {
			proxy.Transport = &http.Transport{
				ResponseHeaderTimeout: time.Duration(p.endpoint.Timeout) * time.Millisecond,
			}
		}

		// Set up the ModifyResponse function to execute post-backend callbacks
		proxy.ModifyResponse = func(resp *http.Response) error {
			// Execute post-backend callbacks
			for _, callback := range p.postBackendCallbacks {
				resp = callback(resp, r)
			}

			if p.debug {
				LogInfo("Post-backend callbacks executed", map[string]interface{}{
					"path":        r.URL.Path,
					"method":      r.Method,
					"status_code": resp.StatusCode,
				})
			}
			return nil
		}

		// Handle errors
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			LogError("Proxy error", err, map[string]interface{}{
				"path":    r.URL.Path,
				"method":  r.Method,
				"backend": p.endpoint.Backend,
			})
			http.Error(w, "Proxy error", http.StatusBadGateway)
		}

		// Create a logging response writer to capture the status code
		lrw := NewLoggingResponseWriter(w)

		// Serve the request
		proxy.ServeHTTP(lrw, r)

		// Log the response
		duration := time.Since(startTime)
		LogResponse(lrw, r, duration.String(), p.debug)

		// Record metrics if telemetry is enabled
		if p.telemetry != nil {
			p.telemetry.RecordRequest(
				r.Context(),
				p.endpoint.Path,
				r.Method,
				lrw.statusCode,
				float64(duration.Milliseconds()),
			)
		}
	}
}
