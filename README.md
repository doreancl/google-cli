# Google Calendar CLI (base)

CLI en Go para listar y crear eventos en Google Calendar, inspirado en `gogcli`.

## 1) Crear credenciales OAuth

1. Entra a Google Cloud Console.
2. Crea un proyecto (o usa uno existente).
3. Habilita **Google Calendar API**.
4. Configura **OAuth consent screen** (External o Internal).
5. Crea credenciales de tipo **OAuth Client ID** (Desktop app).
6. Descarga el JSON y guárdalo como:
   - macOS: `~/Library/Application Support/dorean_g/client_secret.json`
   - Linux: `~/.config/dorean_g/client_secret.json`
   - o donde quieras y usa `--credentials`.

## 2) Autenticar

```bash
./bin/dorean_g auth --credentials /ruta/client_secret.json
```

También puedes usar:

```bash
export GCAL_CREDENTIALS=/ruta/client_secret.json
./bin/dorean_g auth
```

Esto guarda el token en:
- macOS: `~/Library/Application Support/dorean_g/token.json`
- Linux: `~/.config/dorean_g/token.json`

## 3) Listar eventos

```bash
./bin/dorean_g events list --calendar primary --days 7
```

## 4) Crear evento

```bash
./bin/dorean_g events create \
  --calendar primary \
  --summary "Reunión de seguimiento" \
  --start "2026-02-12T10:00:00-06:00" \
  --end "2026-02-12T10:30:00-06:00" \
  --description "Agenda semanal"
```

Alias corto con Makefile:

```bash
make dg -- events list --calendar primary --days 7
```

## Desarrollo

```bash
make test
make coverage
make coverage-check            # umbral default 90%
make coverage-check THRESHOLD=75
make ci                        # fmt-check + lint + test + coverage-check
```

## Notas

- Formato de fecha/hora: RFC3339.
- Esta es una base inicial. Siguiente paso recomendado: añadir `update`, `delete`, filtros por texto, y salida JSON.
