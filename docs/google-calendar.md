# Google Calendar

## Commands

- `dorean_g auth --credentials /path/client_secret.json`
- `dorean_g events list --calendar primary --days 7`
- `dorean_g events create --calendar primary --summary "Title" --start "2026-02-12T10:00:00-06:00" --end "2026-02-12T10:30:00-06:00"`

## Auth and Credentials

- Token path:
  - macOS: `~/Library/Application Support/dorean_g/token.json`
  - Linux: `~/.config/dorean_g/token.json`
- Credentials path:
  - use `--credentials` or `GCAL_CREDENTIALS`
  - default path:
    - macOS: `~/Library/Application Support/dorean_g/client_secret.json`
    - Linux: `~/.config/dorean_g/client_secret.json`

## Notes

- Time format for create command is RFC3339.
- Event list uses `primary` calendar by default.
