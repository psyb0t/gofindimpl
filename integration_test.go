package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEndToEndIntegration(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	
	// Create go.mod
	goModContent := "module testintegration\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"),
		[]byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create interface file
	interfaceDir := filepath.Join(tempDir, "internal", "app")
	if err := os.MkdirAll(interfaceDir, 0755); err != nil {
		t.Fatalf("failed to create interface directory: %v", err)
	}

	interfaceContent := `package app

type TestService interface {
	Process() error
	GetStatus() string
}
`
	interfaceFile := filepath.Join(interfaceDir, "service.go")
	if err := os.WriteFile(interfaceFile,
		[]byte(interfaceContent), 0644); err != nil {
		t.Fatalf("failed to create interface file: %v", err)
	}

	// Create implementation files
	impl1Dir := filepath.Join(tempDir, "pkg", "impl1")
	if err := os.MkdirAll(impl1Dir, 0755); err != nil {
		t.Fatalf("failed to create impl1 directory: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(impl1Dir, "worker.go"),
		[]byte(impl1Content), 0644); err != nil {
		t.Fatalf("failed to create impl1 file: %v", err)
	}

	impl2Dir := filepath.Join(tempDir, "pkg", "impl2")
	if err := os.MkdirAll(impl2Dir, 0755); err != nil {
		t.Fatalf("failed to create impl2 directory: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(impl2Dir, "processor.go"),
		[]byte(impl2Content), 0644); err != nil {
		t.Fatalf("failed to create impl2 file: %v", err)
	}

	// Create a non-implementing struct
	impl3Dir := filepath.Join(tempDir, "pkg", "impl3")
	if err := os.MkdirAll(impl3Dir, 0755); err != nil {
		t.Fatalf("failed to create impl3 directory: %v", err)
	}

	impl3Content := `package impl3

type IncompleteService struct{}

func (i *IncompleteService) Process() error {
	return nil
}

// Missing GetStatus method
`
	if err := os.WriteFile(filepath.Join(impl3Dir, "incomplete.go"),
		[]byte(impl3Content), 0644); err != nil {
		t.Fatalf("failed to create impl3 file: %v", err)
	}

	// Change to the temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Run the finder
	finder := NewFinder("TestService")

	if err := finder.validateGoModRoot(); err != nil {
		t.Fatalf("validateGoModRoot failed: %v", err)
	}

	if err := finder.loadModulePath(); err != nil {
		t.Fatalf("loadModulePath failed: %v", err)
	}

	relInterfaceFile, _ := filepath.Rel(tempDir, interfaceFile)
	if err := finder.parseInterface(relInterfaceFile); err != nil {
		t.Fatalf("parseInterface failed: %v", err)
	}

	if err := finder.scanDirectory("pkg"); err != nil {
		t.Fatalf("scanDirectory failed: %v", err)
	}

	results := finder.getResults()

	// Verify results
	if len(results) != 2 {
		t.Errorf("expected 2 implementations, got %d", len(results))
		for i, result := range results {
			t.Logf("Result %d: %+v", i, result)
		}
	}

	// Check that we found the correct implementations
	foundWorker := false
	foundProcessor := false

	for _, result := range results {
		switch result.Struct {
		case "Worker":
			foundWorker = true
			if result.Package != "impl1" {
				t.Errorf("Worker should be in package impl1, got %s", result.Package)
			}
		case "Processor":
			foundProcessor = true
			if result.Package != "impl2" {
				t.Errorf("Processor should be in package impl2, got %s", result.Package)
			}
		default:
			t.Errorf("unexpected struct found: %s", result.Struct)
		}
	}

	if !foundWorker {
		t.Error("Worker implementation not found")
	}

	if !foundProcessor {
		t.Error("Processor implementation not found")
	}

	// Verify JSON serialization
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Errorf("failed to marshal results to JSON: %v", err)
	}

	var unmarshalled []Implementation
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(unmarshalled) != len(results) {
		t.Errorf("JSON roundtrip failed: expected %d items, got %d",
			len(results), len(unmarshalled))
	}
}

func TestParseInterfaceSpecIntegration(t *testing.T) {
	// Test the complete flow with parseInterfaceSpec
	tempDir := t.TempDir()
	interfaceFile := filepath.Join(tempDir, "interface.go")
	
	if err := os.WriteFile(interfaceFile,
		[]byte("package test\ntype TestInterface interface{}"), 0644); err != nil {
		t.Fatalf("failed to create interface file: %v", err)
	}

	spec := interfaceFile + ":TestInterface"
	file, name, err := parseInterfaceSpec(spec)
	
	if err != nil {
		t.Fatalf("parseInterfaceSpec failed: %v", err)
	}

	if file != interfaceFile {
		t.Errorf("expected file %s, got %s", interfaceFile, file)
	}

	if name != "TestInterface" {
		t.Errorf("expected name TestInterface, got %s", name)
	}
}

func TestRunFinderIntegration(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create go.mod
	goModContent := "module testapp\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"),
		[]byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create interface
	interfaceDir := filepath.Join(tempDir, "internal", "app")
	if err := os.MkdirAll(interfaceDir, 0755); err != nil {
		t.Fatalf("failed to create interface directory: %v", err)
	}

	interfaceContent := `package app

type Server interface {
	Start() error
	Stop() error
}`

	interfaceFile := filepath.Join(interfaceDir, "server.go")
	if err := os.WriteFile(interfaceFile,
		[]byte(interfaceContent), 0644); err != nil {
		t.Fatalf("failed to create interface file: %v", err)
	}

	// Create implementation
	implDir := filepath.Join(tempDir, "pkg", "impl")
	if err := os.MkdirAll(implDir, 0755); err != nil {
		t.Fatalf("failed to create impl directory: %v", err)
	}

	implContent := `package impl

type WebServer struct{}

func (w *WebServer) Start() error { return nil }
func (w *WebServer) Stop() error { return nil }
`

	if err := os.WriteFile(filepath.Join(implDir, "server.go"),
		[]byte(implContent), 0644); err != nil {
		t.Fatalf("failed to create impl file: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Test runFinder function directly
	searchDir := filepath.Join(tempDir, "pkg")
	
	// This should execute successfully and find the implementation
	// We can't easily capture the JSON output, but we can test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runFinder panicked: %v", r)
		}
	}()

	// This will output to stdout, but at least tests the function
	runFinder(interfaceFile, "Server", searchDir)
}

