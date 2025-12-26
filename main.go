package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	buildDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "vibeops",
		Short: "VibeOps - Templating and configuration management",
		Long:  `A Go-based templating system that processes template files and generates configuration files.`,
	}

	// Template command
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Process template files and generate configuration files",
		Long:  `Process all .tmpl files in the source folder and generate output files in the build folder.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load values from values.json
			values, err := loadValues("values.json")
			if err != nil {
				return fmt.Errorf("error loading values.json: %w", err)
			}

			// Process templates
			if err := processTemplates("source", buildDir, values); err != nil {
				return fmt.Errorf("error processing templates: %w", err)
			}

			fmt.Println("Templates processed successfully!")
			return nil
		},
	}

	// Link command
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Create symlinks from build directory to BaseDir",
		Long:  `Walk through the build directory and create symlinks to the BaseDir specified in values.json.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load values from values.json
			values, err := loadValues("values.json")
			if err != nil {
				return fmt.Errorf("error loading values.json: %w", err)
			}

			// Create symlinks
			if err := createSymlinks(buildDir, values); err != nil {
				return fmt.Errorf("error creating symlinks: %w", err)
			}

			fmt.Println("Symlinks created successfully!")
			return nil
		},
	}

	// Add --build-dir flag to both commands
	templateCmd.Flags().StringVarP(&buildDir, "build-dir", "b", "build", "Output build directory")
	linkCmd.Flags().StringVarP(&buildDir, "build-dir", "b", "build", "Build directory to create symlinks from")

	// Add commands to root
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(linkCmd)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadValues reads and parses the values.json file
func loadValues(filename string) (map[string]interface{}, error) {
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

		// If it's a directory, create it in the build folder
		if d.IsDir() {
			buildPath := filepath.Join(buildDir, relPath)
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
		if outputFile, err = processTemplateFile(path, buildDir, relPath, values); err != nil {
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
	defer outputFile.Close()

	// Execute the template and write to output file
	if err := tmpl.Execute(outputFile, values); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return outputPath, nil
}

// createSymlinks walks through the build directory and creates symlinks to BaseDir
func createSymlinks(buildDir string, values map[string]interface{}) error {
	// Get BaseDir from values
	baseDir, ok := values["BaseDir"].(string)
	if !ok {
		return fmt.Errorf("BaseDir not found in values.json or is not a string")
	}

	// Check if build directory exists
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return fmt.Errorf("build directory does not exist: %s", buildDir)
	}

	// Walk through the build directory
	return filepath.WalkDir(buildDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Calculate relative path from build directory
		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Construct target path in BaseDir
		targetPath := filepath.Join(baseDir, relPath)

		// Create parent directories in target if needed
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
		}

		// Remove existing symlink or file if it exists
		if _, err := os.Lstat(targetPath); err == nil {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("failed to remove existing file %s: %w", targetPath, err)
			}
		}

		// Get absolute path of source file
		absSourcePath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path of %s: %w", path, err)
		}

		// Create symlink
		if err := os.Symlink(absSourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to create symlink from %s to %s: %w", absSourcePath, targetPath, err)
		}

		fmt.Printf("Created symlink: %s -> %s\n", targetPath, absSourcePath)
		return nil
	})
}
