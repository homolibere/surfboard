package main

import (
	"strings"
)

// PathParamExtractor extracts path parameters from URLs
type PathParamExtractor struct{}

// Extract extracts path parameters from a request URL based on the pattern path
// For example, if the pattern path is "/api/users/:id" and the request path is "/api/users/123",
// this function will return a map with "id" -> "123"
func (p PathParamExtractor) Extract(patternPath, requestPath string) map[string]string {
	params := make(map[string]string)

	// Split the paths into segments
	patternSegments := strings.Split(patternPath, "/")
	requestSegments := strings.Split(requestPath, "/")

	// If the paths have different number of segments, return empty map
	if len(patternSegments) != len(requestSegments) {
		return params
	}

	// Compare each segment and extract parameters
	for i, patternSegment := range patternSegments {
		if i < len(requestSegments) {
			// Check if this segment is a parameter (starts with ":")
			if strings.HasPrefix(patternSegment, ":") {
				paramName := patternSegment[1:] // Remove the ":" prefix
				paramValue := requestSegments[i]
				params[paramName] = paramValue
			}
		}
	}

	return params
}
