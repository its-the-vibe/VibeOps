package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// ProjectEntry represents an entry in projects.json.tmpl
type ProjectEntry struct {
	Name       string `json:"name"`
	WorkingDir string `json:"working_dir"`
}

// DispatcherEntry represents an entry in config.json.tmpl for github-dispatcher
type DispatcherEntry struct {
	Repo     string   `json:"repo"`
	Branch   string   `json:"branch"`
	Type     string   `json:"type"`
	Dir      string   `json:"dir"`
	Commands []string `json:"commands"`
}

// AddProjectToProjectsJSON adds a new project to projects.json.tmpl
func AddProjectToProjectsJSON(filePath, projectName string) error {
	// Read the template file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse existing entries
	var projects []ProjectEntry
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if project already exists
	for _, p := range projects {
		if p.Name == projectName {
			return fmt.Errorf("project '%s' already exists in projects.json.tmpl", projectName)
		}
	}

	// Add new project entry
	newProject := ProjectEntry{
		Name:       projectName,
		WorkingDir: fmt.Sprintf("{{.BaseDir}}/{{.OrgName}}/%s", projectName),
	}
	projects = append(projects, newProject)

	// Sort projects alphabetically by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	// Marshal back to JSON with proper formatting
	output, err := json.MarshalIndent(projects, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, append(output, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// AddProjectToDispatcherConfig adds a new project to github-dispatcher config.json.tmpl
func AddProjectToDispatcherConfig(filePath, projectName string, commands []string) error {
	// Read the template file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse existing entries
	var configs []DispatcherEntry
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if project already exists
	repoName := fmt.Sprintf("{{.OrgName}}/%s", projectName)
	for _, c := range configs {
		if c.Repo == repoName {
			return fmt.Errorf("project '%s' already exists in github-dispatcher config.json.tmpl", projectName)
		}
	}

	// Add new dispatcher entry
	newEntry := DispatcherEntry{
		Repo:     repoName,
		Branch:   "refs/heads/main",
		Type:     "git-webhook",
		Dir:      fmt.Sprintf("{{.BaseDir}}/{{.OrgName}}/%s", projectName),
		Commands: commands,
	}
	configs = append(configs, newEntry)

	// Sort configs alphabetically by repo name
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Repo < configs[j].Repo
	})

	// Marshal back to JSON with proper formatting
	output, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, append(output, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
