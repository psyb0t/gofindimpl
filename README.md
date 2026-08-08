# gofindimpl 🔍

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/gofindimpl.svg)](https://pkg.go.dev/github.com/psyb0t/gofindimpl)
[![CI](https://github.com/psyb0t/gofindimpl/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/gofindimpl/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/gofindimpl/badges/coverage.svg)](https://github.com/psyb0t/gofindimpl/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/gofindimpl/badges/version.svg)](https://github.com/psyb0t/gofindimpl/tags)
[![license](https://raw.githubusercontent.com/psyb0t/gofindimpl/badges/license.svg)](LICENSE)

Hunt down Go interface implementations like a bloodhound with trust issues.

Tired of grep-ing through thousands of lines trying to figure out which structs actually implement that damn interface? This tool does the heavy lifting so you don't have to suffer through another existential crisis at 3am.

## Installation 🚀

### Option 1: Install it
```bash
go install github.com/psyb0t/gofindimpl@latest
```

### Option 2: Run directly (no installation needed)
```bash
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./path/to/file.go:InterfaceName \
  -dir ./search/directory
```

### Option 3: Clone and build like it's 2005
```bash
git clone https://github.com/psyb0t/gofindimpl.git
cd gofindimpl
go build -o gofindimpl
```

## Usage 💀

### Basic Hunt (installed)

```bash
gofindimpl \
  -interface ./path/to/file.go:InterfaceName \
  -dir ./search/directory
```

### Basic Hunt (direct run)

```bash
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./path/to/file.go:InterfaceName \
  -dir ./search/directory
```

### Real Example (installed)

```bash
gofindimpl -interface ./internal/app/server.go:Server -dir ./internal/pkg/
```

### Real Example (direct run)

```bash
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/
```

### With Debug Logging (for masochists)

```bash
# Installed
gofindimpl \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/ \
  -debug

# Direct run  
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/ \
  -debug
```

## Output Format 📋

JSON, because XML is for people who hate themselves:

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

## How It Works 🧠

1. **Parse Interface**: Reads the specified Go file and extracts interface methods
2. **Scan Directory**: Recursively walks through Go files (skips test files because reasons)
3. **Type Check**: Uses Go's type checker to validate method signatures
4. **Match Methods**: Finds structs that implement all interface methods
5. **Output Results**: Spits out JSON with implementation details

## Requirements ✅

- **Go 1.26+**: Because living in the past is for historians
- **go.mod**: Must run from a proper Go module root (not some anarchist directory)
- **Valid Go Code**: Broken syntax makes this tool cry

## Features 🎯

- **Method Set Analysis**: Checks both value and pointer receiver methods
- **Recursive Search**: Crawls directories like a determined spider
- **Type Safety**: Uses Go's actual type checker instead of regex nightmares
- **Package Filtering**: Skips vendor directories and hidden folders automatically
- **Error Handling**: Fails gracefully instead of exploding in your face
- **Debug Mode**: For when things go sideways and you need to know why

## Command Line Options 🛠️

| Flag         | Type   | Default  | Description                             |
| ------------ | ------ | -------- | --------------------------------------- |
| `-interface` | string | required | Interface spec: `file.go:InterfaceName` |
| `-dir`       | string | `.`      | Directory to search for implementations |
| `-debug`     | bool   | `false`  | Enable debug logging                    |
| `-help`      | bool   | `false`  | Show help and exit                      |

## Error Messages 💥

The tool will tell you exactly what's wrong instead of leaving you guessing:

- **Interface file not found**: Check your file path
- **Interface not found in file**: Make sure the interface name exists
- **No go.mod found**: Run from a proper Go module root
- **Directory not found**: Search directory doesn't exist
- **Parse errors**: Fix your Go syntax first

## Testing Coverage 🧪

Current coverage: **90.3%**

Because untested code is like unprotected... well, you get it.

## License 📜

MIT - because pkg.go.dev are corporate fuckers who won't index WTFPL packages.
