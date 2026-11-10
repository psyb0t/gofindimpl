---
title: "Home"
description: "gofindimpl is a Go command line tool that finds the structs in a module that implement a given interface and prints them as JSON."
nav_order: 1
permalink: /
seo:
  type: WebSite
---

# gofindimpl

Hunt down Go interface implementations like a bloodhound with trust issues.

You point gofindimpl at an interface in a Go file and a directory to search. It parses the packages under that directory, keeps the structs that have a method for each method name in the interface, and prints them as a JSON array on stdout. No grep, no guessing, no 3am existential crisis.

```bash
gofindimpl \
  -interface ./internal/app/server.go:Server \
  -dir ./internal/pkg/
```

```json
[
  {
    "package": "impl",
    "struct": "WebServer",
    "packagePath": "github.com/yourproject/internal/pkg/impl"
  }
]
```

## Install

```bash
go install github.com/psyb0t/gofindimpl@latest
```

Or run it without installing anything:

```bash
go run github.com/psyb0t/gofindimpl@latest \
  -interface ./path/to/file.go:InterfaceName \
  -dir ./search/directory
```

Or clone and build it like it's 2005:

```bash
git clone https://github.com/psyb0t/gofindimpl.git
cd gofindimpl
go build -o gofindimpl
```

The module declares `go 1.26` in its [`go.mod`](https://github.com/psyb0t/gofindimpl/blob/main/go.mod), so you need Go 1.26 or newer to build it. Run it from the root of the module you are searching, the directory that holds its `go.mod`.

## Where to go next

| You want to | Read |
| --- | --- |
| See every flag, the interface spec format and the JSON output | [Command line](command-line.md) |
| Know exactly what counts as an implementation, and what gets missed | [How it works](how-it-works.md) |
| Decode an error message | [Troubleshooting](troubleshooting.md) |
| Build, lint and test gofindimpl itself | [Development](development.md) |

The source is on [GitHub](https://github.com/psyb0t/gofindimpl).
