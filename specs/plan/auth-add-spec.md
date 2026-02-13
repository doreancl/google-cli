# Spec: `dorean_g auth add <email>`

## Goal

Allow per-account tokens and explicit active-account selection via `GOG_ACCOUNT`.

## Layer 1 (Functional)

- Command: `dorean_g auth add <email>`.
- Requires exactly 1 argument.
- Normalizes email with `strings.ToLower(strings.TrimSpace(email))`.
- Runs the OAuth browser flow with credentials resolved by `ResolveCredentialsPath("")`.
- Requires `refresh_token`.
- Saves token at `ConfigDir()/tokens/<normalized-email>.json`.
- Parseable stdout:
  - `email\t<normalized_email>`
  - `token_path\t<absolute_path>`

## Layer 2 (Non-Functional)

- Required permissions:
  - `tokens/` directory with `0700`
  - token file with `0600`
- Exit codes:
  - `2`: invalid usage
  - `1`: operational/auth/API error

## Layer 8 (Integrations)

- Account resolution for Google API client:
1. If `GOG_ACCOUNT` exists: use `ReadTokenForAccount(GOG_ACCOUNT)`.
2. If it does not: fallback to `ReadToken()` (legacy `token.json`).
- If token is missing for `GOG_ACCOUNT`: error with hint to `dorean_g auth add <email>`.

## Layer 6 (Testing)

- Unit tests for:
  - success with a valid email
  - invalid email or missing argument (`exit code 2`)
  - OAuth exchange error
  - token without `refresh_token`
- Storage tests:
  - `TokenPathForAccount`
  - `SaveTokenForAccount` / `ReadTokenForAccount`
- Regression:
  - `dorean_g auth` legacy still works
  - `dorean_g auth --credentials` still works
