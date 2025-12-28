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

// CatalogOption represents an option entry in OctoCatalog catalog.json.tmpl
type CatalogOption struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// CatalogItem represents an action item in OctoCatalog catalog.json.tmpl
type CatalogItem struct {
	ActionId string          `json:"actionId"`
	Options  []CatalogOption `json:"options"`
}

// AddProjectToOctoCatalog adds a new project option to OctoCatalog's catalog.json.tmpl under both SlackCompose and SlashVibeIssue actions
func AddProjectToOctoCatalog(filePath, projectName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var items []CatalogItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	foundSlack := false
	foundSlash := false
	for i, it := range items {
		if it.ActionId == "SlackCompose" || it.ActionId == "SlashVibeIssue" {
			// Check duplicate
			for _, opt := range it.Options {
				if opt.Value == projectName {
					// mark as found for this action and continue
					if it.ActionId == "SlackCompose" {
						foundSlack = true
					} else {
						foundSlash = true
					}
					goto CONTINUE
				}
			}

			// Add option
			items[i].Options = append(items[i].Options, CatalogOption{Text: projectName, Value: projectName})

			// Sort options by Text
			sort.Slice(items[i].Options, func(a, b int) bool {
				return items[i].Options[a].Text < items[i].Options[b].Text
			})

			if it.ActionId == "SlackCompose" {
				foundSlack = true
			} else {
				foundSlash = true
			}
		}
	CONTINUE:
	}

	missing := []string{}
	if !foundSlack {
		missing = append(missing, "SlackCompose")
	}
	if !foundSlash {
		missing = append(missing, "SlashVibeIssue")
	}
	if len(missing) > 0 {
		return fmt.Errorf("action(s) not found in %s: %v", filePath, missing)
	}

	output, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, append(output, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
