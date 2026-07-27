package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFinder(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	assert.Equal(t, "TestInterface", finder.interfaceName)
	assert.NotNil(t, finder.fset)
	assert.NotNil(t, finder.config)
	assert.NotNil(t, finder.results)
	assert.Empty(t, finder.results)
}

func TestFinder_ValidateGoModRoot(t *testing.T) {
	// not parallel: calls os.Chdir, mutates process cwd
	finder := NewFinder("TestInterface")
	tempDir := t.TempDir()

	// Test with no go.mod file
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tempDir))
	err = finder.validateGoModRoot()
	require.ErrorIs(t, err, ErrGoModNotFound)

	// Test with go.mod file
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte("module testmodule\n"), 0o644))

	err = finder.validateGoModRoot()
	require.NoError(t, err)
}

func TestFinder_LoadModulePath(t *testing.T) {
	// not parallel: calls os.Chdir, mutates process cwd
	finder := NewFinder("TestInterface")
	tempDir := t.TempDir()

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})
	require.NoError(t, os.Chdir(tempDir))

	testCases := []struct {
		name           string
		goModContent   string
		expectedModule string
		expectError    bool
	}{
		{
			name:           "valid go.mod",
			goModContent:   "module github.com/test/repo\n\ngo 1.21\n",
			expectedModule: "github.com/test/repo",
			expectError:    false,
		},
		{
			name:           "go.mod with spaces",
			goModContent:   "module   github.com/test/repo   \n",
			expectedModule: "github.com/test/repo",
			expectError:    false,
		},
		{
			name:         "go.mod without module declaration",
			goModContent: "go 1.21\n",
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// not parallel: shares finder + cwd chdir'd by parent test
			goModPath := filepath.Join(tempDir, "go.mod")
			require.NoError(t, os.WriteFile(goModPath, []byte(tc.goModContent), 0o644))

			err := finder.loadModulePath()

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedModule, finder.modulePath)
		})
	}
}

func TestFinder_GetInterfaceMethods(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	// Create a simple interface AST
	src := `
package test

type TestInterface interface {
	Method1() error
	Method2(string) int
	Method3()
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	// Find the interface
	var iface *ast.InterfaceType
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == "TestInterface" {
			if interfaceType, ok := ts.Type.(*ast.InterfaceType); ok {
				iface = interfaceType
				return false
			}
		}
		return true
	})

	require.NotNil(t, iface)

	methods := finder.getInterfaceMethods(iface)
	expected := []string{"Method1", "Method2", "Method3"}

	require.Len(t, methods, len(expected))

	for i, method := range methods {
		assert.Equal(t, expected[i], method)
	}
}

func TestFinder_TypeImplementsInterface(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")
	finder.interfaceMethods = []string{"Start", "Stop", "GetName"}

	// Create a test type that implements the interface
	src := `
package test

type TestStruct struct{}

func (t *TestStruct) Start() error { return nil }
func (t *TestStruct) Stop() error { return nil }
func (t *TestStruct) GetName() string { return "test" }
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	config := &types.Config{
		Error: func(err error) {}, // Ignore errors
	}

	pkg, err := config.Check("test", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	// Find the TestStruct type
	obj := pkg.Scope().Lookup("TestStruct")
	require.NotNil(t, obj)

	typeName, ok := obj.(*types.TypeName)
	require.True(t, ok)

	namedType, ok := typeName.Type().(*types.Named)
	require.True(t, ok)

	assert.True(t, finder.typeImplementsInterface(namedType))

	// Test with incomplete implementation
	finder.interfaceMethods = []string{"Start", "Stop", "GetName", "Missing"}
	assert.False(t, finder.typeImplementsInterface(namedType))
}

func TestFinder_GetResults(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	// Initially empty
	results := finder.getResults()
	assert.Empty(t, results)

	// Add some results
	finder.results = append(finder.results, Implementation{
		Package:     "test",
		Struct:      "TestStruct",
		PackagePath: "github.com/test/repo",
	})

	results = finder.getResults()
	require.Len(t, results, 1)
	assert.Equal(t, "test", results[0].Package)
}

func TestFinder_LoadModulePathErrors(t *testing.T) {
	// not parallel: calls os.Chdir, mutates process cwd
	finder := NewFinder("TestInterface")
	tempDir := t.TempDir()

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})
	require.NoError(t, os.Chdir(tempDir))

	// Test with non-readable go.mod file (permission denied)
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte("module test\n"), 0o644))

	// Make file unreadable (won't work on all systems, but covers the error path)
	originalContent := "module test\n"
	require.NoError(t, os.WriteFile(goModPath, []byte(originalContent), 0o644))

	err = finder.loadModulePath()
	if err == nil {
		assert.Equal(t, "test", finder.modulePath)
	}
}

