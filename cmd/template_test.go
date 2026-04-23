package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPathVars(t *testing.T) {
	values := map[string]interface{}{
		"OrgName": "my-org",
		"Env":     "prod",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"__.OrgName__/service", "my-org/service"},
		{"__.Env__/__.OrgName__", "prod/my-org"},
		{"no-placeholder", "no-placeholder"},
		{"__.Missing__/path", "__.Missing__/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandPathVars(tt.input, values)
			if result != tt.expected {
				t.Errorf("expandPathVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessTemplateFile_Basic(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple template file
	srcPath := filepath.Join(dir, "hello.tmpl")
	if err := os.WriteFile(srcPath, []byte("Hello, {{.Name}}!"), 0644); err != nil {
		t.Fatal(err)
	}

	values := map[string]interface{}{"Name": "World"}
	outPath, err := processTemplateFile(srcPath, buildDir, "hello.tmpl", values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", string(data))
	}
}

func TestProcessTemplateFile_RemovesTmplExtension(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}

	srcPath := filepath.Join(dir, "config.yaml.tmpl")
	if err := os.WriteFile(srcPath, []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}

	outPath, err := processTemplateFile(srcPath, buildDir, "config.yaml.tmpl", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.HasSuffix(outPath, ".tmpl") {
		t.Errorf("output path should not have .tmpl extension, got %q", outPath)
	}
	if !strings.HasSuffix(outPath, "config.yaml") {
		t.Errorf("output path should end with config.yaml, got %q", outPath)
	}
}

func TestProcessTemplateFile_InvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}

	srcPath := filepath.Join(dir, "bad.tmpl")
	if err := os.WriteFile(srcPath, []byte("{{.Unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := processTemplateFile(srcPath, buildDir, "bad.tmpl", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for invalid template, got nil")
	}
}

func TestProcessTemplates_Basic(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "source")
	buildDir := filepath.Join(dir, "build")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a template file
	if err := os.WriteFile(filepath.Join(sourceDir, "test.txt.tmpl"), []byte("value={{.Key}}"), 0644); err != nil {
		t.Fatal(err)
	}

	values := map[string]interface{}{"Key": "testval"}
	if err := processTemplates(sourceDir, buildDir, values, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(buildDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "value=testval" {
		t.Errorf("expected 'value=testval', got %q", string(data))
	}
}

func TestProcessTemplates_ExpandsPathVars(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "source")
	buildDir := filepath.Join(dir, "build")

	subDir := filepath.Join(sourceDir, "__.OrgName__")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt.tmpl"), []byte("org={{.OrgName}}"), 0644); err != nil {
		t.Fatal(err)
	}

	values := map[string]interface{}{"OrgName": "myorg"}
	if err := processTemplates(sourceDir, buildDir, values, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(buildDir, "myorg", "file.txt"))
	if err != nil {
		t.Fatalf("expected output file in expanded path: %v", err)
	}
	if string(data) != "org=myorg" {
		t.Errorf("expected 'org=myorg', got %q", string(data))
	}
}

func TestProcessTemplateFile_EnvFilePermissions(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}

	srcPath := filepath.Join(dir, ".env.tmpl")
	if err := os.WriteFile(srcPath, []byte("SECRET=value"), 0644); err != nil {
		t.Fatal(err)
	}

	outPath, err := processTemplateFile(srcPath, buildDir, ".env.tmpl", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0400 {
		t.Errorf("expected .env file permissions 0400, got %o", info.Mode().Perm())
	}
}
