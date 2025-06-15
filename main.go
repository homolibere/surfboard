package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 0, "Port to listen on (overrides config)")
	configFile := flag.String("config", "", "Path to configuration file")
	debug := flag.Bool("debug", false, "Enable debug mode with verbose logging")
	flag.Parse()

	// Create a config manager
	configManager := NewConfigManager()

	// Load configuration
	var config Config
	if *configFile != "" {
		// Load configuration from file
		var err error
		config, err = configManager.LoadFromFile(*configFile)
		if err != nil {
			LogFatal("Failed to load configuration", err, nil)
		}
		LogInfo("Loaded configuration from file", map[string]interface{}{
			"file": *configFile,
		})
	} else {
		// Use default configuration
		config = configManager.LoadDefault()
		LogInfo("Using default configuration", nil)
	}

	// Override port if specified on command line
	if *port > 0 {
		config.Port = *port
	}

	// Override debug mode if specified on command line
	if *debug {
		config.Debug = true
		LogInfo("Debug mode enabled", nil)
	}

	// Initialize telemetry
	telemetry, err := NewTelemetryManager(config.Telemetry)
	if err != nil {
		LogFatal("Failed to initialize telemetry", err, nil)
	}
	if config.Telemetry.Enabled {
		LogInfo("Telemetry enabled", map[string]interface{}{
			"service_name": config.Telemetry.ServiceName,
			"metrics_url":  config.Telemetry.MetricsURL,
		})
	}

	// Create a context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		LogInfo("Received shutdown signal", nil)
		cancel()
	}()

	// Create and configure the gateway
	gateway := NewGateway(config, telemetry)
	gateway.RegisterEndpoints()
	gateway.RegisterHealthCheck()
	gateway.RegisterMetricsEndpoint()

	// Start the gateway in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- gateway.Start()
	}()

	// Wait for either context cancellation or an error from the gateway
	select {
	case <-ctx.Done():
		LogInfo("Shutting down gracefully", nil)
		// Shutdown telemetry
		if err := telemetry.Shutdown(context.Background()); err != nil {
			LogError("Error shutting down telemetry", err, nil)
		}
	case err := <-errCh:
		if err != nil {
			LogFatal("Failed to start gateway", err, nil)
		}
	}
}
