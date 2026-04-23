package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProjects_Valid(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")
	data := `[{"name":"ProjectA","allowVibeDeploy":true,"isDockerProject":true,"useWithSlackCompose":true,"useWithGitHubIssue":true}]`
	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	projects, err := LoadProjects(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "ProjectA" {
		t.Errorf("expected name='ProjectA', got %v", projects[0].Name)
	}
}

func TestLoadProjects_NotExist(t *testing.T) {
	_, err := LoadProjects("/nonexistent/projects.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadProjects_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(file, []byte(`not json`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadProjects(file)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestAddProjectToProjectsFile_New(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")
	// Start with an empty projects file
	if err := os.WriteFile(file, []byte(`[]`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := AddProjectToProjectsFile(file, "NewProject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	projects, err := LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "NewProject" {
		t.Errorf("expected name='NewProject', got %v", projects[0].Name)
	}
}

func TestAddProjectToProjectsFile_NoDuplicate(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")
	data := `[{"name":"ExistingProject","allowVibeDeploy":true,"isDockerProject":true,"useWithSlackCompose":true,"useWithGitHubIssue":true}]`
	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	// Adding the same project again should not duplicate
	if err := AddProjectToProjectsFile(file, "ExistingProject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	projects, err := LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project (no duplicate), got %d", len(projects))
	}
}

func TestAddProjectToProjectsFile_CreatesFileIfNotExist(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "new_projects.json")

	if err := AddProjectToProjectsFile(file, "BrandNewProject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	projects, err := LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Name != "BrandNewProject" {
		t.Errorf("unexpected projects: %+v", projects)
	}
}

func TestAddProjectToProjectsFile_SortedAlphabetically(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(file, []byte(`[]`), 0644); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"Zebra", "Apple", "Mango"} {
		if err := AddProjectToProjectsFile(file, name); err != nil {
			t.Fatal(err)
		}
	}

	projects, err := LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}
	expected := []string{"Apple", "Mango", "Zebra"}
	for i, p := range projects {
		if p.Name != expected[i] {
			t.Errorf("expected projects[%d].Name=%s, got %s", i, expected[i], p.Name)
		}
	}
}

func TestLoadProjectsMap_Valid(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")
	data := `[{"name":"MapProject","allowVibeDeploy":true,"isDockerProject":false,"useWithSlackCompose":false,"useWithGitHubIssue":false}]`
	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	projects, err := LoadProjectsMap(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0]["name"] != "MapProject" {
		t.Errorf("expected name='MapProject', got %v", projects[0]["name"])
	}
}
