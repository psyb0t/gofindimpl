# CI

Prefix: CI.

- **CI1:** `.github/workflows/pipeline.yml` runs `on: [push]` with top-level `permissions: {}`. Each job grants only what its reusable workflow needs.
- **CI2:** The `code` job calls `go-workflow.yml` with `make dep`, `make lint` and `make test-coverage`, and uploads `coverage-percent.txt` for the badges. CI calls Make targets; a new check goes into the `Makefile` first, so it runs the same locally.
- **CI3:** On a `v*` tag the checks are not rerun: `wait` blocks on the `main` run of the same commit, then `release` cuts the GitHub release from the annotated tag message, and `badges` reads that run's coverage. Push the commit to `main` and let it pass before tagging.
- **CI4:** `mirror-and-archive.yml`, `issue-pull.yml` and `collaborators-only.yml` are separate concerns. Leave them alone when changing the pipeline.
- **CI5:** Validate workflow changes with `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:1.7.7`.
