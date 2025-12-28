package cmd

import (
	"fmt"
	"os"

	"github.com/its-the-vibe/VibeOps/internal/utils"
	"github.com/spf13/cobra"
)

// NewNewProjectCmd creates the new-project command
func NewNewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new-project [project-name]",
		Short: "Add a new project to configuration files",
		Long: `Add a new project to both projects.json.tmpl and github-dispatcher config.json.tmpl.
The project will be added with default commands: git pull, docker compose build, docker compose down, docker compose up -d`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Define file paths
			projectsFile := "source/its-the-vibe/SlackCompose/projects.json.tmpl"
			dispatcherFile := "source/its-the-vibe/github-dispatcher/config.json.tmpl"
			catalogFile := "source/its-the-vibe/OctoCatalog/catalog.json.tmpl"

			// Default commands as specified in the issue
			defaultCommands := []string{
				"git pull",
				"docker compose build",
				"docker compose down",
				"docker compose up -d",
			}

			// Add project to projects.json.tmpl
			fmt.Printf("Adding project '%s' to %s...\n", projectName, projectsFile)
			if err := utils.AddProjectToProjectsJSON(projectsFile, projectName); err != nil {
				return fmt.Errorf("failed to add project to projects.json.tmpl: %w", err)
			}
			fmt.Printf("✓ Added project to %s\n", projectsFile)

			// Add project to github-dispatcher config.json.tmpl
			fmt.Printf("Adding project '%s' to %s...\n", projectName, dispatcherFile)
			if err := utils.AddProjectToDispatcherConfig(dispatcherFile, projectName, defaultCommands); err != nil {
				return fmt.Errorf("failed to add project to config.json.tmpl: %w", err)
			}
			fmt.Printf("✓ Added project to %s\n", dispatcherFile)

			// Add project to OctoCatalog catalog.json.tmpl
			fmt.Printf("Adding project '%s' to %s...\n", projectName, catalogFile)
			if err := utils.AddProjectToOctoCatalog(catalogFile, projectName); err != nil {
				return fmt.Errorf("failed to add project to OctoCatalog catalog.json.tmpl: %w", err)
			}
			fmt.Printf("✓ Added project to %s\n", catalogFile)

			// Create project directory and .env.tmpl file
			if err := createProjectDirAndEnv(projectName); err != nil {
				return err
			}

			fmt.Printf("\n✓ Successfully added project '%s' to configuration files!\n", projectName)
			return nil
		},
	}

	return cmd
}

// createProjectDirAndEnv creates source/its-the-vibe/<projectName> and an empty .env file, idempotently
func createProjectDirAndEnv(projectName string) error {
	projectDir := fmt.Sprintf("source/its-the-vibe/%s", projectName)
	envFile := fmt.Sprintf("%s/.env.tmpl", projectDir)

	// Create directory if it doesn't exist
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
		fmt.Printf("✓ Created directory %s\n", projectDir)
	} else {
		fmt.Printf("Directory %s already exists\n", projectDir)
	}

	// Create empty .env.tmpl file if it doesn't exist
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		f, err := os.Create(envFile)
		if err != nil {
			return fmt.Errorf("failed to create .env.tmpl file: %w", err)
		}
		f.Close()
		fmt.Printf("✓ Created empty .env.tmpl file in %s\n", projectDir)
	} else {
		fmt.Printf(".env.tmpl file already exists in %s\n", projectDir)
	}

	return nil
}
