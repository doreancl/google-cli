# Spec: Minimal auth + calendar parity

## Goal

Support this end-to-end flow:

```bash
dorean_g auth credentials ~/Downloads/client_secret.json
dorean_g auth add you@gmail.com
export GOG_ACCOUNT=you@gmail.com
dorean_g calendar calendars --max 5 --json | jq '.calendars[].summary'
```

## Layer 1 (Functional)

The CLI must support:

- `dorean_g auth credentials <credentials.json>`
- `dorean_g auth add <email>`
- `dorean_g calendar calendars [--max N] [--json]`

`calendar calendars` contract:

- Lists calendars for the active account.
- `--max` defaults to `50`; `--max <= 0` is a usage error.
- `--json` returns:

```json
{
  "calendars": [
    {
      "id": "primary",
      "summary": "Personal",
      "primary": true
    }
  ]
}
```

## Layer 2 (Non-Functional)

- Parseable output in text mode (stable columns).
- Exit codes:
  - `0` success
  - `2` invalid usage
  - `1` operational/auth/API error

## Layer 8 (Integrations)

- Active account is selected by `GOG_ACCOUNT`.
- Legacy fallback to `token.json` if `GOG_ACCOUNT` is not set.

## Layer 6 (Testing)

- Unit:
  - `auth credentials` valid/invalid
  - `auth add` valid/invalid/without refresh token
  - per-account storage
- Integration-ish:
  - API client uses `GOG_ACCOUNT` when present
  - legacy fallback when absent
  - `calendar calendars --json` keeps a stable shape for `jq`

## Compatibility

- Do not break:
  - `dorean_g auth` legacy
  - `dorean_g auth --credentials ...`
  - `dorean_g events list/create`
