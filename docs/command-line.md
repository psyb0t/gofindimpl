---
title: "Command line"
description: "Every gofindimpl flag, the file.go:InterfaceName spec format, the JSON output fields, exit codes and where the logs go."
nav_order: 2
---

# Command line

gofindimpl is one command with four flags. Run it from the root of the Go module you want to search, the directory that holds `go.mod`, because it reads the module path from there and refuses to start without it.

```bash
gofindimpl \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/
```

The same thing without installing:

```bash
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/
```

## Flags

The flags are defined in [`main.go`](https://github.com/psyb0t/gofindimpl/blob/main/main.go).

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-interface` | string | none, required | Interface spec in the form `file.go:InterfaceName`. |
| `-dir` | string | `.` | Directory to search for implementations, walked recursively. |
| `-debug` | bool | `false` | Log every step at debug level on stderr. |
| `-help` | bool | `false` | Print the usage text and exit with status 0. |

Go's `flag` package parses them, so `-flag value`, `-flag=value` and `--flag` all work, and `-h` prints the same usage text as `-help`.

## The interface spec

`-interface` takes a file path and an interface name joined by one colon:

```text
./internal/app/server.go:Server
```

- The spec is split on `:` and must have exactly two parts. A path that contains a colon itself, such as a Windows drive letter, is rejected with a format error.
- Spaces around either part are trimmed.
- The path is relative to the current directory, or absolute.
- The file is parsed on its own. The first type declaration named `InterfaceName` that is an interface wins. A type with that name that is not an interface is skipped.

## Search directory

`-dir` is walked recursively. Every directory found on the way is analyzed as one package, except:

- `vendor` and `node_modules`, which are skipped along with everything under them;
- directories whose name starts with a dot, which are not analyzed themselves, although the walk still goes into their subdirectories.

Two catches, both covered in [How it works](how-it-works.md#what-gets-scanned):

- The default `.` (and `./`) is itself a name that starts with a dot, so the package in the module root is never analyzed, only the packages below it.
- Give `-dir` as a path relative to the current directory. An absolute path does reach the root package, but every `packagePath` in the output then comes out as the bare module path.

## Output

stdout gets one JSON array, indented with two spaces and followed by a newline, and nothing else:

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

| Field | Meaning |
| --- | --- |
| `package` | The package name from the `package` clause. |
| `struct` | The struct type name. |
| `packagePath` | The module path from `go.mod`, joined with the directory's path relative to the current directory, with forward slashes. |

No match prints `[]` and still exits with status 0. Results come in walk order, directories in lexical order, and structs sorted by name inside each package.

Because stdout holds only the JSON, you can pipe it straight into `jq`:

```bash
gofindimpl \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/ |
  jq -r '.[].struct'
```

## Logs and exit codes

All log output goes to stderr through Go's `log/slog` text handler, so it never mixes with the JSON. Without `-debug` only errors are logged. With `-debug` you also see the parsed arguments, the interface methods found, and every directory walked and analyzed.

```bash
gofindimpl \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/ \
  -debug
```

| Exit status | When |
| --- | --- |
| `0` | The search finished, with or without matches, or `-help` or `-h` was given. |
| `1` | A bad interface spec, a missing file or directory, a missing `go.mod`, or a parse error. The message is on stderr; see [Troubleshooting](troubleshooting.md). |
| `2` | An unknown flag. Go's `flag` package prints the usage text. |
