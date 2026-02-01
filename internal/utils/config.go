package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TurnItOffAndOnAgainConfig represents the configuration for the diff command
type TurnItOffAndOnAgainConfig struct {
	TurnItOffAndOnAgainUrl string `json:"TurnItOffAndOnAgainUrl"`
	RestartWaitSeconds     int    `json:"RestartWaitSeconds"`
}

// formatConfigJSONError creates a user-friendly error message for JSON parsing failures
func formatConfigJSONError(filename string, err error) error {
	errMsg := err.Error()
	
	// Extract line and column information if available
	if strings.Contains(errMsg, "line") || strings.Contains(errMsg, "offset") {
		return fmt.Errorf("invalid JSON in file '%s': %s\nPlease check the JSON syntax and ensure the file is properly formatted", filename, errMsg)
	}
	
	return fmt.Errorf("failed to parse JSON in file '%s': %w\nPlease check the JSON syntax and ensure the file is properly formatted", filename, err)
}

// LoadTurnItOffAndOnAgainConfig reads and parses the turn_it_off_and_on_again_config.json file
func LoadTurnItOffAndOnAgainConfig(filename string) (*TurnItOffAndOnAgainConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w\nPlease ensure the file exists and you have read permissions", filename, err)
	}

	var config TurnItOffAndOnAgainConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, formatConfigJSONError(filename, err)
	}

	// Set default wait time if not specified
	if config.RestartWaitSeconds == 0 {
		config.RestartWaitSeconds = 5
	}

	return &config, nil
}
