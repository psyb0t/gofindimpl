# Errors and logging

Prefix: ERR.

- **ERR1:** Every error a user can see is a sentinel in `errors.go`, named `Err...`. Add a new one there instead of an inline `errors.New` or a bare `fmt.Errorf` string.
- **ERR2:** Add context with `fmt.Errorf("...: %w", err)` or `fmt.Errorf("%w: %s", ErrX, detail)`, as `validation.go` and `finder.go` do, so `errors.Is` keeps working in tests.
- **ERR3:** Error text is documented word for word in `docs/troubleshooting.md`. Changing or adding a message updates that page in the same change.
- **ERR4:** `main` logs a failure once with `slog.Error` and exits with status 1. Functions below it return errors; they do not log and exit.
- **ERR5:** Log through the package-level `log/slog` functions with structured key-value fields in snake_case (`interface_file`, `search_dir`). Per-directory details are `slog.Debug`, visible only with `-debug`.
