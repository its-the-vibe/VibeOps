package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Project represents a project entry in projects.json
type Project struct {
	Name                string   `json:"name"`
	AllowVibeDeploy     *bool    `json:"allowVibeDeploy"`
	IsDockerProject     *bool    `json:"isDockerProject,omitempty"`
	BuildCommands       []string `json:"buildCommands,omitempty"`
	UseWithSlackCompose *bool    `json:"useWithSlackCompose,omitempty"`
	UseWithGitHubIssue  *bool    `json:"useWithGitHubIssue,omitempty"`
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

	// Set defaults
	for i := range projects {
		if projects[i].IsDockerProject == nil {
			def := true
			projects[i].IsDockerProject = &def
		}
		if projects[i].AllowVibeDeploy == nil {
			def := true
			projects[i].AllowVibeDeploy = &def
		}
		if projects[i].UseWithSlackCompose == nil {
			def := true
			projects[i].UseWithSlackCompose = &def
		}
		if projects[i].UseWithGitHubIssue == nil {
			def := true
			projects[i].UseWithGitHubIssue = &def
		}
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
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse existing projects
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if project already exists
	for _, p := range projects {
		if p.Name == projectName {
			fmt.Printf("project '%s' already exists in projects.json\n", projectName)
			return nil
		}
	}

	// Add new project entry with default values
	trueVal := true
	newProject := Project{
		Name:                projectName,
		AllowVibeDeploy:     &trueVal,
		IsDockerProject:     &trueVal,
		UseWithSlackCompose: &trueVal,
		UseWithGitHubIssue:  &trueVal,
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
