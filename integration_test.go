package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// not parallel: os.Chdir mutates process-wide working directory
func TestEndToEndIntegration(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create go.mod
	goModContent := "module testintegration\n\ngo 1.21\n"
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err, "failed to create go.mod")

	// Create interface file
	interfaceDir := filepath.Join(tempDir, "internal", "app")
	err = os.MkdirAll(interfaceDir, 0o755)
	require.NoError(t, err, "failed to create interface directory")

	interfaceContent := `package app

type TestService interface {
	Process() error
	GetStatus() string
}
`
	interfaceFile := filepath.Join(interfaceDir, "service.go")
	err = os.WriteFile(interfaceFile, []byte(interfaceContent), 0o644)
	require.NoError(t, err, "failed to create interface file")

	// Create implementation files
	impl1Dir := filepath.Join(tempDir, "pkg", "impl1")
	err = os.MkdirAll(impl1Dir, 0o755)
	require.NoError(t, err, "failed to create impl1 directory")

	impl1Content := `package impl1

import "fmt"

type Worker struct {
	name string
}

func (w *Worker) Process() error {
	fmt.Printf("Worker %s processing\n", w.name)
	return nil
}

func (w *Worker) GetStatus() string {
	return "active"
}
`
	err = os.WriteFile(filepath.Join(impl1Dir, "worker.go"), []byte(impl1Content), 0o644)
	require.NoError(t, err, "failed to create impl1 file")

	impl2Dir := filepath.Join(tempDir, "pkg", "impl2")
	err = os.MkdirAll(impl2Dir, 0o755)
	require.NoError(t, err, "failed to create impl2 directory")

	impl2Content := `package impl2

type Processor struct {
	id int
}

func (p *Processor) Process() error {
	return nil
}

func (p *Processor) GetStatus() string {
	return "ready"
}
`
	err = os.WriteFile(filepath.Join(impl2Dir, "processor.go"), []byte(impl2Content), 0o644)
	require.NoError(t, err, "failed to create impl2 file")

	// Create a non-implementing struct
	impl3Dir := filepath.Join(tempDir, "pkg", "impl3")
	err = os.MkdirAll(impl3Dir, 0o755)
	require.NoError(t, err, "failed to create impl3 directory")

	impl3Content := `package impl3

type IncompleteService struct{}

func (i *IncompleteService) Process() error {
	return nil
}

// Missing GetStatus method
`
	err = os.WriteFile(filepath.Join(impl3Dir, "incomplete.go"), []byte(impl3Content), 0o644)
	require.NoError(t, err, "failed to create impl3 file")

	// Change to the temp directory
	oldDir, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(oldDir)
	})
	err = os.Chdir(tempDir)
	require.NoError(t, err, "failed to chdir into tempDir")

	// Run the finder
	finder := NewFinder("TestService")

	err = finder.validateGoModRoot()
	require.NoError(t, err, "validateGoModRoot failed")

	err = finder.loadModulePath()
	require.NoError(t, err, "loadModulePath failed")

	relInterfaceFile, _ := filepath.Rel(tempDir, interfaceFile)
	err = finder.parseInterface(relInterfaceFile)
	require.NoError(t, err, "parseInterface failed")

	err = finder.scanDirectory("pkg")
	require.NoError(t, err, "scanDirectory failed")

	results := finder.getResults()

	// Verify results
	require.Len(t, results, 2, "expected 2 implementations")

	// Check that we found the correct implementations
	foundWorker := false
	foundProcessor := false

	for _, result := range results {
		switch result.Struct {
		case "Worker":
			foundWorker = true
			assert.Equal(t, "impl1", result.Package, "Worker should be in package impl1")
		case "Processor":
			foundProcessor = true
			assert.Equal(t, "impl2", result.Package, "Processor should be in package impl2")
		default:
			assert.Fail(t, "unexpected struct found", "struct=%s", result.Struct)
		}
	}

	assert.True(t, foundWorker, "Worker implementation not found")
	assert.True(t, foundProcessor, "Processor implementation not found")

	// Verify JSON serialization
	jsonData, err := json.MarshalIndent(results, "", "  ")
	require.NoError(t, err, "failed to marshal results to JSON")

	var unmarshalled []Implementation
	err = json.Unmarshal(jsonData, &unmarshalled)
	require.NoError(t, err, "failed to unmarshal JSON")

	assert.Len(t, unmarshalled, len(results), "JSON roundtrip failed")
}

func TestParseInterfaceSpecIntegration(t *testing.T) {
	t.Parallel()

	// Test the complete flow with parseInterfaceSpec
	tempDir := t.TempDir()
	interfaceFile := filepath.Join(tempDir, "interface.go")

	err := os.WriteFile(interfaceFile, []byte("package test\ntype TestInterface interface{}"), 0o644)
	require.NoError(t, err, "failed to create interface file")

	spec := interfaceFile + ":TestInterface"
	file, name, err := parseInterfaceSpec(spec)
	require.NoError(t, err, "parseInterfaceSpec failed")

	assert.Equal(t, interfaceFile, file)
	assert.Equal(t, "TestInterface", name)
}

// not parallel: os.Chdir mutates process-wide working directory
func TestRunFinderIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod
	goModContent := "module testapp\n\ngo 1.21\n"
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err, "failed to create go.mod")

	// Create interface
	interfaceDir := filepath.Join(tempDir, "internal", "app")
	err = os.MkdirAll(interfaceDir, 0o755)
	require.NoError(t, err, "failed to create interface directory")

	interfaceContent := `package app

type Server interface {
	Start() error
	Stop() error
}`

	interfaceFile := filepath.Join(interfaceDir, "server.go")
	err = os.WriteFile(interfaceFile, []byte(interfaceContent), 0o644)
	require.NoError(t, err, "failed to create interface file")

	// Create implementation
	implDir := filepath.Join(tempDir, "pkg", "impl")
	err = os.MkdirAll(implDir, 0o755)
	require.NoError(t, err, "failed to create impl directory")

	implContent := `package impl

type WebServer struct{}

func (w *WebServer) Start() error { return nil }
func (w *WebServer) Stop() error { return nil }
`

	err = os.WriteFile(filepath.Join(implDir, "server.go"), []byte(implContent), 0o644)
	require.NoError(t, err, "failed to create impl file")

	// Change to temp directory
	oldDir, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(oldDir)
	})
	err = os.Chdir(tempDir)
	require.NoError(t, err, "failed to chdir into tempDir")

	// Test runFinder function directly
	searchDir := filepath.Join(tempDir, "pkg")

	// This should execute successfully and find the implementation
	// We can't easily capture the JSON output, but we can test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "runFinder panicked", "panic=%v", r)
		}
	}()

	// This will output to stdout, but at least tests the function
	runFinder(interfaceFile, "Server", searchDir)
}
