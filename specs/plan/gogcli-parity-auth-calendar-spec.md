# Spec: Paridad mínima auth + calendar

## Goal

Soportar este flujo de punta a punta:

```bash
dorean_g auth credentials ~/Downloads/client_secret.json
dorean_g auth add you@gmail.com
export GOG_ACCOUNT=you@gmail.com
dorean_g calendar calendars --max 5 --json | jq '.calendars[].summary'
```

## Layer 1 (Functional)

El CLI debe soportar:

- `dorean_g auth credentials <credentials.json>`
- `dorean_g auth add <email>`
- `dorean_g calendar calendars [--max N] [--json]`

Contrato de `calendar calendars`:

- Lista calendarios de la cuenta activa.
- `--max` default `50`; `--max <= 0` es error de uso.
- `--json` devuelve:

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

- Salida parseable en modo texto (columnas estables).
- Exit codes:
  - `0` éxito
  - `2` uso inválido
  - `1` error operativo/auth/API

## Layer 8 (Integrations)

- Cuenta activa por `GOG_ACCOUNT`.
- Fallback legado a `token.json` si `GOG_ACCOUNT` no está seteado.

## Layer 6 (Testing)

- Unit:
  - `auth credentials` válido/inválido
  - `auth add` válido/inválido/sin refresh token
  - storage por cuenta
- Integration-ish:
  - cliente API usa `GOG_ACCOUNT` cuando existe
  - fallback legacy cuando no existe
  - `calendar calendars --json` mantiene shape estable para `jq`

## Compatibility

- No romper:
  - `dorean_g auth` legacy
  - `dorean_g auth --credentials ...`
  - `dorean_g events list/create`