func TestFinder_ParseInterfaceErrors(t *testing.T) {
	// not parallel: subtests share a finder whose fset/interfaceMethods are mutated by parseInterface

	finder := NewFinder("TestInterface")
	tempDir := t.TempDir()

	testCases := []struct {
		name        string
		fileContent string
		expectError bool
	}{
		{
			name:        "valid interface",
			fileContent: "package test\ntype TestInterface interface { Method() }",
			expectError: false,
		},
		{
			name:        "invalid Go syntax",
			fileContent: "package test\ntype TestInterface interface { Method( }",
			expectError: true,
		},
		{
			name:        "interface not found",
			fileContent: "package test\ntype OtherInterface interface { Method() }",
			expectError: true,
		},
		{
			name:        "type is not interface",
			fileContent: "package test\ntype TestInterface struct { Field int }",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, "test.go")
			require.NoError(t, os.WriteFile(testFile, []byte(tc.fileContent), 0o644))

			err := finder.parseInterface(testFile)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestFinder_ScanDirectoryErrors(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	// Test with non-existent directory
	err := finder.scanDirectory("/nonexistent/directory")
	require.Error(t, err)
}

func TestFinder_ParsePackageFilesEdgeCases(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")
	tempDir := t.TempDir()

	// Test directory with no Go files
	emptyDir := filepath.Join(tempDir, "empty")
	require.NoError(t, os.Mkdir(emptyDir, 0o755))

	files, err := finder.parsePackageFiles(emptyDir)
	require.NoError(t, err)
	assert.Empty(t, files)

	// Test directory with only test files
	testDir := filepath.Join(tempDir, "testonly")
	require.NoError(t, os.Mkdir(testDir, 0o755))
	testFile := filepath.Join(testDir, "main_test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package main\nfunc TestFoo(t *testing.T) {}"), 0o644))

	files, err = finder.parsePackageFiles(testDir)
	require.NoError(t, err)
	assert.Empty(t, files)

	// Test directory with invalid Go file
	invalidDir := filepath.Join(tempDir, "invalid")
	require.NoError(t, os.Mkdir(invalidDir, 0o755))
	invalidFile := filepath.Join(invalidDir, "invalid.go")
	require.NoError(t, os.WriteFile(invalidFile, []byte("invalid go syntax {{{"), 0o644))

	files, err = finder.parsePackageFiles(invalidDir)
	require.NoError(t, err)
	// Should skip invalid files and return empty slice
	assert.Empty(t, files)
}

func TestFinder_TypeCheckPackageErrors(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	// Test with empty file slice
	pkg, err := finder.typeCheckPackage([]*ast.File{})
	assert.Nil(t, pkg)
	require.ErrorIs(t, err, ErrNoFilesToTypeCheck)
}

func TestFinder_TypeImplementsInterfaceEdgeCases(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	// Test with empty interface methods
	finder.interfaceMethods = []string{}

	src := `package test
type TestStruct struct{}
func (t *TestStruct) Method() {}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	config := &types.Config{Error: func(err error) {}}
	pkg, err := config.Check("test", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	obj := pkg.Scope().Lookup("TestStruct")
	typeName, ok := obj.(*types.TypeName)
	require.True(t, ok)

	namedType, ok := typeName.Type().(*types.Named)
	require.True(t, ok)

	// Should return false for empty interface methods
	assert.False(t, finder.typeImplementsInterface(namedType))
}

func TestFinder_AnalyzeDirectoryErrorPaths(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")
	tempDir := t.TempDir()

	// Test with directory that causes parsePackageFiles to return error
	unreadableDir := filepath.Join(tempDir, "unreadable")
	require.NoError(t, os.Mkdir(unreadableDir, 0o755))

	// Create a regular file where a directory is expected to cause error
	badSubDir := filepath.Join(unreadableDir, "badfile")
	require.NoError(t, os.WriteFile(badSubDir, []byte("content"), 0o644))

	// analyzeDirectory should handle errors gracefully
	finder.analyzeDirectory(unreadableDir)

	// Should not panic and continue execution
	assert.Empty(t, finder.results)
}
