# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v1.0.12 — 2026-08-08

Documentation. No code change.

- The README stated coverage as **90.8%**; `go test -cover` reports **90.3%**.
  A hand-maintained number in prose drifts the moment a test lands, which is why
  the badge on the same page is generated instead.

## v1.0.11 — 2026-08-08

Dependency bump only. No behaviour changed.

- `slogging` v1.6.1 → v1.7.0, which rebuilt its handler API. Nothing here needed
  editing: this only calls `SetHandlers`, which kept both its name and its
  meaning — replace the whole chain. That is the right call for a CLI that wants
  its own output and nothing else, and the one call the rework deliberately left
  alone as the start-over escape hatch.

## v1.0.10 — 2026-08-08

Dependency rename only. No behaviour changed.

- `github.com/psyb0t/slog-configurator` became `github.com/psyb0t/slogging`, with
  the configurator now at `slogging/slogconf`. Updated the import and dropped the
  `slogconf` alias it no longer needs — the package is already named `slogconf`.
  The `SetHandlers` call and every logging default are unchanged.

## v1.0.9 — 2026-08-01

Repository infrastructure only. No library code changed.

- Added `.github/workflows/mirror-and-archive.yml`: pushes are mirrored to GitLab
  and Codeberg, and the default branch plus tags are archived to the Wayback
  Machine and Software Heritage. Feature-branch pushes skip the archive because
  it is rate-limited. Gitee mirroring is wired but disabled — it silently creates
  the repo private unless the account has a mobile number bound.
- Added `.github/workflows/issue-pull.yml`: issues opened on the GitLab and
  Codeberg mirrors are pulled back into this repo every six hours. The schedule
  is staggered per repo and jitters on top of that, since GitHub fires an
  account's crons at the same instant; a manual `workflow_dispatch` run skips
  the jitter.
- Added `.dockerignore` excluding `.telemetry/` so a local scratch dir cannot
  enter the build context.

## v1.0.8 — 2026-07-27

Lint tooling.

- `make lint` now runs `go fix -diff` as a read-only check (it previously applied
  fixes in-place during `lint`); run `make lint-fix` to apply them. No library
  code changed.

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
