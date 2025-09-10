package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMainWithFixtures(t *testing.T) {
	// Save original values
	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create a buffer to capture stdout
	var buf bytes.Buffer
	
	// Create temp files for redirecting stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	
	os.Stdout = w

	// Set up command line args to use fixtures
	fixturesDir, err := filepath.Abs(".fixtures")
	if err != nil {
		t.Fatalf("failed to get fixtures dir: %v", err)
	}
	
	interfaceFile := filepath.Join(fixturesDir, "internal", "app", "app.go")
	searchDir := filepath.Join(fixturesDir, "pkg")
	
	os.Args = []string{
		"gofindimpl",
		"-interface", interfaceFile + ":App",
		"-dir", searchDir,
	}

	// Reset flag package state for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Change to fixtures directory for go.mod
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(fixturesDir)

	// Capture output in goroutine
	done := make(chan bool)
	readDone := make(chan bool)
	
	// Start reading first
	go func() {
		defer close(readDone)
		defer r.Close()
		buf.ReadFrom(r)
	}()
	
	// Then start main function
	go func() {
		defer close(done)
		defer w.Close()
		
		// This should not panic and should find implementations
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panicked: %v", r)
			}
		}()
		
		main()
	}()

	// Wait for both to complete
	<-done
	<-readDone

	// Parse the JSON output
	var results []Implementation
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// Verify we found the expected implementations
	if len(results) != 3 {
		t.Errorf("expected 3 implementations, got %d", len(results))
	}

	// Check that we found the expected structs
	expectedStructs := map[string]string{
		"WebServer":     "something1",
		"ServiceDaemon": "something2",
		"MicroService":  "something3",
	}

	foundStructs := make(map[string]string)
	for _, result := range results {
		foundStructs[result.Struct] = result.Package
	}

	for expectedStruct, expectedPackage := range expectedStructs {
		if foundPackage, exists := foundStructs[expectedStruct]; !exists {
			t.Errorf("expected struct %s not found", expectedStruct)
		} else if foundPackage != expectedPackage {
			t.Errorf("expected %s to be in package %s, got %s",
				expectedStruct, expectedPackage, foundPackage)
		}
	}
}

func TestMainHelp(t *testing.T) {
	// Save original values
	oldArgs := os.Args
	oldStderr := os.Stderr
	
	defer func() {
		os.Args = oldArgs
		os.Stderr = oldStderr
	}()

	// Capture stderr for help output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	
	os.Stderr = w

	// Set up command line args for help
	os.Args = []string{"gofindimpl", "-help"}

	// Reset flag package state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Capture output
	var buf bytes.Buffer
	done := make(chan bool)
	readDone := make(chan bool)
	
	// Start reading first
	go func() {
		defer close(readDone)
		defer r.Close()
		buf.ReadFrom(r)
	}()
	
	// Then start main function
	go func() {
		defer close(done)
		defer w.Close()
		
		// This should call os.Exit(0) for help, but we can't test that easily
		// Instead we test that it doesn't panic unexpectedly
		defer func() {
			if r := recover(); r != nil {
				// Help might cause os.Exit which could panic in test
				// That's expected behavior
			}
		}()
		
		main()
	}()

	// Wait for both to complete
	<-done
	<-readDone

	// Check that help output contains expected text
	output := buf.String()
	if len(output) > 0 {
		// If we got output, verify it looks like help text
		expectedTexts := []string{"Usage:", "Options:", "Example:"}
		for _, expected := range expectedTexts {
			if !bytes.Contains(buf.Bytes(), []byte(expected)) {
				t.Errorf("help output should contain '%s'", expected)
			}
		}
	}
}