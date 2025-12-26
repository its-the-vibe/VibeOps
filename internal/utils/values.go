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
