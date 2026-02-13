# Spec: `dorean_g auth add <email>`

## Goal

Permitir token por cuenta y selección explícita de cuenta activa vía `GOG_ACCOUNT`.

## Layer 1 (Functional)

- Comando: `dorean_g auth add <email>`.
- Requiere exactamente 1 argumento.
- Normaliza email con `strings.ToLower(strings.TrimSpace(email))`.
- Ejecuta OAuth browser flow con credenciales resueltas por `ResolveCredentialsPath("")`.
- Requiere `refresh_token`.
- Guarda token en `ConfigDir()/tokens/<normalized-email>.json`.
- Stdout parseable:
  - `email\t<normalized_email>`
  - `token_path\t<absolute_path>`

## Layer 2 (Non-Functional)

- Permisos obligatorios:
  - directorio `tokens/` con `0700`
  - archivo token con `0600`
- Exit codes:
  - `2`: uso inválido
  - `1`: error operativo/auth/API

## Layer 8 (Integrations)

- Resolución de cuenta para cliente Google API:
1. Si `GOG_ACCOUNT` existe: usar `ReadTokenForAccount(GOG_ACCOUNT)`.
2. Si no existe: fallback a `ReadToken()` (`token.json` legacy).
- Si falta token para `GOG_ACCOUNT`: error con hint a `dorean_g auth add <email>`.

## Layer 6 (Testing)

- Unit tests para:
  - éxito con email válido
  - email inválido o argumento faltante (`exit code 2`)
  - error en exchange OAuth
  - token sin `refresh_token`
- Tests de storage:
  - `TokenPathForAccount`
  - `SaveTokenForAccount` / `ReadTokenForAccount`
- Regression:
  - `dorean_g auth` legacy sigue funcionando
  - `dorean_g auth --credentials` sigue funcionando
