package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSONError creates a user-friendly error message for JSON parsing failures
func FormatJSONError(filename string, err error) error {
	errMsg := err.Error()

	// Extract line and column information if available
	if strings.Contains(errMsg, "line") || strings.Contains(errMsg, "offset") {
		return fmt.Errorf("invalid JSON in file '%s': %s\nPlease check the JSON syntax and ensure the file is properly formatted", filename, errMsg)
	}

	return fmt.Errorf("failed to parse JSON in file '%s': %w\nPlease check the JSON syntax and ensure the file is properly formatted", filename, err)
}

// ValidateJSON validates that the given data is valid JSON
func ValidateJSON(data []byte, filename string) error {
	var js interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		return FormatJSONError(filename, err)
	}
	return nil
}
