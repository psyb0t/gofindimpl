---
title: "Troubleshooting"
description: "Every gofindimpl error message with its cause and fix, from interface spec format errors to a missing go.mod, plus why a search can come back empty."
nav_order: 4
---

# Troubleshooting

Errors go to stderr as one `slog` line and the exit status is 1. The text after `err=` is what to search for on this page:

```text
time=2026-11-10T06:51:26.201Z level=ERROR msg="finder failed" err="interface file does not exist: x.go"
```

The messages come from [`errors.go`](https://github.com/psyb0t/gofindimpl/blob/main/errors.go) and the wrapping in [`finder.go`](https://github.com/psyb0t/gofindimpl/blob/main/finder.go) and [`validation.go`](https://github.com/psyb0t/gofindimpl/blob/main/validation.go). Checks run in this order, so the first problem found is the one reported: spec format, interface file, search directory, `go.mod`, module line, interface parse.

## Spec errors

These are logged as `msg="failed to parse interface spec"`.

| Message | Cause | Fix |
| --- | --- | --- |
| `interface specification cannot be empty. Use -interface flag with format 'file.go:InterfaceName'` | `-interface` was not given. | Pass `-interface path/to/file.go:Name`. |
| `interface specification must be in format 'file.go:InterfaceName'` | The spec has no colon, or more than one. | Use exactly one colon. A path with a colon in it, such as `C:\...`, cannot work; use a relative path. |
| `interface file path cannot be empty` | Nothing before the colon. | Put the file path before the colon. |
| `interface name cannot be empty` | Nothing after the colon. | Put the interface name after the colon. |

## Run errors

These are logged as `msg="finder failed"`.

| Message | Cause | Fix |
| --- | --- | --- |
| `interface file does not exist: <path>` | The file part of the spec does not exist, relative to the current directory. | Fix the path, or `cd` to where it makes sense. |
| `search directory does not exist: <dir>` | `-dir` does not exist. | Fix the path. |
| `go.mod not found in current directory - must launch from go module root` | There is no `go.mod` in the current directory. gofindimpl does not look in parent directories. | `cd` to the module root and run it from there. |
| `no module declaration found in go.mod` | `go.mod` has no line starting with `module `. | Fix `go.mod`. |
| `failed to read go.mod: <reason>` | `go.mod` exists but cannot be read, for example because of permissions. | Fix the file permissions. |
| `failed to parse interface file <path>: <parser error>` | The interface file has a syntax error. | Fix the Go syntax. The parser error gives the line and column. |
| `interface not found in file '<Name>' in <path>` | No interface with that name is declared in the file. A non-interface type with the same name does not count. Names are case-sensitive. | Check the spelling and the file. |
| `failed to scan directory: <reason>` | The walk hit a directory it cannot read. | Fix the permissions, or point `-dir` somewhere narrower. |

## The search finds nothing

`[]` with exit status 0 means the search ran and nothing matched. Common reasons, all explained in [How it works](how-it-works.md#limits):

- The implementation lives in the module root and `-dir` is `.`, which never analyzes the root package.
- The implementation is in a `_test.go` file, under `vendor`, `node_modules`, or in a directory whose name starts with a dot.
- The interface only embeds other interfaces, or is empty, so there are no method names to match.
- The struct gets some of its methods from a struct embedded from another package. Imports are not resolved, so those methods are invisible.
- The implementing type is not a struct.

Run again with `-debug` to see each directory walked and analyzed, how many files were parsed in each, and the interface methods collected.

## The search finds too much

Matching is by method name only, so a struct with the right names and the wrong signatures is reported. Check the candidates with the compiler, for example with a compile-time assertion in your own code:

```go
var _ Server = (*WebServer)(nil)
```

## Every packagePath is the same

`-dir` was an absolute path. Use a path relative to the module root instead.
