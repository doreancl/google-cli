# Spec: Paridad mínima con demo de gogcli

## Objective

Soportar este flujo end-to-end:

```bash
gog auth credentials ~/Downloads/client_secret.json
gog auth add you@gmail.com
export GOG_ACCOUNT=you@gmail.com
gog calendar calendars --max 5 --json | jq '.calendars[].summary'
```

## Command Surface (MVP)

El CLI DEBE soportar:

- `gog auth credentials <credentials.json>`
- `gog auth add <email>`
- `gog calendar calendars [--max N] [--json]`

## 1) `gog auth credentials <credentials.json>`

## Contract

- Recibe exactamente 1 argumento posicional (`credentials.json`).
- Expande `~` y acepta ruta relativa/absoluta.
- Lee y valida JSON OAuth client con `google.ConfigFromJSON(..., CalendarScope)`.
- Guarda credenciales en ruta administrada por app: `ConfigDir()/client_secret.json`.
- Crea directorio con `0700` si no existe.
- Escribe archivo con `0600`.

## Output

- stdout parseable (una línea):
  - `credentials_path\t<absolute_path>`

## Errors

- uso inválido -> exit code `2`.
- archivo no legible -> error con path.
- JSON inválido -> `credenciales invalidas: ...`.

## 2) `gog auth add <email>`

## Contract

- Recibe exactamente 1 argumento posicional (`email`).
- Email se normaliza: `strings.ToLower(strings.TrimSpace(email))`.
- Si email vacío/sin `@`/con espacios -> exit code `2`.
- Carga credenciales desde `ConfigDir()/client_secret.json` (o resolución actual compatible).
- Ejecuta OAuth browser flow (URL + pegar código) como hoy.
- Intercambia código por token OAuth.
- Si no hay `refresh_token`, devuelve error explícito (usuario debe re-consentir).
- Guarda token por cuenta en `ConfigDir()/tokens/<normalized-email>.json`.
- `tokens/` con permisos `0700`, archivo token `0600`.

## Output

- stdout parseable:
  - `email\t<normalized_email>`
  - `token_path\t<absolute_path>`

## Errors

- uso inválido -> exit code `2`.
- credenciales faltantes -> mensaje que sugiera `gog auth credentials <credentials.json>`.
- token exchange fail -> `no se pudo intercambiar el token: ...`.

## 3) Selección de cuenta (`GOG_ACCOUNT`)

## Contract

- `GOG_ACCOUNT` define la cuenta activa para comandos API.
- Resolución para leer token:
  1. Si `GOG_ACCOUNT` está seteado: `ReadTokenForAccount(GOG_ACCOUNT)`.
  2. Si no está seteado: fallback legado a `ReadToken()` (`token.json`).
- Si `GOG_ACCOUNT` apunta a token inexistente:
  - error: `token no encontrado para <email>, corre \`gog auth add <email>\``.

## 4) `gog calendar calendars [--max N] [--json]`

## Contract

- Nuevo comando en namespace `calendar`.
- Lista calendarios del usuario autenticado (cuenta activa por `GOG_ACCOUNT`).
- `--max` default `50` (si no se define).
- `--json` devuelve objeto estable:

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

- Sin `--json`, salida tabular parseable en stdout.

## Compatibility requirements

- `gog auth` legacy existente NO se rompe.
- `gog auth --credentials ...` legacy NO se rompe durante transición.
- `events list/create` existentes NO se rompen.

## Exit codes

- `0`: éxito.
- `2`: uso inválido/flags inválidas.
- `1`: error operativo/API/auth.

## Test Requirements

## Unit

- `auth credentials`:
  - guarda archivo válido en destino.
  - falla con arg faltante/extra (code 2).
  - falla con JSON inválido.
- `auth add`:
  - guarda token por cuenta.
  - valida email.
  - falla sin refresh token.
- `authstore`:
  - `TokenPathForAccount`, `SaveTokenForAccount`, `ReadTokenForAccount`.

## Integration-ish

- `googleapi/newHTTPClient` usa `GOG_ACCOUNT` cuando existe.
- fallback a token legacy si `GOG_ACCOUNT` vacío.
- `calendar calendars --json` serializa shape esperado (`calendars[].summary` usable por jq).

## Non-goals (MVP)

- `auth credentials list`.
- `auth list/remove/status`.
- keyring nativo del OS.
- multi-client (`--client`).

## Future parity (post-MVP)

- `gog auth credentials list`.
- `gog auth list`.
- `gog auth remove <email>`.
- `gog --client <name> ...`.
