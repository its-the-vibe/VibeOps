package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Project represents a project entry in projects.json
type Project struct {
	Name                string   `json:"name"`
	AllowVibeDeploy     bool     `json:"allowVibeDeploy"`
	IsDockerProject     bool     `json:"isDockerProject"`
	BuildCommands       []string `json:"buildCommands,omitempty"`
	UseWithSlackCompose bool     `json:"useWithSlackCompose"`
	UseWithGitHubIssue  bool     `json:"useWithGitHubIssue"`
}

// LoadProjects reads and parses the projects.json file, setting defaults
func LoadProjects(filename string) ([]Project, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
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
		return nil, fmt.Errorf("error marshalling projects: %w", err)
	}
	if err := json.Unmarshal(b, &projectsList); err != nil {
		return nil, fmt.Errorf("error unmarshalling projects: %w", err)
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
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		// Parse existing projects
		if err := json.Unmarshal(data, &projects); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
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
	}
	projects = append(projects, newProject)

	// Sort projects alphabetically by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	// Marshal back to JSON with proper formatting
	output, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, append(output, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
