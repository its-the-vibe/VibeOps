package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadValuesFromFile reads and parses the specified file
func LoadValuesFromFile(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist, returning empty values map:", filename)
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read file '%s': %w. Please check file permissions", filename, err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, FormatJSONError(filename, err)
	}

	return values, nil
}

// MergeValues merges two maps
// Values from the second map will override any existing values with the same key
func MergeValues(map1, map2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy all values from map1
	for k, v := range map1 {
		merged[k] = v
	}

	// Merge in map2 (will override if keys conflict)
	for k, v := range map2 {
		merged[k] = v
	}

	return merged
}
