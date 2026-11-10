# Documentation

Prefix: DOC.

- **DOC1:** `README.md` is the front door: what it is, install, one example, a short limits summary, and the Documentation table. Reference material lives in `docs/`, one home per fact.
- **DOC2:** README links to the docs use full `https://psyb0t.github.io/gofindimpl/...` URLs, and README links to repository files use full `https://github.com/psyb0t/gofindimpl/blob/main/...` URLs, so they work on mirrors and pkg.go.dev.
- **DOC3:** Every page in `docs/` has front matter with a unique `title` and a `description` of 50 to 170 characters, and exactly one `#` heading. Page-to-page links use the relative `.md` path. Links to files outside `docs/` use the full GitHub URL.
- **DOC4:** A change to flags, output, matching rules or error text updates `docs/command-line.md`, `docs/how-it-works.md` or `docs/troubleshooting.md` in the same change. Describe what the code does, including the limits that make it look worse.
- **DOC5:** `make docs` builds the site into `.testing/docs-site` with the image GitHub Pages uses. Build it after editing `docs/` and fix any broken link it shows.
- **DOC6:** No em dashes, "directory" never "folder", plain words. Keep the README's blunt voice in human-facing prose.
