# Errors and logging

Prefix: ERR.

- **ERR2:** This repository wraps with the standard library, not `ctxerrors`: add context with `fmt.Errorf("...: %w", err)` or `fmt.Errorf("%w: %s", ErrX, detail)`, as `validation.go` and `finder.go` do, so `errors.Is` keeps working in tests.
- **ERR3:** Error text is documented word for word in `docs/troubleshooting.md`. Changing or adding a message updates that page in the same change.
- **ERR4:** `main` logs a failure once with `slog.Error` and exits with status 1. Functions below it return errors; they do not log and exit.
- **ERR5:** Log through the package-level `log/slog` functions, not a logger from the context, with fields such as `interface_file` and `search_dir`. Per-directory details are `slog.Debug`, visible only with `-debug`.
