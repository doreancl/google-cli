# Spec: `dorean_g auth credentials <credentials.json>`

## Goal

Register OAuth credentials in an explicit, parseable way while keeping compatibility with `auth --credentials`.

## Layer 1 (Functional)

- Command: `dorean_g auth credentials <credentials.json>`.
- Requires exactly 1 argument.
- Validates Desktop OAuth JSON with `google.ConfigFromJSON(..., CalendarScope)`.
- Saves a managed copy at `ConfigDir()/client_secret.json`.
- Expands `~` and accepts relative or absolute paths.

## Layer 2 (Non-Functional)

- Required permissions:
  - directory `0700`
  - file `0600`
- Parseable stdout response:
  - `credentials_path\t<absolute_path>`

## Layer 6 (Testing)

- Unit tests for:
  - success with a valid file
  - missing or extra argument (`exit code 2`)
  - missing/unreadable file
  - invalid JSON
- Regression:
  - `dorean_g auth --credentials <path>` still works
  - `dorean_g auth` without subcommand keeps the legacy flow

## Compatibility

- No breaking changes to `GCAL_CREDENTIALS`.
- No token format changes.
