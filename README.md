# Google Calendar CLI

A simple Go CLI to authenticate with Google Calendar, list events, and create events.

## Build

```bash
make build
```

This creates `bin/dorean_g`. In this README, `dorean_g` refers to that executable binary.

## 1) Setup OAuth credentials

1. Open Google Cloud Console.
2. Create a project (or use an existing one).
3. Enable **Google Calendar API**.
4. Configure the **OAuth consent screen**.
5. Create an **OAuth Client ID** for **Desktop app**.
6. Download the JSON file and save it as:
   - macOS: `~/Library/Application Support/dorean_g/client_secret.json`
   - Linux: `~/.config/dorean_g/client_secret.json`
   - or anywhere and pass `--credentials`.

## 2) Authenticate

```bash
./bin/dorean_g auth --credentials /path/to/client_secret.json
```

Or:

```bash
export GCAL_CREDENTIALS=/path/to/client_secret.json
./bin/dorean_g auth
```

Token path:
- macOS: `~/Library/Application Support/dorean_g/token.json`
- Linux: `~/.config/dorean_g/token.json`

## 3) List events

```bash
./bin/dorean_g events list --calendar primary --days 7
```

## 4) Create event

```bash
./bin/dorean_g events create \
  --calendar primary \
  --summary "Weekly sync" \
  --start "2026-02-12T10:00:00-06:00" \
  --end "2026-02-12T10:30:00-06:00" \
  --description "Weekly agenda"
```

## Shortcut with Make

```bash
make dg -- events list --calendar primary --days 7
```

`make dg` is a shortcut that builds and runs `bin/dorean_g`.

## Development

```bash
make ci
make coverage
```

## Credits

Inspired by [gogcli](https://github.com/steipete/gogcli).

## Notes

- Date/time format: RFC3339.
- This is a base CLI; next steps can include `update`, `delete`, text filters, and JSON output.
