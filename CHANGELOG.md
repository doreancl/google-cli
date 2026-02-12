# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Initial Google Calendar CLI commands:
  - `dorean_g auth`
  - `dorean_g events list`
  - `dorean_g events create`
- Google API layer with a reusable generic client constructor in `internal/googleapi/client.go`.
- Calendar domain client with request/response models, input validation, and field masks in `internal/googleapi/calendar.go`.
- OAuth credential/token storage layer in `internal/authstore/store.go`.
- Build and developer workflow:
  - `Makefile` (build, fmt, lint, test, ci, aliases `dorean_g` and `dg`)
  - `.golangci.yml`
  - `.lefthook.yml`
  - `.gitignore` (based on `gogcli`)
- Project docs:
  - `docs/plan/features.md`
  - `docs/google-calendar.md`
  - `AGENTS.md`
  - `docs/base-repo.md`
  - `docs/RELEASING.md`
- Coverage and quality gate workflow:
  - `scripts/check-coverage.sh` (coverage threshold enforcement, default `90%`)
  - `scripts/lint.sh`
  - `make coverage` and `make coverage-check`
  - `pre-push` hook runs `coverage-check`
- Comprehensive unit test suite added across:
  - `cmd/gog`
  - `internal/authstore`
  - `internal/cmd`
  - `internal/googleapi`

### Changed
- Project layout aligned with `gogcli` pattern:
  - entrypoint moved to `cmd/gog/main.go`
  - main delegates exit code handling via `internal/cmd/exit.go`
- CLI now accepts optional leading `--` before subcommands.
- `make ci` now includes coverage validation through `make coverage-check`.
- Coverage threshold configuration is centralized in `Makefile` via `COVERAGE_THRESHOLD` (default `90`).

### Fixed
- Help command exits with code `0` (`--help` and empty args path).
- Error messages are printed to stderr from main entrypoint.
