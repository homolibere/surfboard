# SurfBoard Tests

This document describes the tests implemented for the SurfBoard API Gateway.

## Test Coverage

The test suite covers the following components:

1. **PathParamExtractor** - Tests the extraction of path parameters from URLs (via the `extractPathParams` compatibility function)
2. **LoggingResponseWriter** - Tests the response writer wrapper that captures status codes
3. **ConfigManager** - Tests loading configuration from JSON files (via the `LoadConfigFromFile` compatibility function)
4. **Proxy** - Tests the core proxy functionality (via the `ProxyHandler` compatibility function)
5. **Gateway** - Tests the health check endpoint
6. **ConfigManager** - Tests the default configuration loading (via the `LoadConfig` compatibility function)
7. **Endpoint** - Tests the ExtractPathParams method
8. **Gateway** - Tests the RegisterEndpoints and RegisterHealthCheck methods
9. **Proxy** - Tests error handling for invalid backend URLs and HTTP methods

Note: The tests use both compatibility functions that maintain the original API and direct class method calls to test the new class-based architecture.

## Running the Tests

To run all tests:

```bash
go test -v
```

To run a specific test:

```bash
go test -v -run TestExtractPathParams
go test -v -run TestProxyHandler
```

## Test Details

### TestExtractPathParams

Tests the function that extracts path parameters from request URLs. It covers:
- No path parameters
- Single path parameter
- Multiple path parameters
- Different segment count (error case)

### TestLoggingResponseWriter

Tests the wrapper around http.ResponseWriter that captures status codes. It verifies:
- Status code is correctly captured
- Status code is correctly written to the underlying ResponseWriter

### TestLoadConfigFromFile

Tests loading configuration from a JSON file. It verifies:
- Configuration is correctly parsed
- All fields are correctly loaded
- Port is correctly set
- Endpoint configuration is correctly loaded

### TestLoadConfigFromFileInvalid

Tests error handling when loading configuration from invalid files:
- Non-existent file
- Invalid JSON content

### TestProxyHandler

Tests the core proxy functionality with a mock backend server. It covers:
- Basic proxying
- Method filtering
- Path parameter handling
- Custom headers and query parameters

### TestHealthCheckEndpoint

Tests the health check endpoint to ensure it returns the correct status and response.

### TestLoadConfig

Tests the default configuration loading to ensure it provides the expected default values.

## Edge Cases

The tests cover several edge cases:
- Invalid configuration files
- Method not allowed
- Different path segment counts
- Various HTTP status codes

## Mock Server

For testing the proxy functionality, a mock HTTP server is created that echoes back request details, allowing verification of:
- Correct path handling
- Header forwarding
- Query parameter handling
- Path parameter substitution
