package cmd

import (
	"fmt"
	"os"

	"github.com/its-the-vibe/VibeOps/internal/utils"
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the new-project command
func NewProjectCmd() *cobra.Command {
	var noEnv bool

	cmd := &cobra.Command{
		Use:   "new-project [project-name]",
		Short: "Add a new project to projects.json",
		Long: `Add a new project to projects.json. The project will be added as a Docker project 
with default settings. All configuration files (SlackCompose, github-dispatcher, OctoCatalog) 
will be automatically generated from projects.json when you run 'vibeops template'.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Add project to root projects.json
			fmt.Printf("Adding project '%s' to projects.json...\n", projectName)
			if err := utils.AddProjectToProjectsFile("projects.json", projectName); err != nil {
				return fmt.Errorf("failed to add project to projects.json: %w", err)
			}
			fmt.Printf("✓ Added project to projects.json\n")

			// Create project directory and .env.tmpl file
			if err := createProjectDirAndEnv(projectName, noEnv); err != nil {
				return err
			}

			fmt.Printf("\n✓ Successfully added project '%s'!\n", projectName)
			fmt.Printf("Run 'vibeops template' to generate configuration files.\n")
			return nil
		},
	}

	cmd.Flags().BoolVar(&noEnv, "no-env", false, "Skip creation of the sample .env.tmpl file")

	return cmd
}

// createProjectDirAndEnv creates source/its-the-vibe/<projectName> and an empty .env file, idempotently
func createProjectDirAndEnv(projectName string, noEnv bool) error {
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

	// Skip .env.tmpl creation if --no-env flag is set
	if noEnv {
		fmt.Printf("Skipping .env.tmpl file creation (--no-env flag set)\n")
		return nil
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
