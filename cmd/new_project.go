package cmd

import (
	"fmt"

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

			fmt.Printf("\n✓ Successfully added project '%s' to configuration files!\n", projectName)
			return nil
		},
	}

	return cmd
}
