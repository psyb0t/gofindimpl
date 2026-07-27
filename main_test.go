package main

import (
	"flag"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInterfaceSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file, name, err := parseInterfaceSpec(tc.spec)

			if tc.expectedError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedFile, file)
			assert.Equal(t, tc.expectedName, name)
		})
	}
}

func TestConfigureLogging(t *testing.T) {
	// not parallel: configureLogging reconfigures the global default slog
	// handler via slogconf.SetHandlers

	testCases := []struct {
		name      string
		debug     bool
		wantLevel slog.Level
	}{
		{
			name:      "debug enabled",
			debug:     true,
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "debug disabled",
			debug:     false,
			wantLevel: slog.LevelError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// not parallel: shares the global logger state mutated by
			// configureLogging

			assert.Equal(t, tc.wantLevel, logLevel(tc.debug))

			configureLogging(tc.debug)
		})
	}
}

func TestSetupUsage(t *testing.T) {
	// not parallel: mutates os.Args and the global flag.Usage

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	os.Args = []string{"testprog"}

	setupUsage()

	require.NotNil(t, flag.Usage, "flag.Usage should be set after setupUsage()")

	assert.NotPanics(t, func() {
		// This would normally print to stderr, but we just want to test it runs.
		flag.Usage()
	})
}

func TestRunFinder(t *testing.T) {
	// not parallel: subtests call os.Chdir, which mutates global process state

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWd))
	})

	testCases := []struct {
		name          string
		interfaceFile string
		interfaceName string
		searchDir     string
		setup         func(t *testing.T) string // returns temp dir
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
				err := os.WriteFile(tempDir+"/test.go", []byte("package main"), 0o644)
				require.NoError(t, err, "failed to create test file")
				require.NoError(t, os.Chdir(tempDir))

				return tempDir
			},
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
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0o644)
				require.NoError(t, err, "failed to create test file")
				require.NoError(t, os.Chdir(tempDir))
				// Don't create go.mod - this will trigger the error

				return tempDir
			},
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
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0o644)
				require.NoError(t, err, "failed to create test file")
				// Create go.mod without module declaration
				goModContent := `go 1.21
require example.com/test v1.0.0`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0o644)
				require.NoError(t, err, "failed to create go.mod")
				require.NoError(t, os.Chdir(tempDir))

				return tempDir
			},
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
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0o644)
				require.NoError(t, err, "failed to create test file")
				// Create valid go.mod
				goModContent := `module test.com/example
go 1.21`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0o644)
				require.NoError(t, err, "failed to create go.mod")
				require.NoError(t, os.Chdir(tempDir))

				return tempDir
			},
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
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0o644)
				require.NoError(t, err, "failed to create test file")
				// Create valid go.mod
				goModContent := `module test.com/example
go 1.21`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0o644)
				require.NoError(t, err, "failed to create go.mod")
				require.NoError(t, os.Chdir(tempDir))

				return tempDir
			},
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
				err := os.WriteFile(tempDir+"/test.go", []byte(interfaceContent), 0o644)
				require.NoError(t, err, "failed to create test file")
				// Create valid go.mod
				goModContent := `module test.com/example
go 1.21`
				err = os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0o644)
				require.NoError(t, err, "failed to create go.mod")
				// Create an implementation
				implContent := `package main
type TestStruct struct{}
func (t *TestStruct) Test() error { return nil }`
				err = os.WriteFile(tempDir+"/impl.go", []byte(implContent), 0o644)
				require.NoError(t, err, "failed to create impl file")
				require.NoError(t, os.Chdir(tempDir))

				return tempDir
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// not parallel: setup calls os.Chdir, which mutates global
			// process state

			tc.setup(t)

			err := runFinder(tc.interfaceFile, tc.interfaceName, tc.searchDir)

			if tc.expectedError {
				require.Error(t, err)

				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}

				return
			}

			require.NoError(t, err)
		})
	}
}
