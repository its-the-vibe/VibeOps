package main

import (
	"os"

	"github.com/its-the-vibe/VibeOps/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "vibeops",
		Short: "VibeOps - Templating and configuration management",
		Long:  `A Go-based templating system that processes template files and generates configuration files.`,
	}

	// Add commands to root
	rootCmd.AddCommand(cmd.NewTemplateCmd())
	rootCmd.AddCommand(cmd.NewLinkCmd())
	rootCmd.AddCommand(cmd.NewProjectCmd())
	rootCmd.AddCommand(cmd.NewDiffCmd())
	rootCmd.AddCommand(cmd.NewValidateCmd())

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
