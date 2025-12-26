package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	// Parse command line flags
	linkMode := flag.Bool("link", false, "Create symlinks from build directory to BaseDir")
	flag.Parse()

	// Load values from values.json
	values, err := loadValues("values.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading values.json: %v\n", err)
		os.Exit(1)
	}

	// Handle link mode
	if *linkMode {
		if err := createSymlinks("build", values); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating symlinks: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Symlinks created successfully!")
		return
	}

	// Process templates
	if err := processTemplates("source", "build", values); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing templates: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Templates processed successfully!")
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
		if err := processTemplateFile(path, buildDir, relPath, values); err != nil {
			return fmt.Errorf("failed to process template %s: %w", path, err)
		}

		fmt.Printf("Processed: %s\n", path)
		return nil
	})
}

// processTemplateFile reads a template file, applies values, and writes the output
func processTemplateFile(srcPath, buildDir, relPath string, values map[string]interface{}) error {
	// Read the template file
	tmplContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse the template
	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Remove .tmpl extension from the output filename
	outputRelPath := strings.TrimSuffix(relPath, ".tmpl")
	outputPath := filepath.Join(buildDir, outputRelPath)

	// Create parent directories if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Execute the template and write to output file
	if err := tmpl.Execute(outputFile, values); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
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
			return fmt.Errorf("failed to create symlink from %s to %s: %w", targetPath, absSourcePath, err)
		}

		fmt.Printf("Created symlink: %s -> %s\n", targetPath, absSourcePath)
		return nil
	})
}
