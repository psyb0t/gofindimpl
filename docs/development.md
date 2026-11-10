---
title: "Development"
description: "Build, lint and test gofindimpl with its Make targets, the 90 percent coverage gate, the test fixtures module, and how CI and releases run."
nav_order: 5
---

# Development

Hacking on gofindimpl itself. You need Go 1.26 or newer and Make. golangci-lint comes from the module's own `tool` dependency in [`go.mod`](https://github.com/psyb0t/gofindimpl/blob/main/go.mod), so there is nothing to install globally. The `build`, `build-race` and `docs` targets also need Docker.

## Make targets

Run `make help` for the list. The [`Makefile`](https://github.com/psyb0t/gofindimpl/blob/main/Makefile) has:

| Target | What it does |
| --- | --- |
| `make all` | Runs `dep`, `lint-fix`, `test-coverage` and `build`, in that order. |
| `make dep` | `go mod tidy`. |
| `make lint` | Fails if `go fix -diff ./...` has suggestions, then runs `go tool golangci-lint run ./...` with the rules in [`.golangci.yml`](https://github.com/psyb0t/gofindimpl/blob/main/.golangci.yml). |
| `make lint-fix` | Applies `go fix ./...` and runs golangci-lint with `--fix`. |
| `make test` | `go test -race ./...`. |
| `make test-coverage` | Runs the tests with the race detector and a coverage profile, writes the total to `coverage-percent.txt`, and fails below 90 percent. |
| `make build` | Builds a static, stripped binary into `build/gofindimpl` inside a `golang:1.24.6-alpine` container. |
| `make build-race` | The same, with CGO and the race detector. |
| `make docs` | Builds this site into `.testing/docs-site` with the image GitHub Pages uses. |

[`pre-commit.sh`](https://github.com/psyb0t/gofindimpl/blob/main/pre-commit.sh) runs `make lint && make test-coverage`, the same pair CI runs.

## Layout

Everything is one `main` package at the repository root.

| Path | Holds |
| --- | --- |
| [`main.go`](https://github.com/psyb0t/gofindimpl/blob/main/main.go) | Flags, usage text, logging setup, the interface spec parser and the JSON output. |
| [`finder.go`](https://github.com/psyb0t/gofindimpl/blob/main/finder.go) | `Finder`: the `go.mod` checks, interface parsing, the directory walk, type checking and method matching. |
| [`analysis.go`](https://github.com/psyb0t/gofindimpl/blob/main/analysis.go) | The struct check and the `Implementation` result for each match. |
| [`validation.go`](https://github.com/psyb0t/gofindimpl/blob/main/validation.go) | Checks that the interface file and search directory exist. |
| [`errors.go`](https://github.com/psyb0t/gofindimpl/blob/main/errors.go) | Every sentinel error. |
| [`.fixtures/`](https://github.com/psyb0t/gofindimpl/tree/main/.fixtures) | A small module (`testapp`) with an `App` interface in `internal/app`, three structs that implement it and one that does not, used by the tests. |
| [`docs/`](https://github.com/psyb0t/gofindimpl/tree/main/docs) | This site, built by GitHub Pages from the `main` branch. |

## Tests

The tests sit next to the code as `*_test.go` in `package main` and use testify. Several of them change the working directory with `os.Chdir`, into a temporary directory or into `.fixtures/`, because gofindimpl reads `./go.mod`. Those tests cannot run in parallel and say so in a comment; the rest call `t.Parallel()`. The golangci-lint rules skip test files.

## CI

[`.github/workflows/pipeline.yml`](https://github.com/psyb0t/gofindimpl/blob/main/.github/workflows/pipeline.yml) runs on every push, to any branch or tag.

- On a branch push, the shared Go workflow sets up Go 1.26 and runs `make dep`, `make lint`, `make test-coverage` and govulncheck. None of them need credentials.
- When that run is on `main` and passes, the badges job publishes the coverage, version and license badges to the `badges` branch.
- On a `v*` tag, the pipeline does not rerun the checks. It waits for the `main` run of the same commit to pass, cuts the GitHub release with the annotated tag message as its notes, and refreshes the badges with the coverage from that `main` run. The `main` run keeps its coverage artifact for one day, so a tag pushed later than that still gets its release, but the badges job fails.

The other workflows mirror the repository to GitLab and Codeberg, pull issues opened on those mirrors back into GitHub, and close pull requests from anyone who is not a collaborator.

## Releases

The Git tag is the version; there is no version file. A release is a commit on `main` that has passed CI, then an annotated `vX.Y.Z` tag on it whose message is that version's [CHANGELOG](https://github.com/psyb0t/gofindimpl/blob/main/CHANGELOG.md) section. The release has no binaries attached; people install with `go install`.
