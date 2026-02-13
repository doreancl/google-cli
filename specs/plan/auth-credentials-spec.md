# Spec: `dorean_g auth credentials <credentials.json>`

## Goal

Registrar credenciales OAuth de forma explícita y parseable, manteniendo compatibilidad con `auth --credentials`.

## Layer 1 (Functional)

- Comando: `dorean_g auth credentials <credentials.json>`.
- Requiere exactamente 1 argumento.
- Valida JSON OAuth Desktop con `google.ConfigFromJSON(..., CalendarScope)`.
- Guarda copia administrada en `ConfigDir()/client_secret.json`.
- Expande `~` y acepta ruta relativa o absoluta.

## Layer 2 (Non-Functional)

- Permisos obligatorios:
  - directorio `0700`
  - archivo `0600`
- Respuesta en stdout parseable:
  - `credentials_path\t<absolute_path>`

## Layer 6 (Testing)

- Unit tests para:
  - éxito con archivo válido
  - argumento faltante o extra (`exit code 2`)
  - archivo inexistente/no legible
  - JSON inválido
- Regression:
  - `dorean_g auth --credentials <path>` sigue funcionando
  - `dorean_g auth` sin subcomando mantiene flujo legacy

## Compatibility

- Sin ruptura de `GCAL_CREDENTIALS`.
- Sin cambios al formato de token.
