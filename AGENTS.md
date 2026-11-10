# gofindimpl

Instructions for coding agents (Codex, Claude Code and the like) and the people who drive them when changing this repository. Users of the tool want the [README](README.md) and the documentation site instead.

## Agent setup

`CLAUDE.md` is a committed symlink to `AGENTS.md`, and `.claude` is a committed symlink to `.agents`, so Claude Code reads the same instructions, rules and skills as Codex. Claude's Edit and Write tools refuse to write through a symlink, so edit `AGENTS.md` and `.agents/rules/*.md` directly.

## What it is

A Go command line tool, installed with `go install github.com/psyb0t/gofindimpl@latest`. Given `-interface file.go:InterfaceName` and `-dir`, it prints a JSON array of the structs under `-dir` that have a method for every method name the interface declares. It must run from the root of the module being searched, because it reads `./go.mod` for the module path. It ships no library API, no Docker image and no release binaries.

## Architecture

One `main` package at the repository root. Data flow in `runFinder` (`main.go`):

1. `parseInterfaceSpec` splits the spec on `:` into a file and a name.
2. `validateArgs` (`validation.go`) checks that the file and `-dir` exist.
3. `Finder.validateGoModRoot` and `Finder.loadModulePath` (`finder.go`) require `./go.mod` and read its `module` line.
4. `Finder.parseInterface` parses the interface file with `go/parser` and stores the interface's method names. Embedded interfaces contribute nothing.
5. `Finder.scanDirectory` walks `-dir` with `filepath.Walk`. It skips `vendor` and `node_modules` subtrees and does not analyze directories whose name starts with `.` (it still descends into them, and `.` itself counts).
6. `Finder.analyzeDirectory` parses the non-`_test.go` files, type-checks them with `go/types` and a `noopImporter` that returns an empty package for every import, and ignores type errors.
7. `Finder.processTypeInScope` (`analysis.go`) keeps package-scope named struct types whose value or pointer method set contains every interface method name (`typeImplementsInterface`). Names only, signatures are not compared.
8. The results are marshalled with two-space indentation and written to stdout. Logs go to stderr.

## Stack

- Go 1.26 (`go 1.26` in `go.mod`, `go_version: "1.26"` in CI).
- Standard library `go/ast`, `go/parser`, `go/token`, `go/types` and `log/slog`.
- `github.com/psyb0t/slogging/slogconf` to install the stderr `slog` handler.
- `github.com/stretchr/testify` for tests.
- golangci-lint v2 as a `tool` dependency in `go.mod`, run with `go tool golangci-lint`, configured in `.golangci.yml`.

## Layout

| Path | Holds |
| --- | --- |
| `main.go` | Flags, usage text, `configureLogging`, `parseInterfaceSpec`, `runFinder` and the JSON output. |
| `finder.go` | `Finder`, `Implementation`, `noopImporter`, the `go.mod` checks, interface parsing, the walk, type checking and `typeImplementsInterface`. |
| `analysis.go` | `processTypeInScope`, `isStructType`, `createImplementation` (builds `packagePath`). |
| `validation.go` | `validateArgs`. |
| `errors.go` | Every sentinel error. |
| `*_test.go` | Unit and integration tests, all in `package main`. |
| `.fixtures/` | The `testapp` module the tests run against: an `App` interface in `internal/app` and four packages under `pkg/`. |
| `docs/` | The documentation site, published by GitHub Pages from `main` at https://psyb0t.github.io/gofindimpl/ (Jekyll, just-the-docs theme). |
| `Makefile` | Every dev command. |
| `.github/workflows/pipeline.yml` | CI: lint, tests and govulncheck on every branch push; release and badges on `v*` tags. |
| `.agents/rules/` | The rules below. |

## Commands

You need Go 1.26 and Make. `build`, `build-race` and `docs` also need Docker.

```bash
make help            # every target
make dep             # go mod tidy
make lint            # go fix -diff check, then golangci-lint
make lint-fix        # apply go fix and golangci-lint --fix
make test            # go test -race ./...
make test-coverage   # tests with coverage, fails below 90%, writes coverage-percent.txt
make build           # static binary in build/ via a golang container
make docs            # build the docs site into .testing/docs-site
```

`pre-commit.sh` runs `make lint && make test-coverage`, the same pair CI runs.

## Rules

Read `.agents/rules/` before changing anything. One file per topic: Go code, errors and logging, testing, documentation, CI.
