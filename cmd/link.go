package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/its-the-vibe/VibeOps/internal/utils"
	"github.com/spf13/cobra"
)

// NewLinkCmd creates the link command
func NewLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Create symlinks from build directory to BaseDir",
		Long:  `Walk through the build directory and create symlinks to the BaseDir specified in values.json.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			buildDir, _ := cmd.Flags().GetString("build-dir")

			// Load values from values.json
			values, err := utils.LoadValuesFromFile("values.json")
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

	cmd.Flags().StringP("build-dir", "b", "build", "Build directory to create symlinks from")
	return cmd
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
