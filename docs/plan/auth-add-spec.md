# Spec: `gog auth add <email>`

## Goal

Implementar el flujo:

```bash
gog auth credentials ~/Downloads/client_secret.json
gog auth add you@gmail.com
export GOG_ACCOUNT=you@gmail.com
```

para acercar la UX al demo público de `gogcli`.

## Scope

- Nuevo subcomando: `auth add <email>`.
- Persistencia de refresh token por cuenta (email).
- Selección de cuenta activa vía `GOG_ACCOUNT` para llamadas API.
- Mantener compatibilidad con flujo actual (`gog auth`, token único legado).

## Non-goals (esta iteración)

- No keyring de sistema (seguimos en archivos locales).
- No `auth list`, `auth remove`, `auth status`.
- No clientes múltiples (`--client`).

## CLI contract

## Command

```bash
gog auth add <email>
```

## Validation

- Requiere exactamente 1 argumento posicional (`email`).
- `email` debe tener formato básico válido (contener `@`, sin espacios).
- Si falta/extra argumento: exit code `2`.

## Behavior

1. Carga OAuth config desde `ResolveCredentialsPath("")`.
2. Ejecuta el flujo OAuth (igual al actual: URL + código).
3. Intercambia código por token.
4. Verifica que `refresh_token` esté presente.
5. Guarda token bajo la cuenta indicada (`email`) en storage por cuenta.
6. Imprime confirmación parseable con `email` y `path`.

Salida ejemplo:

```text
Token guardado para you@gmail.com en /.../dorean_g/tokens/you@gmail.com.json
```

## Account selection

Para construir cliente API (`internal/googleapi/client.go`):

1. Si `GOG_ACCOUNT` está seteado: usar esa cuenta.
2. Si no está seteado:
   - fallback legado a `token.json` (comportamiento actual).

Error cuando `GOG_ACCOUNT` no existe en storage:

- `token no encontrado para <email>, corre \`gog auth add <email>\``

## Storage spec

## Paths

- Nuevo directorio: `filepath.Join(ConfigDir(), "tokens")`
- Archivo por cuenta: `tokens/<normalized-email>.json`
- Normalización email: `strings.ToLower(strings.TrimSpace(email))`

## File permissions

- Crear directorio con `0700`.
- Escribir token con `0600`.

## Legacy compatibility

- `TokenPath()` (`token.json`) se conserva para compatibilidad.
- `ReadToken()` actual se conserva.
- Se agregan funciones nuevas, sin romper API existente:
  - `TokenPathForAccount(email string) string`
  - `SaveTokenForAccount(email string, tok *oauth2.Token) error`
  - `ReadTokenForAccount(email string) (*oauth2.Token, error)`

## Command routing

## `auth` behavior after change

- `gog auth` sin subcomando mantiene el flujo actual (login/token legacy).
- `gog auth --credentials ...` se mantiene (compatibilidad).
- Nuevo parser:
  - `gog auth add <email>` -> nuevo flujo por cuenta.
  - `gog auth credentials <path>` -> spec anterior.

## Error model / exit codes

- Usage error: `ExitError{Code:2}`.
- Credenciales faltantes/invalidas: error con contexto de ruta.
- OAuth exchange fail: `no se pudo intercambiar el token: ...`.
- Token sin refresh token: error explícito sugiriendo reintento con consentimiento.
- Error de escritura/lectura: propagar con contexto.

## Testing plan

## Unit tests (`internal/cmd/auth_test.go`)

- `runAuth add`:
  - éxito guarda token con email normalizado.
  - email inválido -> usage/code 2.
  - sin argumento -> usage/code 2.
  - error de oauth -> falla.
  - token sin refresh -> falla.

## Unit tests (`internal/authstore/store_test.go`)

- `TokenPathForAccount` normaliza email.
- `SaveTokenForAccount` + `ReadTokenForAccount` roundtrip.
- lectura cuenta inexistente devuelve error.

## Integration/regression tests

- `gog auth` legacy sigue funcionando.
- `gog auth --credentials` sigue funcionando.
- `newHTTPClient` usa `GOG_ACCOUNT` cuando está seteado.
- `newHTTPClient` usa fallback `ReadToken()` cuando `GOG_ACCOUNT` está vacío.

## Docs changes

- README flujo recomendado:

```bash
gog auth credentials ~/Downloads/client_secret.json
gog auth add you@gmail.com
export GOG_ACCOUNT=you@gmail.com
```

- Aclarar que `GOG_ACCOUNT` controla la cuenta activa para comandos calendar.

## Acceptance criteria

- `gog auth add <email>` existe y guarda token por cuenta.
- `GOG_ACCOUNT=<email>` permite ejecutar llamadas usando ese token.
- El flujo legado no se rompe (`gog auth`, `token.json`).
