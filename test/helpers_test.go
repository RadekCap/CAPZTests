package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsKubectlApplySuccess(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		// Success cases
		{
			name:     "resource created",
			output:   "secret/my-secret created",
			expected: true,
		},
		{
			name:     "resource configured",
			output:   "secret/aso-credential configured",
			expected: true,
		},
		{
			name:     "resource unchanged",
			output:   "secret/cluster-identity-secret unchanged",
			expected: true,
		},
		{
			name:     "multiple resources mixed",
			output:   "secret/cluster-identity-secret unchanged\nsecret/aso-credential configured",
			expected: true,
		},
		{
			name:     "uppercase success indicator",
			output:   "Secret/my-secret CREATED",
			expected: true,
		},

		// Failure cases
		{
			name:     "error message",
			output:   "Error from server: secrets \"my-secret\" not found",
			expected: false,
		},
		{
			name:     "error with colon",
			output:   "error: unable to recognize file.yaml",
			expected: false,
		},
		{
			name:     "failed operation",
			output:   "failed to apply resource",
			expected: false,
		},
		{
			name:     "invalid resource",
			output:   "invalid object spec",
			expected: false,
		},
		{
			name:     "unable to connect",
			output:   "unable to connect to server",
			expected: false,
		},
		{
			name:     "warning message",
			output:   "Warning: resource will be deleted",
			expected: false,
		},
		{
			name:     "forbidden access",
			output:   "Error: forbidden - user does not have permission",
			expected: false,
		},
		{
			name:     "unauthorized",
			output:   "Error: Unauthorized",
			expected: false,
		},
		{
			name:     "not found",
			output:   "Error: not found",
			expected: false,
		},
		{
			name:     "empty output",
			output:   "",
			expected: false,
		},
		{
			name:     "whitespace only",
			output:   "   \n\t  ",
			expected: false,
		},
		{
			name:     "unexpected output",
			output:   "some random text with no indicators",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKubectlApplySuccess(tt.output)
			if result != tt.expected {
				t.Errorf("IsKubectlApplySuccess(%q) = %v, expected %v", tt.output, result, tt.expected)
			}
		})
	}
}

func TestValidateYAMLFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupFile   func(t *testing.T) string // Returns file path
		expectError bool
		errorMsg    string // Substring to match in error message
	}{
		{
			name: "valid YAML file",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "valid.yaml")
				content := []byte(`
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
data:
  key: value
`)
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: false,
		},
		{
			name: "valid YAML with multiple documents",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "multi.yaml")
				content := []byte(`
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
`)
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: false,
		},
		{
			name: "empty file",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "empty.yaml")
				if err := os.WriteFile(path, []byte{}, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "file is empty",
		},
		{
			name: "file with only whitespace",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "whitespace.yaml")
				content := []byte("   \n\n\t\t   \n  ")
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "invalid YAML syntax", // Whitespace with tabs/mixed content triggers parsing error
		},
		{
			name: "file with only comments",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "comments.yaml")
				content := []byte(`# This is a comment
# Another comment
# No actual data
`)
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "YAML file contains no data",
		},
		{
			name: "invalid YAML syntax - missing colon",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "invalid-syntax.yaml")
				content := []byte(`
apiVersion v1
kind: Secret
`)
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "invalid YAML syntax",
		},
		{
			name: "invalid YAML syntax - bad indentation",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "bad-indent.yaml")
				content := []byte(`
metadata:
name: test
  namespace: default
`)
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "invalid YAML syntax",
		},
		{
			name: "non-existent file",
			setupFile: func(t *testing.T) string {
				return filepath.Join(tmpDir, "does-not-exist.yaml")
			},
			expectError: true,
			errorMsg:    "file not accessible",
		},
		{
			name: "valid simple key-value YAML",
			setupFile: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "simple.yaml")
				content := []byte(`
key1: value1
key2: value2
nested:
  subkey: subvalue
`)
				if err := os.WriteFile(path, content, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			err := ValidateYAMLFile(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateYAMLFile(%q) expected error containing %q, got nil", filePath, tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateYAMLFile(%q) error = %q, expected to contain %q", filePath, err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateYAMLFile(%q) unexpected error: %v", filePath, err)
				}
			}
		})
	}
}
