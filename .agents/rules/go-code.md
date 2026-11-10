# Go code

Prefix: GO.

- **GO1:** Everything stays in one `main` package at the repository root. There is no library API to keep stable, but the CLI is the API: the flag names, the `file.go:InterfaceName` spec and the JSON field names (`package`, `struct`, `packagePath` on `Implementation` in `finder.go`) change only with a CHANGELOG entry that says so.
- **GO2:** stdout carries the JSON result and nothing else. `runFinder` in `main.go` is the only writer. Never `fmt.Print` to stdout or add a log handler that writes there; `configureLogging` uses one stderr handler on purpose.
- **GO3:** The tool reads source with `go/parser` and `go/types` and the `noopImporter` in `finder.go`. It does not run `go list`, `go/packages` or the `go` binary, and it does not resolve imports. A change to that is a design change: update `docs/how-it-works.md` and its limits in the same change.
- **GO4:** Type errors and unparsable files in the searched tree are skipped (`config.Error` ignores them, `parsePackageFiles` drops files that fail to parse). Failing the whole run because one package is broken is a behaviour change.
- **GO5:** `make lint` must pass with the rules in `.golangci.yml`, which enables most linters (`wsl_v5`, `err113`, `wrapcheck`, `mnd`, `gochecknoglobals`, `funlen`, `lll` and more). Fix the code rather than adding `//nolint` or disabling a linter. Test files are excluded from lint.
- **GO6:** Imports outside the standard library are limited by the `depguard` allowlist in `.golangci.yml`: `github.com/stretchr/testify`, `github.com/psyb0t/...` and `github.com/sirupsen/logrus`. A new dependency needs an allowlist entry and a reason.
