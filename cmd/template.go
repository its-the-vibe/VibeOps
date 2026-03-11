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

// walkDirFollowSymlinks walks the directory tree rooted at root, calling fn for each file
// or directory in the tree, following symlinks. It detects symlink loops using a set of
// visited real paths to prevent infinite recursion.
func walkDirFollowSymlinks(root string, fn fs.WalkDirFunc) error {
	visited := make(map[string]bool)
	return walkDirFollowSymlinksRecursive(root, root, visited, fn)
}

func walkDirFollowSymlinksRecursive(root, path string, visited map[string]bool, fn fs.WalkDirFunc) error {
	// Resolve the real path to detect loops; EvalSymlinks follows all symlinks.
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return fn(path, nil, err)
		}
		return fn(path, fs.FileInfoToDirEntry(info), err)
	}

	// Use os.Stat(realPath) directly — this avoids a race between Lstat and Stat
	// and gives us the resolved target's info in a single call.
	realInfo, err := os.Stat(realPath)
	if err != nil {
		return fn(path, nil, err)
	}

	if realInfo.IsDir() {
		if visited[realPath] {
			// Symlink loop detected — skip silently
			return nil
		}
		visited[realPath] = true
		// Report the path as seen by the caller (preserving the symlink path),
		// but use realInfo so IsDir() returns true.
		if err := fn(path, fs.FileInfoToDirEntry(realInfo), nil); err != nil {
			if err == filepath.SkipDir {
				return nil
			}
			return err
		}
		// Read directory entries from realPath to get the actual contents.
		// Child paths are joined with path (not realPath) to keep them consistent
		// with the caller's sourceDir for correct relative-path resolution.
		entries, err := os.ReadDir(realPath)
		if err != nil {
			return fn(path, fs.FileInfoToDirEntry(realInfo), err)
		}
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			if err := walkDirFollowSymlinksRecursive(root, childPath, visited, fn); err != nil {
				if err == filepath.SkipDir {
					continue
				}
				return err
			}
		}
		return nil
	}

	// Regular file (or symlink to a file) — call fn with the resolved info.
	return fn(path, fs.FileInfoToDirEntry(realInfo), nil)
}

// NewTemplateCmd creates the template command
func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Process template files and generate configuration files",
		Long:  `Process all .tmpl files in the source folder and generate output files in the build folder.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			buildDir, _ := cmd.Flags().GetString("build-dir")
			sourceDir, _ := cmd.Flags().GetString("source-dir")
			followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")

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
			if err := processTemplates(sourceDir, buildDir, mergedValues, followSymlinks); err != nil {
				return fmt.Errorf("error processing templates: %w", err)
			}

			fmt.Println("Templates processed successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("build-dir", "b", "build", "Output build directory")
	cmd.Flags().StringP("source-dir", "s", "source", "Source directory containing template files")
	cmd.Flags().Bool("follow-symlinks", false, "Follow symlinks in the source directory when processing templates")
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
func processTemplates(sourceDir, buildDir string, values map[string]interface{}, followSymlinks bool) error {
	// Create build directory if it doesn't exist
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	walkFn := func(path string, d fs.DirEntry, err error) error {
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
	}

	// Walk through the source directory, optionally following symlinks
	if followSymlinks {
		return walkDirFollowSymlinks(sourceDir, walkFn)
	}
	return filepath.WalkDir(sourceDir, walkFn)
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

	// Set .env and .secret files to read-only for owner (0400) to protect sensitive data
	if strings.HasSuffix(outputPath, ".env") || strings.HasSuffix(outputPath, ".secret") {
		if err := os.Chmod(outputPath, 0400); err != nil {
			return "", fmt.Errorf("failed to set permissions on sensitive file: %w", err)
		}
	}

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
