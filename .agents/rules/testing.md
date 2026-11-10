# Testing

Prefix: TEST.

- **TEST1:** Tests use testify (`assert`, `require`).
- **TEST2:** A test that calls `os.Chdir` changes the whole process's working directory, so it must not call `t.Parallel()` and says so in a comment, as the existing ones do. It restores the old directory in `t.Cleanup`.
- **TEST3:** End-to-end behaviour is tested against the `.fixtures/` module (`fixtures_test.go`, `integration_test.go`) or a module written into `t.TempDir()`. Change the fixtures only together with the tests that count on them.
- **TEST4:** The coverage floor is 90% (`MIN_TEST_COVERAGE` in `Makefile`).
