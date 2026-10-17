# DogSpots

Aplicación Android para descubrir y compartir lugares donde ir con perros. El
contrato HTTP de la API vive en [`docs/openapi.yaml`](docs/openapi.yaml) y es la
referencia compartida entre el backend y la aplicación.

## Estructura

- `android/`: aplicación Android (el proyecto Gradle se crea en la fase 7).
- `dogspots-api/`: API en Go.
- `docs/`: contrato OpenAPI y documentación del proyecto.
- `docker-compose.yml`: entorno local de la API.

## Validar el contrato

```sh
npx --yes @redocly/cli lint docs/openapi.yaml
```

Consulta [CONTRIBUTING.md](CONTRIBUTING.md) para las convenciones de trabajo.