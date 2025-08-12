package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// TelemetryManager handles OpenTelemetry metrics
type TelemetryManager struct {
	config           TelemetryConfig
	meter            metric.Meter
	meterProvider    *sdkmetric.MeterProvider
	requestCounter   metric.Int64Counter
	latencyHistogram metric.Float64Histogram
	errorCounter     metric.Int64Counter
	promHandler      http.Handler
}

// NewTelemetryManager creates a new TelemetryManager
func NewTelemetryManager(config TelemetryConfig) (*TelemetryManager, error) {
	if !config.Enabled {
		return &TelemetryManager{config: config}, nil
	}

	// Create resource
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	// Create Prometheus exporter
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create OTLP exporter for remote metrics collection
	// Parse the metrics URL to extract host and port
	metricsURL, err := url.Parse(config.MetricsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics URL: %w", err)
	}

	// Validate URL scheme (must be http or https)
	if metricsURL.Scheme != "http" && metricsURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid metrics URL scheme: %s (must be http or https)", metricsURL.Scheme)
	}

	// Extract host and port (without path)
	endpoint := metricsURL.Host

	otlpExporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithTimeout(time.Duration(config.ExportTimeout)*time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create meter provider with both exporters
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				otlpExporter,
				sdkmetric.WithInterval(5*time.Second),
			),
		),
		sdkmetric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := meterProvider.Meter("surfboard-gateway")

	// Create metrics
	requestCounter, err := meter.Int64Counter(
		"http.request.count",
		metric.WithDescription("Number of HTTP requests"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	latencyHistogram, err := meter.Float64Histogram(
		"http.request.duration",
		metric.WithDescription("HTTP request duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create latency histogram: %w", err)
	}

	errorCounter, err := meter.Int64Counter(
		"http.request.errors",
		metric.WithDescription("Number of HTTP request errors"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create error counter: %w", err)
	}

	// Create Prometheus HTTP handler
	promHandler := promhttp.Handler()

	return &TelemetryManager{
		config:           config,
		meter:            meter,
		meterProvider:    meterProvider,
		requestCounter:   requestCounter,
		latencyHistogram: latencyHistogram,
		errorCounter:     errorCounter,
		promHandler:      promHandler,
	}, nil
}

// RecordRequest records metrics for an HTTP request
func (tm *TelemetryManager) RecordRequest(ctx context.Context, path, method string, statusCode int, durationMs float64) {
	if !tm.config.Enabled {
		return
	}

	// Create attributes
	attrs := []attribute.KeyValue{
		attribute.String("http.route", path),
		attribute.String("http.method", method),
		attribute.Int("http.status_code", statusCode),
	}

	// Record metrics
	tm.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	tm.latencyHistogram.Record(ctx, durationMs, metric.WithAttributes(attrs...))

	// Record errors (status code >= 400)
	if statusCode >= 400 {
		tm.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// Shutdown shuts down the telemetry manager
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	if !tm.config.Enabled || tm.meterProvider == nil {
		return nil
	}
	return tm.meterProvider.Shutdown(ctx)
}

// GetMetricsHandler returns an HTTP handler for metrics endpoint
func (tm *TelemetryManager) GetMetricsHandler() http.Handler {
	if !tm.config.Enabled || tm.promHandler == nil {
		// Return a simple handler that returns 404 if telemetry is disabled
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Telemetry is disabled", http.StatusNotFound)
		})
	}
	return tm.promHandler
}
