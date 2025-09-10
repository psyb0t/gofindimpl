package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.go")
	
	if err := os.WriteFile(tempFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name          string
		interfaceFile string
		searchDir     string
		expectError   bool
	}{
		{
			name:          "valid args",
			interfaceFile: tempFile,
			searchDir:     tempDir,
			expectError:   false,
		},
		{
			name:          "non-existent interface file",
			interfaceFile: filepath.Join(tempDir, "nonexistent.go"),
			searchDir:     tempDir,
			expectError:   true,
		},
		{
			name:          "non-existent search directory",
			interfaceFile: tempFile,
			searchDir:     filepath.Join(tempDir, "nonexistent"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.interfaceFile, "TestInterface", tt.searchDir)
			
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}