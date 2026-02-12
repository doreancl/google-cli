# Releasing

Este documento define el flujo real para publicar en este repo.

## Referencias

- Base general de repo/proceso: [gogcli](https://github.com/steipete/gogcli)
- Solo para coverage/lint gate: [spogo](https://github.com/steipete/spogo)
- Script de coverage tomado como referencia: [scripts/check-coverage.sh](https://github.com/steipete/spogo/blob/main/scripts/check-coverage.sh)

## 1) Verify build is green
```sh
make ci
```

## Gate obligatorio antes de release

Ejecuta:

```bash
make ci
```

`make ci` incluye:

1. `make fmt-check`
2. `make lint`
3. `make test`
4. coverage gate (interno)

## Coverage (regla actual)

- Umbral por default: `90`
- Implementación: `scripts/check-coverage.sh`
- Configuración del umbral en Make: `COVERAGE_THRESHOLD ?= 90`

Comandos útiles:

```bash
make coverage                           # solo reporte, no falla por umbral
make ci                                 # validación completa
```

## Hooks

- `pre-commit`: `fmt-check`, `lint`, `test`
- `pre-push`: `test`, `coverage-check`

## Checklist de release

1. Confirmar cambios listos en `CHANGELOG.md`.
2. Correr `make ci` y dejar todo en verde.
3. Confirmar que no hay credenciales/tokens en cambios staged.
4. Push de la rama y esperar CI remoto en verde.
5. Crear tag/release.
