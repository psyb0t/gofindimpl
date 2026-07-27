package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFixtures(t *testing.T) {
	// not parallel: os.Chdir mutates process-wide working directory
	wd, err := os.Getwd()
	require.NoError(t, err)

	fixturesDir := filepath.Join(wd, ".fixtures")

	_, err = os.Stat(fixturesDir)
	require.False(t, os.IsNotExist(err), "fixtures directory does not exist: %s", fixturesDir)

	// Change to fixtures directory
	oldDir, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(oldDir)
	})

	require.NoError(t, os.Chdir(fixturesDir))

	// Test the finder with real fixtures
	finder := NewFinder("App")

	require.NoError(t, finder.validateGoModRoot())
	require.NoError(t, finder.loadModulePath())
	require.NoError(t, finder.parseInterface("internal/app/app.go"))

	// Verify interface methods were parsed correctly
	expectedMethods := []string{"Start", "Stop", "GetName"}
	require.Len(t, finder.interfaceMethods, len(expectedMethods))

	for i, method := range expectedMethods {
		assert.Equal(t, method, finder.interfaceMethods[i])
	}

	require.NoError(t, finder.scanDirectory("pkg/"))

	results := finder.getResults()

	// Should find exactly 3 implementations
	require.Len(t, results, 3)

	// Verify specific implementations
	expectedImplementations := map[string]string{
		"WebServer":     "something1",
		"ServiceDaemon": "something2",
		"MicroService":  "something3",
	}

	foundImplementations := make(map[string]string)
	for _, result := range results {
		foundImplementations[result.Struct] = result.Package
	}

	for expectedStruct, expectedPackage := range expectedImplementations {
		foundPackage, found := foundImplementations[expectedStruct]
		require.True(t, found, "expected implementation %s not found", expectedStruct)
		assert.Equal(t, expectedPackage, foundPackage)
	}

	// Verify JSON serialization works
	jsonData, err := json.MarshalIndent(results, "", "  ")
	require.NoError(t, err)

	// Verify JSON contains expected data
	assert.NotEmpty(t, jsonData, "JSON output should not be empty")
}

func TestParseInterfaceSpecWithFixtures(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)

	fixturesDir := filepath.Join(wd, ".fixtures")
	interfaceFile := filepath.Join(fixturesDir, "internal", "app", "app.go")

	// Test with fixtures path
	spec := interfaceFile + ":App"
	file, name, err := parseInterfaceSpec(spec)

	require.NoError(t, err)
	assert.Equal(t, interfaceFile, file)
	assert.Equal(t, "App", name)
}

func TestValidateArgsWithFixtures(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)

	fixturesDir := filepath.Join(wd, ".fixtures")
	interfaceFile := filepath.Join(fixturesDir, "internal", "app", "app.go")
	searchDir := filepath.Join(fixturesDir, "pkg")

	// This should not panic since files exist
	validateArgs(interfaceFile, "App", searchDir) //nolint:errcheck // exercising panic-free call path only

	// Test with non-existent file - this will call log.Fatal but we can't
	// easily test that
	// without changing the implementation, so we'll just verify the files exist
	_, err = os.Stat(interfaceFile)
	assert.False(t, os.IsNotExist(err), "interface file should exist: %s", interfaceFile)

	_, err = os.Stat(searchDir)
	assert.False(t, os.IsNotExist(err), "search directory should exist: %s", searchDir)
}

func TestFixturesDirectoryStructure(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)

	fixturesDir := filepath.Join(wd, ".fixtures")

	// Test that all expected fixture files exist
	expectedFiles := []string{
		"go.mod",
		"internal/app/app.go",
		"pkg/something1/webserver.go",
		"pkg/something2/daemon.go",
		"pkg/something3/microservice.go",
		"pkg/something4/worker.go",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(fixturesDir, file)
		_, err := os.Stat(fullPath)
		assert.False(t, os.IsNotExist(err), "expected fixture file does not exist: %s", fullPath)
	}

	// Verify go.mod content
	goModPath := filepath.Join(fixturesDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	assert.True(t, contains(string(content), "module testapp"),
		"go.mod should contain 'module testapp', got: %s", string(content))

	// Verify interface file content
	interfacePath := filepath.Join(fixturesDir, "internal/app/app.go")
	interfaceContent, err := os.ReadFile(interfacePath)
	require.NoError(t, err)

	assert.True(t, contains(string(interfaceContent), "type App interface"),
		"interface file should contain 'type App interface', got: %s", string(interfaceContent))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
