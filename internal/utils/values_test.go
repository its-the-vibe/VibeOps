package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValuesFromFile_Valid(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "values.json")
	data := `{"key": "value", "num": 42}`
	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	values, err := LoadValuesFromFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values["key"] != "value" {
		t.Errorf("expected key='value', got %v", values["key"])
	}
}

func TestLoadValuesFromFile_NotExist(t *testing.T) {
	values, err := LoadValuesFromFile("/nonexistent/path/values.json")
	if err != nil {
		t.Fatalf("expected empty map for missing file, got error: %v", err)
	}
	if len(values) != 0 {
		t.Errorf("expected empty map, got %v", values)
	}
}

func TestLoadValuesFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "values.json")
	if err := os.WriteFile(file, []byte(`{bad json}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadValuesFromFile(file)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestMergeValues(t *testing.T) {
	map1 := map[string]interface{}{"a": "1", "b": "2"}
	map2 := map[string]interface{}{"b": "override", "c": "3"}

	merged := MergeValues(map1, map2)

	if merged["a"] != "1" {
		t.Errorf("expected a='1', got %v", merged["a"])
	}
	if merged["b"] != "override" {
		t.Errorf("expected b='override' (map2 wins), got %v", merged["b"])
	}
	if merged["c"] != "3" {
		t.Errorf("expected c='3', got %v", merged["c"])
	}
}

func TestMergeValues_Empty(t *testing.T) {
	map1 := map[string]interface{}{"a": "1"}
	map2 := map[string]interface{}{}

	merged := MergeValues(map1, map2)
	if len(merged) != 1 || merged["a"] != "1" {
		t.Errorf("unexpected merged result: %v", merged)
	}
}

func TestMergeValues_DoesNotMutateInputs(t *testing.T) {
	map1 := map[string]interface{}{"a": "1"}
	map2 := map[string]interface{}{"b": "2"}

	_ = MergeValues(map1, map2)

	if _, ok := map1["b"]; ok {
		t.Error("MergeValues should not mutate map1")
	}
	if _, ok := map2["a"]; ok {
		t.Error("MergeValues should not mutate map2")
	}
}

func TestLoadValuesFromFile_Types(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "values.json")
	data := map[string]interface{}{
		"strVal":  "hello",
		"numVal":  42,
		"boolVal": true,
	}
	raw, _ := json.Marshal(data)
	if err := os.WriteFile(file, raw, 0644); err != nil {
		t.Fatal(err)
	}

	values, err := LoadValuesFromFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values["strVal"] != "hello" {
		t.Errorf("expected strVal='hello', got %v", values["strVal"])
	}
}
