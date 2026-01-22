package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadValues reads and parses the values.json file
func LoadValues(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
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
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var ports map[string]interface{}
	if err := json.Unmarshal(data, &ports); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
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
