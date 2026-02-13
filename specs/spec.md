# dorean_g Product Spec

## Goal

Build one clean Go CLI for Google Calendar with explicit OAuth account management and script-friendly output.

## Non-goals

- Backward compatibility with legacy CLIs.
- Migration tooling for old config or token layouts.
- Non-Calendar product surfaces (Gmail, Drive, Classroom, etc.).

## Backlog

- See `/Users/sebecode/code/1dev/google-cli/specs/spec-backlog.md`.

## Target State

### Runtime and binary

- Language: Go (module in `go.mod`).
- Binary name: `dorean_g`.

### Command framework

- Parser: `github.com/alecthomas/kong`.
- Root command: `dorean_g`.
- Global flags:
  - `--color=auto|always|never` (default: `auto`)
  - `--json` (JSON output to `stdout`)
  - `--plain` (TSV output to `stdout`; stable and parseable; disables colors)
  - `--force` (skip confirmations for destructive commands)
  - `--no-input` (never prompt; fail instead)
  - `--version` (print version)
- Environment defaults:
  - `GOG_COLOR=auto|always|never` (overridden by `--color`)
  - `GOG_JSON=1` (overridden by explicit format flags)
  - `GOG_PLAIN=1` (overridden by explicit format flags)
- Output behavior:
  - Errors are explicit and written to `stderr`.
  - `NO_COLOR` is respected.

### Authentication and account model

- OAuth credentials are managed in the per-user config directory by the auth store.
- Supported credential format: Google OAuth client JSON accepted by the auth loader.
- Commands:
  - `dorean_g auth credentials <credentials.json>`
  - `dorean_g auth credentials list`
  - `dorean_g auth add <email> [--credentials path]`
- `auth add` runs the desktop OAuth flow, exchanges the code, and stores a reusable token.
- Email input is validated and normalized.
- Calendar commands resolve the active account from `GOG_ACCOUNT`.
- If token or required scope is missing or invalid, commands fail with actionable errors.

### Configuration

- Credential and token paths are resolved by `internal/authstore`.
- `GCAL_CREDENTIALS` overrides the credentials path for auth flows.
- `GOG_ACCOUNT` selects the active account for Google API commands.

### Command surface

- `dorean_g auth credentials <credentials.json>`
- `dorean_g auth credentials list`
- `dorean_g auth add <email> [--credentials path]`
- `dorean_g events list [--calendar primary] [--days 7]`
- `dorean_g events create --summary S --start RFC3339 --end RFC3339 [--calendar primary] [--description D] [--location L]`
- `dorean_g calendar calendars [--json|--plain]`

### Output contract

- Successful command data goes to `stdout`.
- Errors, help, and hints go to `stderr`.
- Exit codes:
  - `0` for success
  - `1` for operational errors (auth, API, I/O)
  - `2` for usage or argument errors
- Color behavior:
  - Enabled when `--color=always`, or when `--color=auto` with a capable terminal and no `NO_COLOR`.
  - Disabled when `--color=never` or when `NO_COLOR` is set.

### Calendar list contract

For `dorean_g calendar calendars [--json|--plain]`:

- Returns calendars for the active account.
- Stable order by `id` ascending.
- Fields per calendar:
  - `id` (string)
  - `summary` (string, fallback: `(untitled)`)
  - `primary` (boolean)
- `--json`: valid JSON object with a `calendars` array.
- `--plain`: stable TSV with header `ID\tSUMMARY\tPRIMARY`.

### Compatibility

- Existing behavior of `events list` and `events create` remains unchanged while adding `calendar calendars`.
