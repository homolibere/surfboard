# SurfBoard API Gateway

SurfBoard is a lightweight API Gateway inspired by KrakenD. It acts as a proxy between client applications and backend services, providing features like request routing, path parameter handling, and request/response logging.

## Features

- HTTP request proxying to backend services
- Path parameter support (e.g., `/api/users/:id`)
- Request method filtering
- Custom headers and query parameters
- Request/response logging
- Configurable timeouts
- JSON configuration file support
- Request/response callback hooks for custom processing

## Getting Started

### Installation

```bash
go get github.com/homolibere/surfboard
```

Or clone the repository:

```bash
git clone https://github.com/homolibere/SurfBoard.git
cd SurfBoard
go build
```

### Running the API Gateway

Run with default configuration:

```bash
./SurfBoard
```

Run with a custom configuration file:

```bash
./SurfBoard -config config.json
```

Specify a custom port:

```bash
./SurfBoard -port 9000
```

## Configuration

SurfBoard can be configured using a JSON file. Here's an example configuration:

```json
{
  "endpoints": [
    {
      "path": "/api/users",
      "method": "GET",
      "backend": "https://jsonplaceholder.typicode.com/users",
      "timeout": 5000,
      "headers": {
        "Content-Type": "application/json"
      },
      "query_params": {},
      "has_path_params": false
    },
    {
      "path": "/api/users/:id",
      "method": "GET",
      "backend": "https://jsonplaceholder.typicode.com/users/:id",
      "timeout": 5000,
      "headers": {
        "Content-Type": "application/json"
      },
      "query_params": {},
      "has_path_params": true
    }
  ],
  "port": 8080
}
```

### Configuration Options

- `endpoints`: Array of endpoint configurations
  - `path`: The path to match for incoming requests
  - `method`: The HTTP method to match (GET, POST, etc.)
  - `backend`: The backend service URL to proxy requests to
  - `timeout`: Request timeout in milliseconds
  - `headers`: Custom headers to add to the request
  - `query_params`: Custom query parameters to add to the request
  - `has_path_params`: Whether the path contains parameters (e.g., `:id`)
- `port`: The port to listen on

## Usage Examples

### Basic Request

```
GET http://localhost:8080/api/users
```

This will proxy the request to `https://jsonplaceholder.typicode.com/users`.

### Request with Path Parameters

```
GET http://localhost:8080/api/users/1
```

This will proxy the request to `https://jsonplaceholder.typicode.com/users/1`.

## Health Check

The API Gateway provides a health check endpoint:

```
GET http://localhost:8080/health
```

This will return a JSON response with status "ok" if the gateway is running.

## Architecture

SurfBoard uses a class-based architecture to organize its code. The main components are:

### Gateway

The `Gateway` class is the main entry point for the API gateway. It manages the HTTP server and routing.

```
gateway := NewGateway(config)
gateway.RegisterEndpoints()
gateway.RegisterHealthCheck()
gateway.Start()
```

### Proxy

The `Proxy` class handles the proxying of requests to backend services. Each endpoint has its own proxy.

```
proxy := NewProxy(endpoint)
handler := proxy.Handler()
```

### ConfigManager

The `ConfigManager` class handles loading and managing configuration.

```
configManager := NewConfigManager()
config := configManager.LoadFromFile("config.json")
// or
config := configManager.LoadDefault()
```

### PathParamExtractor

The `PathParamExtractor` class extracts path parameters from URLs.

```
extractor := PathParamExtractor{}
params := extractor.Extract("/api/users/:id", "/api/users/123")
```

### LoggingResponseWriter

The `LoggingResponseWriter` class is a wrapper around `http.ResponseWriter` that logs the status code.

```
lrw := NewLoggingResponseWriter(w)
// Use lrw instead of w
```

### Request/Response Callbacks

SurfBoard supports callbacks for custom request and response processing:

- **Pre-backend callbacks**: Executed after the proxy processes the request but before it's sent to the backend
- **Post-backend callbacks**: Executed after the backend processes the request but before the response is sent back to the client

#### Adding Pre-backend Callbacks

```
// Define a callback that modifies the request
preCallback := func(req *http.Request) *http.Request {
    // Modify the request
    req.Header.Set("X-Custom-Header", "custom-value")
    return req
}

// Add the callback to a specific endpoint
gateway.AddPreBackendCallback("/api/users", preCallback)
```

#### Adding Post-backend Callbacks

```
// Define a callback that modifies the response
postCallback := func(resp *http.Response, req *http.Request) *http.Response {
    // Modify the response
    resp.Header.Set("X-Response-Header", "response-value")
    return resp
}

// Add the callback to a specific endpoint
gateway.AddPostBackendCallback("/api/users", postCallback)
```

#### Registering Callbacks for All Endpoints

You can also register callbacks for all endpoints at once:

```
// Register a pre-backend callback for all endpoints
gateway.RegisterPreBackendCallbacks(func(req *http.Request) *http.Request {
    // Modify the request for all endpoints
    req.Header.Set("X-Global-Header", "global-value")
    return req
})

// Register a post-backend callback for all endpoints
gateway.RegisterPostBackendCallbacks(func(resp *http.Response, req *http.Request) *http.Response {
    // Modify the response for all endpoints
    resp.Header.Set("X-Global-Response-Header", "global-response-value")
    return resp
})
```

These callbacks can be used for various purposes such as:
- Adding custom authentication/authorization
- Request/response transformation
- Logging and monitoring
- Caching
- Rate limiting

## Testing

SurfBoard comes with a comprehensive test suite that covers all major components:

- Gateway class tests
- Proxy class tests
- ConfigManager class tests
- PathParamExtractor class tests
- LoggingResponseWriter class tests
- Endpoint class tests

The tests cover both normal operation and error conditions, ensuring the application behaves correctly in all scenarios.

For detailed information about the tests, see [README_TESTS.md](README_TESTS.md).

To run the tests:

```bash
go test -v
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
