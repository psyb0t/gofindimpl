package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateArgs(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.go")

	err := os.WriteFile(tempFile, []byte("package main"), 0644)
	require.NoError(t, err)

	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateArgs(tc.interfaceFile, "TestInterface", tc.searchDir)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
