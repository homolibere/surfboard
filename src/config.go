package main

// Config represents the API gateway configuration
type Config struct {
	Endpoints []Endpoint      `json:"endpoints"`
	Port      int             `json:"port"`
	Debug     bool            `json:"debug"`
	Telemetry TelemetryConfig `json:"telemetry"`
}

// TelemetryConfig represents OpenTelemetry configuration
type TelemetryConfig struct {
	Enabled       bool   `json:"enabled"`
	MetricsURL    string `json:"metrics_url"`
	ServiceName   string `json:"service_name"`
	ExportTimeout int    `json:"export_timeout"`
}

// Endpoint represents a backend service endpoint configuration
type Endpoint struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Backend     string            `json:"backend"`
	Timeout     int               `json:"timeout"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	// HasPathParams indicates if the path contains parameters (e.g., /api/users/:id)
	HasPathParams bool `json:"has_path_params"`
}

// ExtractPathParams extracts path parameters from a request URL based on the endpoint path pattern
func (e *Endpoint) ExtractPathParams(requestPath string) map[string]string {
	return PathParamExtractor{}.Extract(e.Path, requestPath)
}
