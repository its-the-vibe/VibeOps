package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/its-the-vibe/VibeOps/internal/utils"
	"github.com/spf13/cobra"
)

// NewTemplateCmd creates the template command
func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Process template files and generate configuration files",
		Long:  `Process all .tmpl files in the source folder and generate output files in the build folder.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			buildDir, _ := cmd.Flags().GetString("build-dir")

			// Load values from values.json
			values, err := utils.LoadValuesFromFile("values.json")
			if err != nil {
				return fmt.Errorf("error loading values.json: %w", err)
			}

			// Load projects as []map[string]interface{} for template use
			projectsList, err := utils.LoadProjectsMap("projects.json")
			if err != nil {
				return fmt.Errorf("error loading projects.json: %w", err)
			}
			values["Projects"] = projectsList

			// Load ports from ports.json (optional)
			ports, err := utils.LoadValuesFromFile("ports.json")
			if err != nil {
				return fmt.Errorf("error loading ports.json: %w", err)
			}

			// Merge ports into values
			mergedValues := utils.MergeValues(values, ports)

			// Load bootstrap config (optional)
			bootstrapConfig, err := utils.LoadBootstrapConfig("bootstrap.json")
			if err != nil {
				// Bootstrap config is optional, silently skip if not found
			} else if bootstrapConfig.GCPSecretName != "" {
				// Load GCP secret if configured
				ctx := context.Background()
				gcpSecrets, err := utils.LoadGCPSecret(ctx, bootstrapConfig.GCPSecretName)
				if err != nil {
					return fmt.Errorf("error loading GCP secret: %w", err)
				}
				fmt.Printf("Loaded %d values from GCP Secret Manager\n", len(gcpSecrets))
				// Merge GCP secrets into values (GCP secrets override local values)
				mergedValues = utils.MergeValues(mergedValues, gcpSecrets)
			}

			// Process templates
			if err := processTemplates("source", buildDir, mergedValues); err != nil {
				return fmt.Errorf("error processing templates: %w", err)
			}

			fmt.Println("Templates processed successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("build-dir", "b", "build", "Output build directory")
	return cmd
}

// expandPathVars replaces __.Key__ placeholders in a path with values from the values map.
// For example, __.OrgName__ is replaced with the value of values["OrgName"].
func expandPathVars(path string, values map[string]interface{}) string {
	for key, val := range values {
		placeholder := "__." + key + "__"
		if str, ok := val.(string); ok {
			path = strings.ReplaceAll(path, placeholder, str)
		}
	}
	return path
}

// processTemplates walks through the source directory and processes .tmpl files
func processTemplates(sourceDir, buildDir string, values map[string]interface{}) error {
	// Create build directory if it doesn't exist
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Walk through the source directory
	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the source directory itself
		if path == sourceDir {
			return nil
		}

		// Calculate relative path from source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Expand __.Key__ placeholders in the relative path
		expandedRelPath := expandPathVars(relPath, values)

		// If it's a directory, create it in the build folder
		if d.IsDir() {
			buildPath := filepath.Join(buildDir, expandedRelPath)
			if err := os.MkdirAll(buildPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", buildPath, err)
			}
			return nil
		}

		// Only process .tmpl files
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Process the template file
		var outputFile string
		if outputFile, err = processTemplateFile(path, buildDir, expandedRelPath, values); err != nil {
			return fmt.Errorf("failed to process template %s: %w", path, err)
		}

		fmt.Printf("Processed: %s\n", outputFile)
		return nil
	})
}

// processTemplateFile reads a template file, applies values, and writes the output
func processTemplateFile(srcPath, buildDir, relPath string, values map[string]interface{}) (string, error) {
	// Read the template file
	tmplContent, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse the template
	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Remove .tmpl extension from the output filename
	outputRelPath := strings.TrimSuffix(relPath, ".tmpl")
	outputPath := filepath.Join(buildDir, outputRelPath)

	// Create parent directories if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}

	// Execute the template and write to output file
	if err := tmpl.Execute(outputFile, values); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Close the file before validation to ensure all data is flushed
	outputFile.Close()

	// Validate JSON files after generation
	if strings.HasSuffix(outputPath, ".json") {
		data, err := os.ReadFile(outputPath)
		if err != nil {
			return "", fmt.Errorf("failed to read generated file for validation: %w", err)
		}
		if err := utils.ValidateJSON(data, outputPath); err != nil {
			return "", fmt.Errorf("generated invalid JSON: %w", err)
		}
	}

	// Validate YAML files after generation
	if strings.HasSuffix(outputPath, ".yaml") || strings.HasSuffix(outputPath, ".yml") {
		data, err := os.ReadFile(outputPath)
		if err != nil {
			return "", fmt.Errorf("failed to read generated file for validation: %w", err)
		}
		if err := utils.ValidateYAML(data, outputPath); err != nil {
			return "", fmt.Errorf("generated invalid YAML: %w", err)
		}
	}

	return outputPath, nil
}
