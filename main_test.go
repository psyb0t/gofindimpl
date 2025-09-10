package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestParseInterfaceSpec(t *testing.T) {
	tests := []struct {
		name          string
		spec          string
		expectedFile  string
		expectedName  string
		expectedError bool
	}{
		{
			name:          "valid spec",
			spec:          "internal/app/server.go:Server",
			expectedFile:  "internal/app/server.go",
			expectedName:  "Server",
			expectedError: false,
		},
		{
			name:          "valid spec with spaces",
			spec:          " internal/app/server.go : Server ",
			expectedFile:  "internal/app/server.go",
			expectedName:  "Server",
			expectedError: false,
		},
		{
			name:          "empty spec",
			spec:          "",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
		{
			name:          "missing colon",
			spec:          "internal/app/server.go",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
		{
			name:          "too many colons",
			spec:          "internal/app/server.go:Server:Extra",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
		{
			name:          "empty file path",
			spec:          ":Server",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
		{
			name:          "empty interface name",
			spec:          "internal/app/server.go:",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
		{
			name:          "only spaces in file path",
			spec:          "   :Server",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
		{
			name:          "only spaces in interface name",
			spec:          "internal/app/server.go:   ",
			expectedFile:  "",
			expectedName:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, name, err := parseInterfaceSpec(tt.spec)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if file != tt.expectedFile {
					t.Errorf("expected file '%s', got '%s'", tt.expectedFile, file)
				}
				if name != tt.expectedName {
					t.Errorf("expected name '%s', got '%s'", tt.expectedName, name)
				}
			}
		})
	}
}

func TestConfigureLogging(t *testing.T) {
	tests := []struct {
		name          string
		debug         bool
		expectedLevel logrus.Level
	}{
		{
			name:          "debug enabled",
			debug:         true,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "debug disabled",
			debug:         false,
			expectedLevel: logrus.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configureLogging(tt.debug)

			if logrus.GetLevel() != tt.expectedLevel {
				t.Errorf("expected log level %v, got %v",
					tt.expectedLevel, logrus.GetLevel())
			}

			// Check formatter type
			formatter := logrus.StandardLogger().Formatter
			if _, ok := formatter.(*logrus.TextFormatter); !ok {
				t.Error("expected TextFormatter")
			}
		})
	}
}

func TestSetupUsage(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"testprog"}

	setupUsage()

	if flag.Usage == nil {
		t.Error("flag.Usage should be set after setupUsage()")
	}

	// Test that the usage function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("setupUsage() created a Usage function that panics: %v", r)
		}
	}()

	// This would normally print to stderr, but we just want to test it runs
	flag.Usage()
}

func TestRunFinder(t *testing.T) {
	// Save original working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	tests := []struct {
		name          string
		interfaceFile string
		interfaceName string
		searchDir     string
		setup         func(t *testing.T) string // returns temp dir
		cleanup       func(string)
		expectedError bool
		errorContains string
	}{
		{
			name:          "validateArgs error - non-existent interface file",
			interfaceFile: "/nonexistent/interface.go",
			interfaceName: "TestInterface",
			searchDir:     ".",
			setup: func(t *testing.T) string {
				return ""
			},
			cleanup:       func(string) {},
			expectedError: true,
			errorContains: "interface file does not exist",
		},
		{
			name:          "validateArgs error - non-existent search dir",
			interfaceFile: "test.go",
			interfaceName: "TestInterface",
			searchDir:     "/nonexistent/dir",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create a test interface file
				err := os.WriteFile(tempDir+"/test.go", []byte("package main"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				os.Chdir(tempDir)
				return tempDir
			},
			cleanup:       func(string) {},
			expectedError: true,
			errorContains: "search directory does not exist",
		},
		{
			name:          "validateGoModRoot error - no go.mod",
			interfaceFile: "test.go",
			interfaceName: "TestInterface", 
			searchDir:     ".",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create a test interface file with valid interface
				interfaceContent := `package main
type TestInterface interface {
	Test() error
}`
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				os.Chdir(tempDir)
				// Don't create go.mod - this will trigger the error
				return tempDir
			},
			cleanup:       func(string) {},
			expectedError: true,
			errorContains: "go.mod not found in current directory",
		},
		{
			name:          "loadModulePath error - malformed go.mod",
			interfaceFile: "test.go",
			interfaceName: "TestInterface",
			searchDir:     ".",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create a test interface file
				interfaceContent := `package main
type TestInterface interface {
	Test() error
}`
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				// Create go.mod without module declaration
				goModContent := `go 1.21
require example.com/test v1.0.0`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0644)
				if err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				os.Chdir(tempDir)
				return tempDir
			},
			cleanup:       func(string) {},
			expectedError: true,
			errorContains: "no module declaration found",
		},
		{
			name:          "parseInterface error - interface not found",
			interfaceFile: "test.go",
			interfaceName: "NonExistentInterface",
			searchDir:     ".",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create a test file without the interface we're looking for
				interfaceContent := `package main
type SomeOtherInterface interface {
	Test() error
}`
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				// Create valid go.mod
				goModContent := `module test.com/example
go 1.21`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0644)
				if err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				os.Chdir(tempDir)
				return tempDir
			},
			cleanup:       func(string) {},
			expectedError: true,
			errorContains: "interface not found",
		},
		{
			name:          "parseInterface error - malformed Go file",
			interfaceFile: "test.go",
			interfaceName: "TestInterface",
			searchDir:     ".",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create malformed Go file
				interfaceContent := `package main
invalid go syntax here @@#$#@$`
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				// Create valid go.mod
				goModContent := `module test.com/example
go 1.21`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0644)
				if err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				os.Chdir(tempDir)
				return tempDir
			},
			cleanup:       func(string) {},
			expectedError: true,
			errorContains: "failed to parse interface file",
		},
		{
			name:          "successful run",
			interfaceFile: "test.go",
			interfaceName: "TestInterface",
			searchDir:     ".",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create a valid interface file
				interfaceContent := `package main
type TestInterface interface {
	Test() error
}`
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				// Create valid go.mod
				goModContent := `module test.com/example
go 1.21`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0644)
				if err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				// Create an implementation
				implContent := `package main
type TestStruct struct{}
func (t *TestStruct) Test() error { return nil }`
				err = os.WriteFile(tempDir+"/impl.go", []byte(implContent), 0644)
				if err != nil {
					t.Fatalf("failed to create impl file: %v", err)
				}
				os.Chdir(tempDir)
				return tempDir
			},
			cleanup:       func(string) {},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setup(t)
			defer tt.cleanup(tempDir)

			err := runFinder(tt.interfaceFile, tt.interfaceName, tt.searchDir)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}