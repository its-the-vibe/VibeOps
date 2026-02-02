package utils

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const yamlSyntaxSuggestion = "Please check the YAML syntax and ensure the file is properly formatted"

// FormatYAMLError creates a user-friendly error message for YAML parsing failures
func FormatYAMLError(filename string, err error) error {
	errMsg := err.Error()

	// Extract line and column information if available
	if strings.Contains(errMsg, "line") || strings.Contains(errMsg, "column") {
		return fmt.Errorf("invalid YAML in file '%s': %s\n%s", filename, errMsg, yamlSyntaxSuggestion)
	}

	return fmt.Errorf("failed to parse YAML in file '%s': %w\n%s", filename, err, yamlSyntaxSuggestion)
}

// ValidateYAML validates that the given data is valid YAML
func ValidateYAML(data []byte, filename string) error {
	var y interface{}
	if err := yaml.Unmarshal(data, &y); err != nil {
		return FormatYAMLError(filename, err)
	}
	return nil
}
