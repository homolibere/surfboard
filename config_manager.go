package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ConfigManager handles loading and managing configuration
type ConfigManager struct{}

// NewConfigManager creates a new ConfigManager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadFromFile loads the API gateway configuration from a JSON file
func (cm *ConfigManager) LoadFromFile(filePath string) (Config, error) {
	// Read the configuration file
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Parse the JSON configuration
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// LoadDefault loads the default API gateway configuration
func (cm *ConfigManager) LoadDefault() Config {
	// This is a hardcoded default configuration
	// In a real application, this would be more minimal or load from environment variables
	return Config{
		Endpoints: []Endpoint{
			{
				Path:    "/api/users",
				Method:  "GET",
				Backend: "https://jsonplaceholder.typicode.com/users",
				Timeout: 5000,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				QueryParams:   map[string]string{},
				HasPathParams: false,
			},
			{
				Path:    "/api/posts",
				Method:  "GET",
				Backend: "https://jsonplaceholder.typicode.com/posts",
				Timeout: 5000,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				QueryParams:   map[string]string{},
				HasPathParams: false,
			},
			{
				Path:    "/api/users/:id",
				Method:  "GET",
				Backend: "https://jsonplaceholder.typicode.com/users/:id",
				Timeout: 5000,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				QueryParams:   map[string]string{},
				HasPathParams: true,
			},
			{
				Path:    "/api/posts/:id/comments",
				Method:  "GET",
				Backend: "https://jsonplaceholder.typicode.com/posts/:id/comments",
				Timeout: 5000,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				QueryParams:   map[string]string{},
				HasPathParams: true,
			},
		},
		Port:  9080,
		Debug: false,
		Telemetry: TelemetryConfig{
			Enabled:       true,
			MetricsURL:    "http://localhost:4318/v1/metrics",
			ServiceName:   "surfboard-gateway",
			ExportTimeout: 10000,
		},
	}
}
