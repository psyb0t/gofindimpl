# gofindimpl 🔍

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/gofindimpl.svg)](https://pkg.go.dev/github.com/psyb0t/gofindimpl)
[![CI](https://github.com/psyb0t/gofindimpl/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/gofindimpl/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/gofindimpl/badges/coverage.svg)](https://github.com/psyb0t/gofindimpl/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/gofindimpl/badges/version.svg)](https://github.com/psyb0t/gofindimpl/tags)
[![license](https://raw.githubusercontent.com/psyb0t/gofindimpl/badges/license.svg)](https://github.com/psyb0t/gofindimpl/blob/main/LICENSE)

Hunt down Go interface implementations like a bloodhound with trust issues.

Tired of grep-ing through thousands of lines trying to figure out which structs actually implement that damn interface? This tool does the heavy lifting so you don't have to suffer through another existential crisis at 3am.

## Installation

```bash
go install github.com/psyb0t/gofindimpl@latest
```

Or skip the install and run it directly:

```bash
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./path/to/file.go:InterfaceName \
  -dir ./search/directory
```

Needs Go 1.26 or newer.

## Usage

Run it from your module root, the directory with `go.mod` in it. Not some anarchist directory, it checks.

```bash
gofindimpl \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/
```

You get JSON on stdout, because XML is for people who hate themselves:

```json
[
  {
    "package": "impl",
    "struct": "WebServer",
    "packagePath": "github.com/yourproject/internal/pkg/impl"
  },
  {
    "package": "mock",
    "struct": "MockServer",
    "packagePath": "github.com/yourproject/internal/pkg/mock"
  }
]
```

Add `-debug` if you're a masochist and want to watch every directory get walked. Logs go to stderr, so the JSON stays clean enough to pipe into `jq`. Every flag, the output fields and the exit codes are on [Command line](https://psyb0t.github.io/gofindimpl/command-line/).

## How it works

It parses the interface file, walks `-dir`, type-checks every package it finds with Go's own type checker and a fake importer, and reports each struct whose value or pointer method set has every method name the interface declares. Test files, `vendor`, `node_modules` and dot-directories don't get analyzed.

Read the fine print before you trust the output. It compares method names, not signatures, it ignores embedded interfaces, it can't see methods promoted from another package, and with the default `-dir .` it skips the package in the module root itself. [How it works](https://psyb0t.github.io/gofindimpl/how-it-works/) has all of it.

## Documentation

The full docs live at **[psyb0t.github.io/gofindimpl](https://psyb0t.github.io/gofindimpl/)**.

| Page | What's in it |
| --- | --- |
| [Command line](https://psyb0t.github.io/gofindimpl/command-line/) | Flags, the `file.go:InterfaceName` spec, JSON output, logs and exit codes |
| [How it works](https://psyb0t.github.io/gofindimpl/how-it-works/) | What gets scanned, what counts as an implementation, and the limits |
| [Troubleshooting](https://psyb0t.github.io/gofindimpl/troubleshooting/) | Every error message, why a search comes back empty, and why it finds too much |
| [Development](https://psyb0t.github.io/gofindimpl/development/) | Make targets, tests, the coverage gate, CI and releases |

## Sponsor

<a href="https://www.buymeacoffee.com/psyb0t"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me a Coffee" height="60" width="217"></a>

## License

MIT, because pkg.go.dev are corporate fuckers who won't index WTFPL packages.
