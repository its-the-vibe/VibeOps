package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// TurnItOffAndOnAgainConfig represents the configuration for the diff command
type TurnItOffAndOnAgainConfig struct {
	TurnItOffAndOnAgainUrl string `json:"TurnItOffAndOnAgainUrl"`
	RestartWaitSeconds     int    `json:"RestartWaitSeconds"`
}

// LoadTurnItOffAndOnAgainConfig reads and parses the turn_it_off_and_on_again_config.json file
func LoadTurnItOffAndOnAgainConfig(filename string) (*TurnItOffAndOnAgainConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file '%s' not found. Please ensure the file exists", filename)
		}
		return nil, fmt.Errorf("failed to read file '%s': %w. Please check file permissions", filename, err)
	}

	var config TurnItOffAndOnAgainConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, FormatJSONError(filename, err)
	}

	// Set default wait time if not specified
	if config.RestartWaitSeconds == 0 {
		config.RestartWaitSeconds = 5
	}

	return &config, nil
}
