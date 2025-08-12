# SurfBoard Development Guidelines

This document provides guidelines and information for developing and maintaining the SurfBoard API Gateway project.

## Build and Configuration

### Build Instructions

SurfBoard is a Go application that requires Go 1.24 or later. To build the project:

```bash
# Build the binary
go build -o SurfBoard

# Run the binary
./SurfBoard
```

### Command-Line Flags

SurfBoard supports the following command-line flags:

- `-port`: Port to listen on (overrides config)
- `-config`: Path to configuration file
- `-debug`: Enable debug mode with verbose logging

Example:

```bash
./SurfBoard -port 8080 -config config.json -debug
```

### Configuration

SurfBoard uses a JSON configuration file to define endpoints and other settings. If no configuration file is provided, a default configuration is used.

#### Configuration Format

```json
{
  "endpoints": [
    {
      "path": "/api/users",
      "method": "GET",
      "backend": "https://example.com/users",
      "timeout": 5000,
      "headers": {
        "Content-Type": "application/json"
      },
      "query_params": {
        "limit": "10"
      },
      "has_path_params": false
    },
    {
      "path": "/api/users/:id",
      "method": "GET",
      "backend": "https://example.com/users/:id",
      "timeout": 5000,
      "headers": {
        "Content-Type": "application/json"
      },
      "query_params": {},
      "has_path_params": true
    }
  ],
  "port": 9080,
  "debug": false,
  "telemetry": {
    "enabled": true,
    "metrics_url": "http://localhost:4318/v1/metrics",
    "service_name": "surfboard-gateway",
    "export_timeout": 10000
  }
}
```

#### Configuration Options

- `endpoints`: Array of endpoint configurations
  - `path`: URL path to match
  - `method`: HTTP method to match
  - `backend`: Backend service URL to proxy to
  - `timeout`: Request timeout in milliseconds
  - `headers`: Custom headers to add to the request
  - `query_params`: Query parameters to add to the request
  - `has_path_params`: Flag indicating if the path contains parameters (e.g., `/api/users/:id`)
- `port`: Port to listen on
- `debug`: Enable debug mode with verbose logging
- `telemetry`: Telemetry configuration
  - `enabled`: Enable telemetry
  - `metrics_url`: URL for exporting metrics
  - `service_name`: Service name for telemetry
  - `export_timeout`: Timeout for exporting metrics in milliseconds

## Testing

### Running Tests

To run all tests:

```bash
go test -v
```

To run a specific test:

```bash
go test -v -run TestName
```

For example:

```bash
go test -v -run TestLogInfo
```

### Adding Tests

Tests in SurfBoard follow the standard Go testing pattern with table-driven tests. Here's an example of a test for the `LogInfo` function:

``` go
func TestLogInfo(t *testing.T) {
    // Test cases
    tests := []struct {
        name       string
        message    string
        additional map[string]interface{}
        checkFunc  func(t *testing.T, logEntry map[string]interface{})
    }{
        {
            name:       "Basic message",
            message:    "Test message",
            additional: nil,
            checkFunc: func(t *testing.T, logEntry map[string]interface{}) {
                // Check basic fields
                if logEntry["level"] != "info" {
                    t.Errorf("Expected level to be 'info', got %v", logEntry["level"])
                }
                // ... more assertions ...
            },
        },
        // ... more test cases ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Capture stdout for this test case
            oldStdout := os.Stdout
            r, w, _ := os.Pipe()
            os.Stdout = w

            // Restore stdout when the test completes
            defer func() {
                os.Stdout = oldStdout
            }()

            // Call the function
            LogInfo(tt.message, tt.additional)

            // Close the write end of the pipe to flush the output
            w.Close()

            // Read the output
            var buf bytes.Buffer
            io.Copy(&buf, r)
            output := buf.String()

            // Parse the JSON output
            var logEntry map[string]interface{}
            if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry); err != nil {
                t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
            }

            // Run the check function
            tt.checkFunc(t, logEntry)
        })
    }
}
```

### Testing Patterns

1. **Table-Driven Tests**: Define test cases in a slice of structs, then iterate through them using `t.Run()` to create subtests.

2. **Mock HTTP Servers**: For testing HTTP handlers and proxies, use `httptest.NewServer` to create a mock HTTP server.

3. **Response Recorders**: Use `httptest.NewRecorder` to capture HTTP responses without making actual network requests.

4. **Stdout Capture**: For testing logging functions, capture stdout using `os.Pipe()` and restore it after the test.

5. **Error Checking**: Use `t.Errorf()` for non-fatal errors and `t.Fatalf()` for fatal errors.

## Project Structure

### Key Components

- **Gateway**: The main API gateway class that handles HTTP routing and endpoint registration.
- **Proxy**: Handles proxying requests to backend services.
- **ConfigManager**: Manages loading and parsing configuration.
- **TelemetryManager**: Manages metrics and telemetry using OpenTelemetry.
- **LoggingResponseWriter**: A wrapper around http.ResponseWriter that captures status codes and response bodies.

### Code Style

1. **Error Handling**: Errors are wrapped with context using `fmt.Errorf("failed to X: %w", err)`.

2. **Logging**: Structured logging is used throughout the project, with logs in JSON format.

3. **Configuration**: Configuration is loaded from a JSON file or defaults, with command-line flags taking precedence.

4. **Callbacks**: The proxy supports pre-backend and post-backend callbacks for request/response modification.

5. **Telemetry**: OpenTelemetry is used for metrics and observability.

### Development Workflow

1. **Adding a New Endpoint**: Add the endpoint configuration to the config.json file.

2. **Adding Custom Logic**: Use pre-backend and post-backend callbacks to add custom logic to endpoints.

3. **Adding Telemetry**: Use the TelemetryManager to record custom metrics.

4. **Debugging**: Enable debug mode with the `-debug` flag for verbose logging.

## Additional Information

### Path Parameters

SurfBoard supports path parameters in endpoint paths, such as `/api/users/:id`. The path parameters are extracted from the request URL and can be used in the backend URL.

### Health Check

SurfBoard automatically adds a `/health` endpoint that returns a simple status check.

### Metrics

If telemetry is enabled, SurfBoard adds a `/metrics` endpoint that exposes Prometheus metrics.

### Graceful Shutdown

SurfBoard supports graceful shutdown on SIGINT and SIGTERM signals.