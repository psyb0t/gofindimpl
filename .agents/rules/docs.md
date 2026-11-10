# Documentation

Prefix: DOC.

- **DOC2:** README links to the docs use full `https://psyb0t.github.io/gofindimpl/...` URLs, and README links to repository files use full `https://github.com/psyb0t/gofindimpl/blob/main/...` URLs, so they work on mirrors and pkg.go.dev.
- **DOC4:** A change to flags, output, matching rules or error text updates `docs/command-line.md`, `docs/how-it-works.md` or `docs/troubleshooting.md` in the same change.
- **DOC5:** `make docs` builds the site into `.testing/docs-site` with the image GitHub Pages uses. Build it after editing `docs/` and fix any broken link it shows.
- **DOC6:** Keep the README's blunt voice in human-facing prose.
