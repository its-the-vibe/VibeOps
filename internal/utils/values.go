package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// formatJSONError creates a user-friendly error message for JSON parsing failures
func formatJSONError(filename string, err error) error {
	errMsg := err.Error()
	
	// Extract line and column information if available
	if strings.Contains(errMsg, "line") || strings.Contains(errMsg, "offset") {
		return fmt.Errorf("invalid JSON in file '%s': %s\nPlease check the JSON syntax and ensure the file is properly formatted", filename, errMsg)
	}
	
	return fmt.Errorf("failed to parse JSON in file '%s': %w\nPlease check the JSON syntax and ensure the file is properly formatted", filename, err)
}

// LoadValues reads and parses the values.json file
func LoadValues(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w\nPlease ensure the file exists and you have read permissions", filename, err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, formatJSONError(filename, err)
	}

	return values, nil
}

// LoadPorts reads and parses the ports.json file
// Returns an empty map if the file doesn't exist (optional file)
func LoadPorts(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// If file doesn't exist, return empty map (ports.json is optional)
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read file '%s': %w\nPlease ensure you have read permissions", filename, err)
	}

	var ports map[string]interface{}
	if err := json.Unmarshal(data, &ports); err != nil {
		return nil, formatJSONError(filename, err)
	}

	return ports, nil
}

// MergeValues merges port values into the main values map
// Port values will override any existing values with the same key
func MergeValues(values, ports map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy all values
	for k, v := range values {
		merged[k] = v
	}

	// Merge in ports (will override if keys conflict)
	for k, v := range ports {
		merged[k] = v
	}

	return merged
}
