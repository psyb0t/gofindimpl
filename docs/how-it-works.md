---
title: "How it works"
description: "How gofindimpl parses the interface, walks the directory, type-checks each package and matches structs by method name, plus what it misses."
nav_order: 3
---

# How it works

gofindimpl does not shell out to `go list` or load your dependencies. It parses source files with `go/parser`, type-checks each directory with `go/types` and a fake importer, and compares method names. That makes it fast and happy with half-broken code, and it also decides what it can and cannot see. Read the [limits](#limits) before you trust an empty result.

The code is small: [`main.go`](https://github.com/psyb0t/gofindimpl/blob/main/main.go) handles flags and output, [`finder.go`](https://github.com/psyb0t/gofindimpl/blob/main/finder.go) does the parsing, walking and matching, and [`analysis.go`](https://github.com/psyb0t/gofindimpl/blob/main/analysis.go) turns a matching type into a result.

## The steps

1. **Check the arguments.** The interface file and the search directory must exist.
2. **Check the module root.** `./go.mod` must exist in the current directory, and its first line that starts with `module ` gives the module path used in `packagePath`.
3. **Parse the interface.** gofindimpl parses the interface file and collects the names of the methods declared in the interface.
4. **Walk the directory.** Every directory under `-dir` is a candidate package. See [what gets scanned](#what-gets-scanned).
5. **Type-check each package.** All non-test `.go` files in the directory are parsed and type-checked together.
6. **Match.** Every struct type declared at package level whose method set has all the interface's method names is a result.
7. **Print.** The results go to stdout as a JSON array. See [Command line](command-line.md#output).

## What gets scanned

The walk visits every directory under `-dir`, and each directory is analyzed as one package.

- `vendor` and `node_modules` are skipped, with everything under them.
- A directory whose name starts with a dot is not analyzed, but the walk still descends into it. `.cache/foo` is analyzed even though `.cache` is not.
- The search directory itself follows the same rule. `.` and `./` start with a dot, so with the default `-dir .` the package in the module root is never analyzed.
- Only files ending in `.go` count, and files ending in `_test.go` are left out, so test helpers and mocks declared in test files are never found.
- A file that fails to parse is dropped silently, and the rest of the directory is still analyzed.
- The package name comes from the first file that parsed.

`packagePath` is the module path joined with the directory's path relative to the current directory. That only works with a relative `-dir`. With an absolute one the relative path cannot be computed, and every result gets the bare module path.

## What counts as an implementation

A type is reported when all of this is true:

- it is a named type declared at package scope;
- its underlying type is a struct;
- for every method name in the interface, the value method set or the pointer method set of the type has a method with that name.

So a struct with only pointer-receiver methods still counts, and so do methods promoted from a struct embedded from the same package.

## Limits

These come straight from the code, and they matter:

- **Names, not signatures.** Only method names are compared. A struct with `Start()` matches an interface that wants `Start(ctx context.Context) error`. Treat the output as candidates, and let the compiler have the last word.
- **Embedded interfaces are ignored.** Only methods written directly in the interface body are collected. An interface that only embeds others, like `io.ReadWriter`, has no methods to match, and so does an empty interface. Both give `[]`.
- **No imports are resolved.** Every import gets an empty stand-in package. Methods promoted from a struct embedded from another package are invisible, so a type that gets its methods that way is missed.
- **Structs only.** Named non-struct types with methods, such as `type HandlerFunc func(...)`, are never reported.
- **Type errors are ignored.** Unresolved imports and other type errors do not stop the search, which is why broken or incomplete code still works.

The interface file is parsed on its own, so it can sit anywhere, even outside `-dir`.
