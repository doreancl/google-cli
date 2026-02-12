# Spec: `gog auth credentials`

## Goal

Introducir un comando explícito para registrar credenciales OAuth:

```bash
gog auth credentials ~/Downloads/client_secret_....json
```

alineado con el flujo de referencia (`gogcli`) y manteniendo compatibilidad con el flujo actual basado en `auth --credentials`.

## Scope

- Nuevo subcomando: `auth credentials <path>`.
- Mantener soporte actual de `auth --credentials` (compatibilidad hacia atrás).
- No cambia el flujo de OAuth/token en esta iteración; solo cómo se define la ruta de credenciales.

## Non-goals (esta iteración)

- No se implementa `auth credentials list`.
- No se implementan clientes múltiples (`--client`, `--domain`).
- No se cambia el formato de almacenamiento de token.

## UX / CLI Contract

## Command

```bash
gog auth credentials <credentials.json>
```

## Behavior

- Recibe exactamente 1 argumento posicional: ruta al JSON de OAuth Client (Desktop app).
- Valida que el archivo exista y que el JSON sea parseable por `google.ConfigFromJSON` con `CalendarScope`.
- Copia el archivo al path administrado por la app:
  - `filepath.Join(ConfigDir(), "client_secret.json")`
- Crea el directorio si no existe (`0700`).
- Escribe el archivo destino con permisos `0600`.
- Imprime confirmación parseable en stdout con la ruta guardada.

Ejemplo de salida:

```text
Credentials guardadas en /.../dorean_g/client_secret.json
```

## Errors and exit codes

- Uso inválido (sin argumento o argumentos extra): exit code `2`.
- Archivo no encontrado / no legible: error con contexto de ruta.
- JSON inválido o no OAuth Desktop client: `credenciales invalidas: ...`.
- Error de escritura/permisos en destino: propagar error con contexto.

## Path rules

- Debe expandir `~` antes de leer el archivo de entrada.
- Debe aceptar ruta absoluta o relativa.
- Debe rechazar path vacío.

## Backward compatibility

Durante transición:

- `gog auth --credentials <path>` sigue funcionando.
- `gog auth` (sin flags) sigue usando resolución actual (`GCAL_CREDENTIALS` o default path).
- Recomendación en help/docs migra a `gog auth credentials <path>` + `gog auth`.

## Integration with current flow

### Antes

```bash
gog auth --credentials ~/Downloads/client_secret.json
```

### Después (preferido)

```bash
gog auth credentials ~/Downloads/client_secret.json
gog auth
```

`gog auth` continuará leyendo credenciales desde el path gestionado por `ResolveCredentialsPath("")`.

## Proposed internal changes

- `internal/cmd/root.go`
  - `auth` pasa de comando plano a grupo con subcomandos (`credentials`, `login`/default actual).
- `internal/cmd/auth.go`
  - separar:
    - `runAuthLogin` (lógica actual de OAuth + token)
    - `runAuthCredentials` (nuevo flujo de copy+validate)
- `internal/authstore/store.go`
  - agregar helper público `CredentialsPath()` para evitar duplicar `filepath.Join(ConfigDir(), "client_secret.json")`.
- `usage()`
  - actualizar ejemplos al nuevo flujo recomendado.

## Testing plan

## Unit tests

- `internal/cmd/auth_test.go`
  - `auth credentials <valid file>` guarda en destino y devuelve nil.
  - sin argumento -> `ExitError{Code:2}`.
  - con argumento extra -> `ExitError{Code:2}`.
  - archivo inexistente -> error.
  - json inválido -> error.
- `internal/authstore/store_test.go`
  - test para `CredentialsPath()`.

## Regression tests

- `auth --credentials <path>` sigue funcionando como hoy.
- `auth` sin flags sigue funcionando con `GCAL_CREDENTIALS` y default path.

## Docs changes

- `README.md`:
  - mover ejemplo principal a:
    - `gog auth credentials /ruta/client_secret.json`
    - `gog auth`
- `docs/google-calendar-spec.md`:
  - reflejar nuevo comando recomendado.

## Acceptance criteria

- Existe comando `gog auth credentials <path>` funcional.
- Credenciales quedan persistidas en `ConfigDir()/client_secret.json`.
- `gog auth` funciona sin `--credentials` después de ejecutar el comando nuevo.
- No rompe los flujos existentes (`--credentials`, `GCAL_CREDENTIALS`).
