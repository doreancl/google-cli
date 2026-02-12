# Google Calendar Spec

## Goal

Exponer una salida JSON estable para listar calendarios, compatible con:

```bash
dorean_g calendar calendars --max 5 --json | jq '.calendars[].summary'
```

## Commands (MVP)

- `dorean_g calendar calendars [--max N] [--json]`

## Calendar Calendars

Command:

- `dorean_g calendar calendars [--max N] [--json]`

Behavior:

- Lista calendarios de la cuenta activa.
- Usa autenticacion resuelta por `GOG_ACCOUNT` (ver `docs/google-auth-spec.md`).
- `--max` limita resultados; default `50`.
- Si `--max <= 0`, retorna error de uso (exit code `2`).

JSON output (`--json`):

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

Text output (sin `--json`):

- Tabla parseable con columnas:
  - `ID`
  - `SUMMARY`
  - `PRIMARY`

Errors:

- Uso invalido/flags invalidas: exit code `2`.
- Error de auth/API: exit code `1`.

## Compatibility

- Los comandos existentes `events list` y `events create` no se rompen.
- El modelo de token legacy (`token.json`) sigue funcionando como fallback.

## Out Of Scope (MVP)

- `dorean_g calendar events/get/create/update/delete`
- paginacion explicita (`--page`)
- filtros avanzados (`--query`, `--from`, `--to`)
