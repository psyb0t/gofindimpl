# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v1.0.7 — 2026-07-27

Coverage badge now uses the dumb-reader badge chain.

- **`make test-coverage` writes the coverage percentage to `coverage-percent.txt`.**
  The `pipeline.yml` test job uploads it as an artifact, and the `badges` job reads
  that value and bakes it into `coverage.svg` — the badge workflow no longer runs
  tests or computes coverage itself. No library code changed.

## v1.0.6 — 2026-07-27

Self-hosted README badges; drop the third-party badge service.

- **Coverage / version / license badges are now self-rendered SVGs** committed to
  a `badges` branch by a new `badges` job in `pipeline.yml`, and served from
  `raw.githubusercontent.com/psyb0t/gofindimpl/badges/*.svg`. No shields.io or
  other external render service in the path.
- **CI badge switched to GitHub's native `badge.svg`** (served by GitHub) instead
  of the shields.io status badge. The `pkg.go.dev` reference badge stays. No
  library code changed.

## v1.0.5 — 2026-07-27

Swap logging from logrus to stdlib `log/slog`; refactor the test suite to
testify.

- **Logging: `github.com/sirupsen/logrus` → stdlib `log/slog`.** Diagnostics now
  emit via `slog` with structured fields, installed through
  `github.com/psyb0t/slog-configurator`. All log output goes to **stderr** so it
  never mixes with the JSON result on stdout; the `-debug` flag selects debug vs
  error level. logrus is no longer a direct dependency.
- **Tests refactored to `stretchr/testify`.** Every `*_test.go` now uses
  `assert`/`require` instead of hand-rolled `if` + `t.Errorf`, with
  `testCases`/`tc` table naming and `t.Parallel()` on the race-safe tests.

## v1.0.4 — 2026-07-27

- Bump `github.com/sirupsen/logrus` 1.9.3 → 1.9.4.

## v1.0.3 — 2026-07-26

README badges + GitHub Sponsors funding config.

- pkg.go.dev reference + GitHub Actions CI status badges.
- Added a GitHub Sponsors funding config. No library code changed.

## v1.0.2 — 2026-04-15

- Go 1.26 upgrade; CI restricted to collaborators only.

## v1.0.1 — 2025-09-10

- Maintenance release.

## v1.0.0 — 2025-09-10

- Initial release — find Go interface implementations across a codebase.
