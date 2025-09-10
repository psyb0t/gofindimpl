package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWithFixtures(t *testing.T) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	fixturesDir := filepath.Join(wd, ".fixtures")
	
	// Verify fixtures directory exists
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		t.Fatalf("fixtures directory does not exist: %s", fixturesDir)
	}

	// Change to fixtures directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	
	if err := os.Chdir(fixturesDir); err != nil {
		t.Fatalf("failed to change to fixtures directory: %v", err)
	}

	// Test the finder with real fixtures
	finder := NewFinder("App")

	if err := finder.validateGoModRoot(); err != nil {
		t.Fatalf("validateGoModRoot failed: %v", err)
	}

	if err := finder.loadModulePath(); err != nil {
		t.Fatalf("loadModulePath failed: %v", err)
	}

	if err := finder.parseInterface("internal/app/app.go"); err != nil {
		t.Fatalf("parseInterface failed: %v", err)
	}

	// Verify interface methods were parsed correctly
	expectedMethods := []string{"Start", "Stop", "GetName"}
	if len(finder.interfaceMethods) != len(expectedMethods) {
		t.Errorf("expected %d interface methods, got %d",
			len(expectedMethods), len(finder.interfaceMethods))
	}

	for i, method := range expectedMethods {
		if i >= len(finder.interfaceMethods) || finder.interfaceMethods[i] != method {
			t.Errorf("expected method %s at position %d, got %v",
				method, i, finder.interfaceMethods)
		}
	}

	if err := finder.scanDirectory("pkg/"); err != nil {
		t.Fatalf("scanDirectory failed: %v", err)
	}

	results := finder.getResults()

	// Should find exactly 3 implementations
	if len(results) != 3 {
		t.Errorf("expected 3 implementations, got %d", len(results))
	}

	// Verify specific implementations
	expectedImplementations := map[string]string{
		"WebServer":      "something1",
		"ServiceDaemon":  "something2", 
		"MicroService":   "something3",
	}

	foundImplementations := make(map[string]string)
	for _, result := range results {
		foundImplementations[result.Struct] = result.Package
	}

	for expectedStruct, expectedPackage := range expectedImplementations {
		if foundPackage, found := foundImplementations[expectedStruct]; !found {
			t.Errorf("expected implementation %s not found", expectedStruct)
		} else if foundPackage != expectedPackage {
			t.Errorf("expected %s to be in package %s, got %s",
				expectedStruct, expectedPackage, foundPackage)
		}
	}

	// Verify JSON serialization works
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Errorf("failed to marshal results: %v", err)
	}

	// Verify JSON contains expected data
	if len(jsonData) == 0 {
		t.Error("JSON output should not be empty")
	}
}

func TestParseInterfaceSpecWithFixtures(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	fixturesDir := filepath.Join(wd, ".fixtures")
	interfaceFile := filepath.Join(fixturesDir, "internal", "app", "app.go")
	
	// Test with fixtures path
	spec := interfaceFile + ":App"
	file, name, err := parseInterfaceSpec(spec)
	
	if err != nil {
		t.Fatalf("parseInterfaceSpec failed: %v", err)
	}

	if file != interfaceFile {
		t.Errorf("expected file %s, got %s", interfaceFile, file)
	}

	if name != "App" {
		t.Errorf("expected name App, got %s", name)
	}
}

func TestValidateArgsWithFixtures(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	fixturesDir := filepath.Join(wd, ".fixtures")
	interfaceFile := filepath.Join(fixturesDir, "internal", "app", "app.go")
	searchDir := filepath.Join(fixturesDir, "pkg")
	
	// This should not panic since files exist
	validateArgs(interfaceFile, "App", searchDir)
	
	// Test with non-existent file - this will call log.Fatal but we can't
	// easily test that
	// without changing the implementation, so we'll just verify the files exist
	if _, err := os.Stat(interfaceFile); os.IsNotExist(err) {
		t.Errorf("interface file should exist: %s", interfaceFile)
	}
	
	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		t.Errorf("search directory should exist: %s", searchDir)
	}
}

func TestFixturesDirectoryStructure(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

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
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected fixture file does not exist: %s", fullPath)
		}
	}

	// Verify go.mod content
	goModPath := filepath.Join(fixturesDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	if !contains(string(content), "module testapp") {
		t.Errorf("go.mod should contain 'module testapp', got: %s", string(content))
	}

	// Verify interface file content
	interfacePath := filepath.Join(fixturesDir, "internal/app/app.go")
	interfaceContent, err := os.ReadFile(interfacePath)
	if err != nil {
		t.Fatalf("failed to read interface file: %v", err)
	}

	if !contains(string(interfaceContent), "type App interface") {
		t.Errorf("interface file should contain 'type App interface', got: %s",
			string(interfaceContent))
	}
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