package cmd

import (
	"fmt"
	"os"

	"github.com/its-the-vibe/VibeOps/internal/utils"
	"github.com/spf13/cobra"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all JSON configuration files",
		Long:  `Validate that all JSON configuration files (values.json, ports.json, projects.json, config.json) are valid and well-formed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			hasErrors := false

			// Validate values.json
			if err := validateFile("values.json", true); err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				hasErrors = true
			} else {
				fmt.Println("✓ values.json is valid")
			}

			// Validate ports.json (optional)
			if err := validateFile("ports.json", false); err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				hasErrors = true
			} else {
				printOptionalFileStatus("ports.json")
			}

			// Validate projects.json
			if err := validateFile("projects.json", true); err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				hasErrors = true
			} else {
				fmt.Println("✓ projects.json is valid")
			}

			// Validate config.json (optional)
			if err := validateFile("config.json", false); err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				hasErrors = true
			} else {
				printOptionalFileStatus("config.json")
			}

			if hasErrors {
				return fmt.Errorf("validation failed for one or more JSON files")
			}

			fmt.Println("\n✓ All JSON files are valid!")
			return nil
		},
	}

	return cmd
}

// printOptionalFileStatus prints the status of an optional file
func printOptionalFileStatus(filename string) {
	if fileExists(filename) {
		fmt.Printf("✓ %s is valid\n", filename)
	} else {
		fmt.Printf("ℹ %s not found (optional file)\n", filename)
	}
}

// validateFile validates a single JSON file
func validateFile(filename string, required bool) error {
	// Check if file exists
	if !fileExists(filename) {
		if required {
			return fmt.Errorf("required file '%s' not found", filename)
		}
		// Optional file not found is not an error
		return nil
	}

	// Read and validate the file based on its type
	switch filename {
	case "values.json":
		_, err := utils.LoadValues(filename)
		return err
	case "ports.json":
		_, err := utils.LoadPorts(filename)
		return err
	case "projects.json":
		_, err := utils.LoadProjects(filename)
		return err
	case "config.json":
		_, err := utils.LoadTurnItOffAndOnAgainConfig(filename)
		return err
	default:
		// Generic JSON validation
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %w", filename, err)
		}
		return utils.ValidateJSON(data, filename)
	}
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
