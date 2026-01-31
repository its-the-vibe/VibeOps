package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/its-the-vibe/VibeOps/internal/utils"
	"github.com/spf13/cobra"
)

// NewDiffCmd creates the diff command
func NewDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare prev-build and build directories and restart changed services",
		Long: `Compare prev-build and build directories to identify changed services,
then trigger restarts via the TurnItOffAndOnAgain service. If TurnItOffAndOnAgain
itself is changed, it will be restarted first with a configurable wait time before
restarting other services.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			// Load configuration
			config, err := utils.LoadTurnItOffAndOnAgainConfig(configFile)
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			// Run diff command to compare directories
			changedServices, err := getDiffChangedServices()
			if err != nil {
				return fmt.Errorf("error getting changed services: %w", err)
			}

			if len(changedServices) == 0 {
				fmt.Println("No services changed between prev-build and build directories")
				return nil
			}

			fmt.Printf("Found %d changed service(s): %v\n", len(changedServices), changedServices)

			// Restart services
			if err := restartServices(changedServices, config); err != nil {
				return fmt.Errorf("error restarting services: %w", err)
			}

			fmt.Println("All services restarted successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("config", "c", "turn_it_off_and_on_again_config.json", "Path to TurnItOffAndOnAgain configuration file")
	return cmd
}

// getDiffChangedServices runs diff command and extracts unique service names
func getDiffChangedServices() ([]string, error) {
	// Run diff -qr prev-build build
	diffCmd := exec.Command("diff", "-qr", "prev-build", "build")
	output, err := diffCmd.CombinedOutput()

	// diff returns exit code 1 when differences are found, which is expected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means differences found (expected)
			// Exit code 2 means trouble (actual error)
			if exitErr.ExitCode() == 2 {
				return nil, fmt.Errorf("diff command failed: %w\nOutput: %s", err, output)
			}
		} else {
			return nil, fmt.Errorf("failed to run diff command: %w", err)
		}
	}

	// If no differences found, return empty slice
	if len(output) == 0 {
		return []string{}, nil
	}

	// Parse diff output to extract service names
	return parseServiceNames(string(output))
}

// parseServiceNames extracts unique service names from diff output
// Example line: "Files prev-build/its-the-vibe/ServiceName/file.json and build/its-the-vibe/ServiceName/file.json differ"
func parseServiceNames(diffOutput string) ([]string, error) {
	serviceMap := make(map[string]bool)

	// Regex to match "Files prev-build/.../... and build/.../... differ"
	// We'll extract the service name from the prev-build path
	re := regexp.MustCompile(`^Files prev-build/[^/]+/([^/]+)/`)

	lines := strings.Split(diffOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Ignore "Only in" lines as per requirements
		if strings.HasPrefix(line, "Only in") {
			continue
		}

		// Match the pattern and extract service name
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			serviceName := matches[1]
			serviceMap[serviceName] = true
		}
	}

	// Convert map to slice
	services := make([]string, 0, len(serviceMap))
	for service := range serviceMap {
		services = append(services, service)
	}

	return services, nil
}

// restartServices sends restart requests to TurnItOffAndOnAgain service
func restartServices(services []string, config *utils.TurnItOffAndOnAgainConfig) error {
	// Check if TurnItOffAndOnAgain is in the list of changed services
	var turnItOffAndOnAgainChanged bool
	var otherServices []string

	for _, service := range services {
		if service == "TurnItOffAndOnAgain" {
			turnItOffAndOnAgainChanged = true
		} else {
			otherServices = append(otherServices, service)
		}
	}

	// If TurnItOffAndOnAgain changed, restart it first
	if turnItOffAndOnAgainChanged {
		fmt.Println("TurnItOffAndOnAgain service changed, restarting it first...")
		if err := restartService("TurnItOffAndOnAgain", config); err != nil {
			return fmt.Errorf("failed to restart TurnItOffAndOnAgain: %w", err)
		}

		// Wait for configured time before restarting other services
		if len(otherServices) > 0 {
			fmt.Printf("Waiting %d seconds for TurnItOffAndOnAgain to restart...\n", config.RestartWaitSeconds)
			time.Sleep(time.Duration(config.RestartWaitSeconds) * time.Second)
		}
	}

	// Restart other services
	for _, service := range otherServices {
		if err := restartService(service, config); err != nil {
			return fmt.Errorf("failed to restart service %s: %w", service, err)
		}
	}

	return nil
}

// restartService sends a restart request for a single service
func restartService(serviceName string, config *utils.TurnItOffAndOnAgainConfig) error {
	url := fmt.Sprintf("%s/messages", config.TurnItOffAndOnAgainUrl)

	// Create payload
	payload := map[string]string{
		"restart": serviceName,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-success status code: %d", resp.StatusCode)
	}

	fmt.Printf("✓ Restarted service: %s\n", serviceName)
	return nil
}
