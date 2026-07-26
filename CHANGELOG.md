# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v1.0.3 — 2026-07-26

Coverage reporting to Codecov + README badges.

- **Codecov coverage upload.** `pipeline.yml` enables the reusable workflow's
  Codecov step; `make test-coverage` keeps `coverage.txt` (previously deleted
  on exit) so CI can upload it.
- **README badges.** pkg.go.dev reference + GitHub Actions CI status badges.
- Added a GitHub Sponsors funding config. No library code changed.

## v1.0.2 — 2026-04-15

- Go 1.26 upgrade; CI restricted to collaborators only.

## v1.0.1 — 2025-09-10

- Maintenance release.

## v1.0.0 — 2025-09-10

- Initial release — find Go interface implementations across a codebase.
