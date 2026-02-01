package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Project represents a project entry in projects.json
type Project struct {
	Name                string   `json:"name"`
	AllowVibeDeploy     bool     `json:"allowVibeDeploy"`
	IsDockerProject     bool     `json:"isDockerProject"`
	BuildCommands       []string `json:"buildCommands,omitempty"`
	UpCommands          []string `json:"upCommands,omitempty"`
	DownCommands        []string `json:"downCommands,omitempty"`
	UseWithSlackCompose bool     `json:"useWithSlackCompose"`
	UseWithGitHubIssue  bool     `json:"useWithGitHubIssue"`
	IsUpDownProject     bool     `json:"isUpDownProject"`
}

// formatProjectJSONError creates a user-friendly error message for JSON parsing failures
func formatProjectJSONError(filename string, err error) error {
	errMsg := err.Error()
	
	// Extract line and column information if available
	if strings.Contains(errMsg, "line") || strings.Contains(errMsg, "offset") {
		return fmt.Errorf("invalid JSON in file '%s': %s\nPlease check the JSON syntax and ensure the file is properly formatted", filename, errMsg)
	}
	
	return fmt.Errorf("failed to parse JSON in file '%s': %w\nPlease check the JSON syntax and ensure the file is properly formatted", filename, err)
}

// validateJSON validates that the given data is valid JSON
func validateJSON(data []byte, filename string) error {
	var js interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		return formatProjectJSONError(filename, err)
	}
	return nil
}

// LoadProjects reads and parses the projects.json file, setting defaults
func LoadProjects(filename string) ([]Project, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w\nPlease ensure the file exists and you have read permissions", filename, err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, formatProjectJSONError(filename, err)
	}

	return projects, nil
}

// LoadProjectsMap reads and parses the projects.json file, sets defaults, and returns []map[string]interface{} for template use
func LoadProjectsMap(filename string) ([]map[string]interface{}, error) {
	projects, err := LoadProjects(filename)
	if err != nil {
		return nil, err
	}
	var projectsList []map[string]interface{}
	b, err := json.Marshal(projects)
	if err != nil {
		return nil, fmt.Errorf("error marshalling projects from '%s': %w", filename, err)
	}
	if err := json.Unmarshal(b, &projectsList); err != nil {
		return nil, fmt.Errorf("error unmarshalling projects from '%s': %w", filename, err)
	}
	return projectsList, nil
}

// AddProjectToProjectsFile adds a new project to the root projects.json file
func AddProjectToProjectsFile(filePath, projectName string) error {
	// Read the file (or create empty array if file doesn't exist)
	var projects []Project
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File doesn't exist, start with empty array
			projects = []Project{}
		} else {
			return fmt.Errorf("failed to read file '%s': %w", filePath, err)
		}
	} else {
		// Parse existing projects
		if err := json.Unmarshal(data, &projects); err != nil {
			return formatProjectJSONError(filePath, err)
		}
	}

	// Check if project already exists
	for _, p := range projects {
		if p.Name == projectName {
			fmt.Printf("project '%s' already exists in projects.json\n", projectName)
			return nil
		}
	}

	newProject := Project{
		Name:                projectName,
		AllowVibeDeploy:     true,
		IsDockerProject:     true,
		UseWithSlackCompose: true,
		UseWithGitHubIssue:  true,
		IsUpDownProject:     true,
	}
	projects = append(projects, newProject)

	// Sort projects alphabetically by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	// Marshal back to JSON with proper formatting
	output, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for '%s': %w", filePath, err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, append(output, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", filePath, err)
	}

	// Validate the written JSON
	writtenData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s' for validation: %w", filePath, err)
	}
	if err := validateJSON(writtenData, filePath); err != nil {
		return fmt.Errorf("generated invalid JSON in '%s': %w", filePath, err)
	}

	return nil
}
