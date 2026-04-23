package utils

import (
	"testing"
)

func TestValidateYAML_Valid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"simple key-value", []byte("key: value\n")},
		{"list", []byte("- item1\n- item2\n")},
		{"nested", []byte("outer:\n  inner: val\n")},
		{"empty", []byte("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateYAML(tt.data, "test.yaml"); err != nil {
				t.Errorf("ValidateYAML() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateYAML_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"tab character", []byte("key:\n\tvalue\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateYAML(tt.data, "test.yaml"); err == nil {
				t.Errorf("ValidateYAML() expected error for %q, got nil", tt.name)
			}
		})
	}
}
