package utils

import (
	"strings"
	"testing"
)

func TestValidateJSON_Valid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"object", []byte(`{"key": "value"}`)},
		{"array", []byte(`[1, 2, 3]`)},
		{"empty object", []byte(`{}`)},
		{"nested", []byte(`{"a": {"b": 1}}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateJSON(tt.data, "test.json"); err != nil {
				t.Errorf("ValidateJSON() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateJSON_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"bad syntax", []byte(`{bad}`)},
		{"truncated", []byte(`{"key":`)},
		{"empty", []byte(``)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateJSON(tt.data, "test.json"); err == nil {
				t.Error("ValidateJSON() expected error, got nil")
			}
		})
	}
}

func TestFormatJSONError_ContainsFilename(t *testing.T) {
	// Use a real json.Unmarshal error
	err := ValidateJSON([]byte(`{bad}`), "myfile.json")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "myfile.json") {
		t.Errorf("error message should contain filename, got: %v", err)
	}
}
