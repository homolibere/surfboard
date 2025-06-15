package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

// LogEntry represents a structured log entry in JSON format
type LogEntry struct {
	Timestamp   string                 `json:"@timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Type        string                 `json:"type"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	RemoteAddr  string                 `json:"remote_addr,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Duration    string                 `json:"duration,omitempty"`
	Headers     map[string]interface{} `json:"headers,omitempty"`
	Body        string                 `json:"body,omitempty"`
	RequestDump string                 `json:"request_dump,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Additional  map[string]interface{} `json:"additional,omitempty"`
}

// LoggingResponseWriter is a wrapper around http.ResponseWriter that logs the status code
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

// WriteHeader captures the status code for logging
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body for logging
func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	// Write to the buffer for logging
	lrw.body.Write(b)
	// Write to the original ResponseWriter
	return lrw.ResponseWriter.Write(b)
}

// GetBody returns the captured response body
func (lrw *LoggingResponseWriter) GetBody() string {
	return lrw.body.String()
}

// NewLoggingResponseWriter creates a new LoggingResponseWriter
func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK, bytes.Buffer{}}
}

// LogJSON logs a message in JSON format
func LogJSON(entry LogEntry) {
	// Set timestamp if not already set
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to standard logging if JSON marshaling fails
		log.Printf("Error marshaling log entry to JSON: %v", err)
		return
	}

	// Print JSON log entry
	fmt.Println(string(jsonBytes))
}

// LogInfo logs an informational message in JSON format
func LogInfo(message string, additional map[string]interface{}) {
	LogJSON(LogEntry{
		Level:      "info",
		Message:    message,
		Type:       "log",
		Additional: additional,
	})
}

// LogError logs an error message in JSON format
func LogError(message string, err error, additional map[string]interface{}) {
	entry := LogEntry{
		Level:      "error",
		Message:    message,
		Type:       "log",
		Additional: additional,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	LogJSON(entry)
}

// LogFatal logs a fatal error message in JSON format and exits the program
func LogFatal(message string, err error, additional map[string]interface{}) {
	entry := LogEntry{
		Level:      "fatal",
		Message:    message,
		Type:       "log",
		Additional: additional,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	LogJSON(entry)
	os.Exit(1)
}

// LogRequest logs the details of an HTTP request in JSON format
func LogRequest(r *http.Request, debug bool) {
	// Create basic log entry
	entry := LogEntry{
		Type:       "request",
		Level:      "info",
		Message:    fmt.Sprintf("Request: %s %s", r.Method, r.URL.Path),
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
	}

	// Add debug information if enabled
	if debug {
		// Convert headers to map for JSON
		headers := make(map[string]interface{})
		for k, v := range r.Header {
			if len(v) == 1 {
				headers[k] = v[0]
			} else {
				headers[k] = v
			}
		}
		entry.Headers = headers

		// Log request body if present
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				entry.Error = fmt.Sprintf("Error reading request body: %v", err)
			} else {
				// Restore the body for further processing
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Log the body if not empty
				if len(bodyBytes) > 0 {
					entry.Body = string(bodyBytes)
				}
			}
		}

		// Log request dump for detailed debugging
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			entry.Error = fmt.Sprintf("Error dumping request: %v", err)
		} else {
			entry.RequestDump = string(requestDump)
		}
	}

	// Log the entry
	LogJSON(entry)
}

// LogResponse logs the details of an HTTP response in JSON format
func LogResponse(lrw *LoggingResponseWriter, r *http.Request, duration string, debug bool) {
	// Create basic log entry
	entry := LogEntry{
		Type:       "response",
		Level:      "info",
		Message:    fmt.Sprintf("Response: %d %s %s", lrw.statusCode, r.Method, r.URL.Path),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: lrw.statusCode,
		Duration:   duration,
	}

	// Add debug information if enabled
	if debug {
		// Log response body if present
		body := lrw.GetBody()
		if body != "" {
			entry.Body = body
		}
	}

	// Log the entry
	LogJSON(entry)
}
