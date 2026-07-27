package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsStructType(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")

	src := `
package test

type TestStruct struct {
	Field string
}

type TestInterface interface {
	Method()
}

type TestAlias = string
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("test", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	testCases := []struct {
		name       string
		typeName   string
		expectTrue bool
	}{
		{
			name:       "struct type",
			typeName:   "TestStruct",
			expectTrue: true,
		},
		{
			name:       "interface type",
			typeName:   "TestInterface",
			expectTrue: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			obj := pkg.Scope().Lookup(tc.typeName)
			require.NotNil(t, obj, "type %s not found", tc.typeName)

			typeName, ok := obj.(*types.TypeName)
			require.True(t, ok, "%s is not a type name", tc.typeName)

			namedType, ok := typeName.Type().(*types.Named)
			require.True(t, ok, "%s is not a named type", tc.typeName)

			result := finder.isStructType(namedType)
			assert.Equal(t, tc.expectTrue, result)
		})
	}
}

func TestCreateImplementation(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")
	finder.modulePath = "github.com/test/repo"

	src := `
package testpkg

type TestStruct struct{}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("testpkg", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	obj := pkg.Scope().Lookup("TestStruct")
	typeName := obj.(*types.TypeName)

	impl := finder.createImplementation("./pkg/testpkg", pkg, typeName)

	assert.Equal(t, "testpkg", impl.Package)
	assert.Equal(t, "TestStruct", impl.Struct)

	expectedPath := "github.com/test/repo/pkg/testpkg"
	assert.Equal(t, expectedPath, impl.PackagePath)
}

func TestProcessTypeInScope(t *testing.T) {
	t.Parallel()

	finder := NewFinder("TestInterface")
	finder.interfaceMethods = []string{"Start", "Stop"}
	finder.modulePath = "github.com/test/repo"

	src := `
package testpkg

type TestStruct struct{}

func (t *TestStruct) Start() error { return nil }
func (t *TestStruct) Stop() error { return nil }

type IncompleteStruct struct{}

func (i *IncompleteStruct) Start() error { return nil }
// Missing Stop method

type NotAStruct interface {
	Method()
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("testpkg", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	// Test with complete implementation
	obj := pkg.Scope().Lookup("TestStruct")
	finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

	assert.Len(t, finder.results, 1)

	// Reset and test with incomplete implementation
	finder.results = []Implementation{}
	obj = pkg.Scope().Lookup("IncompleteStruct")
	finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

	assert.Empty(t, finder.results)

	// Reset and test with interface (not a struct)
	finder.results = []Implementation{}
	obj = pkg.Scope().Lookup("NotAStruct")
	finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

	assert.Empty(t, finder.results)
}

func TestProcessTypeInScopeEdgeCases(t *testing.T) {
	// not parallel: subtests share and mutate finder.results, a race under t.Parallel()
	finder := NewFinder("TestInterface")
	finder.interfaceMethods = []string{"Method"}

	src := `
package testpkg

var GlobalVar int = 42
const GlobalConst = "test"

func GlobalFunc() {}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("testpkg", fset, []*ast.File{file}, nil)
	require.NoError(t, err)

	// Test with non-TypeName objects (variable, constant, function)
	testCases := []struct {
		name     string
		objName  string
		expected int
	}{
		{
			name:     "global variable",
			objName:  "GlobalVar",
			expected: 0,
		},
		{
			name:     "global constant",
			objName:  "GlobalConst",
			expected: 0,
		},
		{
			name:     "global function",
			objName:  "GlobalFunc",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// not parallel: mutates shared finder.results from the parent test
			finder.results = []Implementation{} // Reset
			obj := pkg.Scope().Lookup(tc.objName)
			require.NotNil(t, obj, "object %s not found", tc.objName)

			finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

			assert.Len(t, finder.results, tc.expected)
		})
	}
}
