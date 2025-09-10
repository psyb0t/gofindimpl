package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestIsStructType(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("test", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := pkg.Scope().Lookup(tt.typeName)
			if obj == nil {
				t.Fatalf("type %s not found", tt.typeName)
			}

			typeName, ok := obj.(*types.TypeName)
			if !ok {
				t.Fatalf("%s is not a type name", tt.typeName)
			}

			namedType, ok := typeName.Type().(*types.Named)
			if !ok {
				t.Fatalf("%s is not a named type", tt.typeName)
			}

			result := finder.isStructType(namedType)
			if result != tt.expectTrue {
				t.Errorf("expected %v for %s, got %v", tt.expectTrue, tt.typeName, result)
			}
		})
	}
}

func TestCreateImplementation(t *testing.T) {
	finder := NewFinder("TestInterface")
	finder.modulePath = "github.com/test/repo"

	src := `
package testpkg

type TestStruct struct{}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("testpkg", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	obj := pkg.Scope().Lookup("TestStruct")
	typeName := obj.(*types.TypeName)

	impl := finder.createImplementation("./pkg/testpkg", pkg, typeName)

	if impl.Package != "testpkg" {
		t.Errorf("expected package 'testpkg', got '%s'", impl.Package)
	}

	if impl.Struct != "TestStruct" {
		t.Errorf("expected struct 'TestStruct', got '%s'", impl.Struct)
	}

	expectedPath := "github.com/test/repo/pkg/testpkg"
	if impl.PackagePath != expectedPath {
		t.Errorf("expected package path '%s', got '%s'",
			expectedPath, impl.PackagePath)
	}
}

func TestProcessTypeInScope(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("testpkg", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	// Test with complete implementation
	obj := pkg.Scope().Lookup("TestStruct")
	finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

	if len(finder.results) != 1 {
		t.Errorf("expected 1 result for complete implementation, got %d",
			len(finder.results))
	}

	// Reset and test with incomplete implementation
	finder.results = []Implementation{}
	obj = pkg.Scope().Lookup("IncompleteStruct")
	finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

	if len(finder.results) != 0 {
		t.Errorf("expected 0 results for incomplete implementation, got %d",
			len(finder.results))
	}

	// Reset and test with interface (not a struct)
	finder.results = []Implementation{}
	obj = pkg.Scope().Lookup("NotAStruct")
	finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

	if len(finder.results) != 0 {
		t.Errorf("expected 0 results for interface type, got %d", len(finder.results))
	}
}

func TestProcessTypeInScopeEdgeCases(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	config := &types.Config{
		Error: func(err error) {},
	}

	pkg, err := config.Check("testpkg", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	// Test with non-TypeName objects (variable, constant, function)
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder.results = []Implementation{} // Reset
			obj := pkg.Scope().Lookup(tt.objName)
			if obj == nil {
				t.Fatalf("object %s not found", tt.objName)
			}

			finder.processTypeInScope(obj, "./pkg/testpkg", pkg)

			if len(finder.results) != tt.expected {
				t.Errorf("expected %d results for %s, got %d",
					tt.expected, tt.name, len(finder.results))
			}
		})
	}
}