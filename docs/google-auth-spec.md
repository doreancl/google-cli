# Google Auth Spec

## Goal

Proveer el flujo mínimo de autenticacion inspirado en `gogcli`, adaptado a este proyecto:

```bash
dorean_g auth credentials ~/Downloads/client_secret.json
dorean_g auth add you@gmail.com
export GOG_ACCOUNT=you@gmail.com
```

## Commands (MVP)

- `dorean_g auth credentials <credentials.json>`
- `dorean_g auth credentials list`
- `dorean_g auth add <email>`

## Auth Credentials

Command:

- `dorean_g auth credentials <credentials.json>`

Behavior:

- Requiere exactamente 1 argumento posicional.
- Expande `~` y acepta path absoluto o relativo.
- Lee archivo y valida con `google.ConfigFromJSON(..., CalendarScope)`.
- Guarda copia administrada en `ConfigDir()/client_secret.json`.
- Crea directorio destino con `0700`.
- Escribe archivo destino con `0600`.

Stdout (parseable):

- `credentials_path\t<absolute_path>`

## Auth Credentials List

Command:

- `dorean_g auth credentials list`

Behavior:

- Lista credenciales OAuth registradas localmente.
- En MVP de cliente unico, devuelve solo la credencial activa (`client_secret.json`) si existe.
- Si no existe credencial almacenada, retorna lista vacia (no error).

Stdout (parseable):

- `credentials_path\t<absolute_path>` por cada entrada (1 en MVP).

JSON output (cuando se habilite `--json` global):

```json
{
  "credentials": [
    {
      "path": "/abs/path/client_secret.json"
    }
  ]
}
```

Errors:

- Uso invalido: exit code `2`.
- Archivo inexistente/no legible: error con contexto de path.
- JSON invalido: `credenciales invalidas: ...`.

## Auth Add

Command:

- `dorean_g auth add <email>`

Behavior:

- Requiere exactamente 1 argumento posicional.
- Normaliza email con `strings.ToLower(strings.TrimSpace(email))`.
- Email invalido (vacio, sin `@`, con espacios): exit code `2`.
- Carga credenciales OAuth desde `ResolveCredentialsPath("")`.
- Ejecuta browser flow (imprime URL, abre navegador, pide codigo, exchange).
- Requiere `refresh_token` en la respuesta; si falta, error explicito.
- Guarda token por cuenta en `ConfigDir()/tokens/<normalized-email>.json`.
- Crea `tokens/` con `0700`; archivo token con `0600`.

Stdout (parseable):

- `email\t<normalized_email>`
- `token_path\t<absolute_path>`

Errors:

- Uso invalido: exit code `2`.
- Credenciales faltantes: sugerir `dorean_g auth credentials <credentials.json>`.
- Error en exchange: `no se pudo intercambiar el token: ...`.
- Sin refresh token: error explicito de re-consentimiento.

## Account Resolution

- `GOG_ACCOUNT` define la cuenta activa para comandos Google API.
- Resolucion de token:
1. Si `GOG_ACCOUNT` esta seteado: leer token de esa cuenta.
2. Si no esta seteado: fallback legado a `token.json`.
- Si no hay token para `GOG_ACCOUNT`:
  - `token no encontrado para <email>, corre \`dorean_g auth add <email>\``.

## Compatibility

- `dorean_g auth` actual se mantiene (legacy path).
- `dorean_g auth --credentials <path>` se mantiene durante transicion.

## Out Of Scope (MVP)

- `dorean_g auth list/remove/status`
- keyring del sistema
- multi-client (`--client`)
