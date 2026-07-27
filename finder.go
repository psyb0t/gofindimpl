package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Implementation struct {
	Package     string `json:"package"`
	Struct      string `json:"struct"`
	PackagePath string `json:"packagePath"`
}

type Finder struct {
	fset             *token.FileSet
	interfaceName    string
	interfaceMethods []string
	modulePath       string
	results          []Implementation
	config           *types.Config
}

type noopImporter struct{}

func (imp *noopImporter) Import(path string) (*types.Package, error) {
	// Return a minimal package for any import
	// This allows type checking to proceed without actually resolving imports
	return types.NewPackage(path, path), nil
}

func NewFinder(interfaceName string) *Finder {
	fset := token.NewFileSet()
	config := &types.Config{
		Importer: &noopImporter{},
		Error: func(_ error) {
			// Ignore type checking errors for incomplete packages
		},
	}

	return &Finder{
		fset:          fset,
		interfaceName: interfaceName,
		results:       make([]Implementation, 0),
		config:        config,
	}
}

func (f *Finder) validateGoModRoot() error {
	if _, err := os.Stat("./go.mod"); os.IsNotExist(err) {
		return ErrGoModNotFound
	}

	return nil
}

func (f *Finder) loadModulePath() error {
	content, err := os.ReadFile("./go.mod")
	if err != nil {
		return fmt.Errorf(
			"failed to read go.mod: %w",
			err,
		)
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			f.modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))

			return nil
		}
	}

	return ErrNoModuleDeclaration
}

func (f *Finder) parseInterface(filePath string) error {
	file, err := parser.ParseFile(f.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf(
			"failed to parse interface file %s: %w",
			filePath,
			err,
		)
	}

	found := false

	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		if ts, ok := n.(*ast.TypeSpec); ok {
			if ts.Name.Name == f.interfaceName {
				if iface, ok := ts.Type.(*ast.InterfaceType); ok {
					f.interfaceMethods = f.getInterfaceMethods(iface)
					found = true

					return false
				}
			}
		}

		return true
	})

	if !found {
		return fmt.Errorf("%w '%s' in %s",
			ErrInterfaceNotFound, f.interfaceName, filePath)
	}

	return nil
}

func (f *Finder) getInterfaceMethods(iface *ast.InterfaceType) []string {
	var methods []string

	for _, method := range iface.Methods.List {
		if len(method.Names) > 0 {
			methods = append(methods, method.Names[0].Name)
		}
	}

	return methods
}

func (f *Finder) scanDirectory(searchDir string) error {
	slog.Debug("starting scan", "dir", searchDir)

	err := filepath.Walk(
		searchDir,
		func(
			path string,
			info os.FileInfo,
			err error,
		) error {
			if err != nil {
				slog.Debug("walk error", "path", path, "err", err)

				return err
			}

			slog.Debug("walking path",
				"path", path,
				"is_dir", info.IsDir(),
				"name", info.Name(),
			)

			if !info.IsDir() || strings.HasPrefix(info.Name(), ".") {
				return nil
			}

			if info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}

			slog.Debug("analyzing directory", "dir", path)
			f.analyzeDirectory(path)

			return nil
		})
	if err != nil {
		return fmt.Errorf(
			"failed to scan directory: %w",
			err,
		)
	}

	return nil
}

func (f *Finder) analyzeDirectory(dirPath string) {
	slog.Debug("analyzing directory", "dir", dirPath)

	files, err := f.parsePackageFiles(dirPath)
	if err != nil {
		slog.Debug("error parsing files", "dir", dirPath, "err", err)

		return
	}

	if len(files) == 0 {
		slog.Debug("no files found", "dir", dirPath)

		return
	}

	slog.Debug("found files", "dir", dirPath, "count", len(files))

	pkg, err := f.typeCheckPackage(files)
	if err != nil {
		slog.Debug("type check failed", "dir", dirPath, "err", err)

		return
	}

	slog.Debug("type-checked package", "package", pkg.Name())
	f.findImplementationsInTypedPackage(dirPath, pkg)
}

func (f *Finder) parsePackageFiles(dirPath string) ([]*ast.File, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read directory: %w",
			err,
		)
	}

	files := make([]*ast.File, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())

		file, err := parser.ParseFile(f.fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		files = append(files, file)
	}

	return files, nil
}

func (f *Finder) typeCheckPackage(files []*ast.File) (*types.Package, error) {
	if len(files) == 0 {
		return nil, ErrNoFilesToTypeCheck
	}

	pkgName := files[0].Name.Name

	pkg, err := f.config.Check(pkgName, f.fset, files, nil)
	if err != nil {
		// Try to continue even if type checking fails
		slog.Debug("type checking had errors, continuing", "err", err)
	}

	if pkg == nil {
		return nil, fmt.Errorf(
			"type checking failed completely: %w",
			err,
		)
	}

	return pkg, nil
}

func (f *Finder) findImplementationsInTypedPackage(
	dirPath string, pkg *types.Package,
) {
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		f.processTypeInScope(obj, dirPath, pkg)
	}
}

func (f *Finder) typeImplementsInterface(namedType *types.Named) bool {
	if len(f.interfaceMethods) == 0 {
		return false
	}

	// Check both value and pointer method sets
	valueMethodSet := types.NewMethodSet(namedType)
	pointerMethodSet := types.NewMethodSet(types.NewPointer(namedType))

	foundMethods := make(map[string]bool)

	// Add methods from both sets
	for method := range valueMethodSet.Methods() {
		foundMethods[method.Obj().Name()] = true
	}

	for method := range pointerMethodSet.Methods() {
		foundMethods[method.Obj().Name()] = true
	}

	for _, requiredMethod := range f.interfaceMethods {
		if !foundMethods[requiredMethod] {
			return false
		}
	}

	return true
}

func (f *Finder) getResults() []Implementation {
	return f.results
}
