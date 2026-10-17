# Contribuir

## Ramas

- `main` contiene el trabajo integrado y revisado.
- Crea ramas desde `main` con el formato `<tipo>/<id>-<resumen>`; por ejemplo,
  `feature/B-2.1-password-hashing` o `docs/P-0.3-workflow`.
- Tipos habituales: `feature`, `fix`, `docs`, `chore` y `test`.
- Integra los cambios mediante pull request; no trabajes directamente sobre
  `main`.

## Commits

Usa el formato Conventional Commits: `tipo(ámbito): resumen`.

Tipos habituales: `feat`, `fix`, `docs`, `chore`, `test`, `refactor` y `ci`.
Escribe el resumen en imperativo y en minúsculas cuando encaje; por ejemplo,
`feat(auth): añadir rotación de refresh tokens`.

## Pull requests

- Abre el pull request contra `main` y explica el problema y la solución.
- Relaciona la tarea por su identificador de la guía (por ejemplo, `B-2.1`).
- Indica las pruebas ejecutadas y cualquier trabajo pendiente.
- Actualiza `docs/openapi.yaml` en el mismo pull request si cambia el contrato.
- No incluyas secretos, ficheros `.env`, datos personales ni artefactos generados.
- Completa la lista de comprobación de `.github/PULL_REQUEST_TEMPLATE.md`.