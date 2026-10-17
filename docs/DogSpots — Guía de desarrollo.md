# DogSpots — Guía de desarrollo

Versión 0.1 · Octubre 2026 · Prototipo (MVP)

## 1. Objetivo y alcance del MVP

App Android donde los usuarios marcan en un mapa los mejores lugares para llevar a su perro, con fotos, etiquetas, valoraciones y comentarios con respuestas.

**Dentro del MVP**

- Registro, login y sesión con JWT (access + refresh con rotación)
- Mapa con lugares cercanos, filtros por categoría y etiquetas
- Crear, editar y borrar lugares propios, con fotos
- Valoración de 1 a 5 (una por usuario y lugar)
- Comentarios con respuestas (un nivel de anidación)
- Denuncia de contenido y borrado de cuenta (requisitos de Google Play)

**Fuera del MVP (backlog)**

Verificación de email, recuperación de contraseña, favoritos, perfiles de perro, moderación/administración, avatar, notificaciones, inicio de sesión con Google, versión iOS.

## 2. Stack acordado

| Capa | Elección |
| --- | --- |
| Android | Kotlin, Jetpack Compose, MapLibre, Retrofit/OkHttp, kotlinx.serialization, Coil, Hilt, MVVM |
| Mapas | OpenFreeMap (prototipo) → PMTiles propio |
| Backend | Go, `net/http` (1.22+), `sqlc`, `goose`, `golang-jwt/jwt/v5` |
| Base de datos | SQLite en modo WAL (driver `modernc.org/sqlite`) → Postgres + PostGIS |
| Fotos | Volumen Docker → S3/R2 |
| Despliegue | Docker Compose detrás de Nginx en un mini PC (dominio y certificado ya configurados) |
| CI | GitHub Actions |

El contrato de la API está en `openapi.yaml` y es la fuente de verdad entre backend y app.

## 3. Cómo usar esta guía

- Las tareas están agrupadas en fases. Cada tarea tiene un identificador (por ejemplo `B-3.2`) para poder referenciarla en issues y commits.
- Prefijos: **B** = backend, **A** = Android, **I** = infraestructura, **P** = producto/publicación.
- Las estimaciones son orientativas, para una persona con dedicación sustancial. Úsalas para ordenar, no como compromiso.
- Cada tarea se considera terminada cuando cumple la *definición de hecho* de la sección 5.

## 4. Plan por fases

### Fase 0 — Preparación (1–2 días)

- P-0.1 Crear el repositorio monorepo con la estructura `android/`, `dogspots-api/`, `docs/`, `docker-compose.yml`.
- P-0.2 Añadir `openapi.yaml` a `docs/` y validarlo con una herramienta (Swagger Editor o `redocly lint`).
- P-0.3 Definir convenciones: ramas (`main` + ramas de feature), mensajes de commit, plantilla de pull request.
- P-0.4 Instalar herramientas: Go 1.22+, Docker, Android Studio, `sqlc`, `goose`.

### Fase 1 — Esqueleto del backend (3–4 días)

- B-1.1 Inicializar el módulo Go y la estructura `cmd/api`, `internal/{auth,places,photos,comments,ratings,users}`, `db/`.
- B-1.2 Configuración por variables de entorno (puerto, ruta de la BD, secreto JWT, directorio de fotos, orígenes permitidos) con validación al arrancar.
- B-1.3 Servidor con `net/http`, apagado ordenado (graceful shutdown) y endpoint `GET /api/v1/healthz`.
- B-1.4 Middlewares: identificador de petición, logging estructurado (`log/slog`), recuperación de pánicos, límite de tamaño de cuerpo, CORS si procede.
- B-1.5 Formato de errores RFC 7807 con helper común y códigos estables (`code`).
- B-1.6 Esquema inicial con migraciones `goose`: `users`, `refresh_tokens`, `places`, `tags`, `place_tags`, `photos`, `comments`, `ratings`, `reports`. IDs UUID generados en la aplicación; fechas en UTC.
- B-1.7 Conexión SQLite con WAL, `busy_timeout` y claves foráneas activadas; una única conexión de escritura.
- B-1.8 Configurar `sqlc` y generar el primer paquete de consultas.
- I-1.9 `Dockerfile` multi-etapa (binario estático, imagen mínima) y `docker-compose.yml` con volúmenes para la BD y las fotos.
- I-1.10 Documentar la conexión del Nginx del mini PC con la API en Docker Compose; el dominio y el certificado de producción ya están configurados.
- I-1.11 Workflow de GitHub Actions: `go vet`, `go test`, build de la imagen.

**Hito:** `docker compose up` levanta la API y `healthz` responde.

### Fase 2 — Autenticación (4–5 días)

- B-2.1 Hash de contraseñas con argon2id (parámetros configurables).
- B-2.2 `POST /auth/register` con validación (email, contraseña mínima de 10 caracteres, nombre).
- B-2.3 Emisión de access token (15 min) con `golang-jwt` y firma HS256.
- B-2.4 Refresh tokens: generar aleatorio, guardar solo su hash, asociarlo a una familia de sesión.
- B-2.5 `POST /auth/login` y `POST /auth/refresh` con rotación; si se reutiliza un token ya rotado, revocar toda la familia.
- B-2.6 `POST /auth/logout`.
- B-2.7 Middleware de autenticación que inyecta el usuario en el contexto; variante opcional para endpoints públicos que mejoran con sesión (por ejemplo `my_rating`).
- B-2.8 Rate limiting en login y registro (por IP y por email).
- B-2.9 `GET/PATCH /users/me` y `GET /users/{id}`.
- B-2.10 Tests: registro duplicado, credenciales erróneas, token caducado, rotación y reutilización de refresh token.

**Hito:** flujo completo de registro, login, renovación y logout probado con tests.

### Fase 3 — Lugares y etiquetas (5–6 días)

- B-3.1 Sembrar el catálogo de etiquetas por migración (zona sin correa, agua, sombra, vallado, bolsas, terraza que admite perros, etc.).
- B-3.2 `GET /tags`.
- B-3.3 `POST /places` con validación de coordenadas, categoría y etiquetas del catálogo.
- B-3.4 `GET /places/{id}` con autor, etiquetas, fotos y agregados.
- B-3.5 `PATCH` y `DELETE /places/{id}` con comprobación de autoría (403 si no es el autor).
- B-3.6 `GET /places` por bounding box, con índice sobre latitud/longitud.
- B-3.7 `GET /places` por radio (bounding box previo + cálculo de distancia con Haversine) y `sort=distance`.
- B-3.8 Filtros: categoría, etiquetas (todas), texto, `min_rating`; paginación por cursor.
- B-3.9 `GET /users/me/places`.
- B-3.10 Tests de integración con SQLite en memoria. Datos de prueba (seed) para desarrollo, con lugares reales de tu ciudad.

**Hito:** puedes crear y consultar lugares con `curl` o Bruno.

### Fase 4 — Fotos (3–4 días)

- B-4.1 `POST /places/{id}/photos` multipart con límite de 5 MB y máximo de 8 fotos por lugar.
- B-4.2 Validar el tipo real del fichero (detección por contenido, no por extensión): JPEG, PNG, WebP.
- B-4.3 Eliminar metadatos EXIF (especialmente la geolocalización), reescalar la imagen y convertir a un formato eficiente.
- B-4.4 Generar miniatura.
- B-4.5 Guardar los ficheros con nombre aleatorio (UUID) tras una interfaz `Storage` con implementación en disco; así luego se cambia a S3/R2 sin tocar el resto.
- B-4.6 Servir `/media/*` con cabeceras de caché largas; configurar Nginx para entregar esos recursos si corresponde.
- B-4.7 `DELETE /places/{id}/photos/{photoId}` y limpieza de ficheros al borrar un lugar.
- B-4.8 Tests: fichero no válido, demasiado grande, EXIF eliminado.

### Fase 5 — Comentarios y valoraciones (4–5 días)

- B-5.1 `POST /places/{id}/comments` con `reply_to_comment_id` opcional y normalización a un nivel (`parent_id` = raíz, `reply_to` = autor del comentario contestado).
- B-5.2 `GET /places/{id}/comments` (primer nivel, con `reply_count`) y `GET /comments/{id}/replies`.
- B-5.3 `PATCH /comments/{id}` y marca `edited`.
- B-5.4 `DELETE /comments/{id}` con borrado lógico si tiene respuestas y borrado real si no.
- B-5.5 `PUT` y `DELETE /places/{id}/rating` con `rating_avg` y `rating_count` actualizados de forma consistente (recalcular en la misma transacción o columnas desnormalizadas).
- B-5.6 Incluir `comment_count` y `my_rating` en el detalle del lugar.
- B-5.7 Rate limiting en creación de comentarios para frenar spam.
- B-5.8 Tests: responder a una respuesta, borrar un comentario con respuestas, valorar dos veces (idempotencia).

### Fase 6 — Denuncias y cuenta (2–3 días)

- B-6.1 `POST /places/{id}/reports` y `POST /comments/{id}/reports`; evitar denuncias duplicadas del mismo usuario.
- B-6.2 `DELETE /users/me`: decidir y documentar qué se borra y qué se anonimiza (por ejemplo, comentarios con respuestas pasan a "eliminado"); revocar sesiones.
- B-6.3 Mecanismo mínimo para revisar denuncias (consulta SQL o un comando CLI; el panel de administración queda para después).

**Hito de backend:** todos los endpoints del contrato implementados. Revisar que lo implementado coincide con `openapi.yaml`.

### Fase 7 — Base de la app Android (4–5 días)

Puede empezar en paralelo a la Fase 2 una vez cerrado el contrato.

- A-7.1 Crear el proyecto (Kotlin, Compose, Material 3, `minSdk` razonable, por ejemplo 26).
- A-7.2 Estructura de módulos o paquetes por capa (`data`, `domain`, `ui`) y por funcionalidad. Hilt configurado.
- A-7.3 Cliente de red: Retrofit + OkHttp + kotlinx.serialization; interfaces generadas o escritas a mano a partir de `openapi.yaml`.
- A-7.4 Modelo de errores: parsear `application/problem+json` y mapear `code` a errores de dominio.
- A-7.5 Almacén de tokens cifrado (DataStore con Android Keystore o `EncryptedSharedPreferences`).
- A-7.6 Interceptor que añade el token y `Authenticator` que renueva con el refresh token al recibir un 401 (con exclusión mutua para que no haya renovaciones simultáneas).
- A-7.7 Navegación (Navigation Compose) con grafo de pantallas: autenticación, mapa, detalle, creación, perfil.
- A-7.8 Tema, tipografía, componentes comunes (botones, estados de carga, error y vacío).
- A-7.9 Pantallas de registro e inicio de sesión con validación en cliente.
- A-7.10 Configuración por *build variants*: `debug` apuntando a `10.0.2.2:8080` (emulador) o a tu IP local, y `release` al dominio real. Permitir tráfico HTTP solo en `debug`.

**Hito:** puedes registrarte e iniciar sesión desde la app contra tu backend local.

### Fase 8 — Mapa (5–6 días)

- A-8.1 Integrar MapLibre en Compose (SDK nativo en `AndroidView` o `maplibre-compose`; comprobar el estado de la librería antes de decidir).
- A-8.2 Cargar el estilo de OpenFreeMap y comprobar la atribución visible (obligatoria para OpenStreetMap y OpenFreeMap).
- A-8.3 Permiso de ubicación en tiempo de ejecución con explicación previa y manejo del rechazo; centrar el mapa con Fused Location Provider.
- A-8.4 Al terminar de mover el mapa (con *debounce*), pedir `GET /places` con el bounding box visible.
- A-8.5 Pintar marcadores con una capa de símbolos o círculos; agrupación (clustering) cuando haya muchos.
- A-8.6 Iconos por categoría.
- A-8.7 Hoja inferior con resumen del lugar al tocar un marcador (miniatura, título, valoración, distancia).
- A-8.8 Filtros (categoría, etiquetas, valoración mínima) y botón "mi ubicación".
- A-8.9 Caché en memoria de las consultas recientes y gestión de cancelaciones al mover el mapa rápido.
- A-8.10 Buscador de direcciones con Nominatim (respetando 1 petición por segundo y la política de uso) o dejarlo para después.

**Hito:** mapa funcional con lugares reales cargados desde el backend.

### Fase 9 — Crear y ver lugares (6–7 días)

- A-9.1 Pantalla de detalle: galería de fotos (Coil), descripción, etiquetas, autor, valoración media y mapa pequeño.
- A-9.2 Flujo de creación: elegir ubicación (pulsación larga en el mapa o ubicación actual), título, categoría, etiquetas, descripción.
- A-9.3 Selección de fotos con Photo Picker y/o cámara; comprimir y redimensionar en el cliente antes de subir.
- A-9.4 Subida de fotos en segundo plano con reintentos (WorkManager) tras crear el lugar; indicador de progreso y gestión de fallos parciales.
- A-9.5 Editar y borrar lugares propios.
- A-9.6 Pantalla "Mis lugares".
- A-9.7 Estados de error de red y modo sin conexión básico (mensajes claros, reintento).

### Fase 10 — Comentarios y valoraciones en la app (5–6 días)

- A-10.1 Selector de estrellas en el detalle del lugar (`PUT /rating`), mostrando `my_rating` y la media.
- A-10.2 Lista de comentarios con paginación (Paging 3 o paginación manual por cursor).
- A-10.3 Escribir un comentario y responder a otro comentario o respuesta; mostrar "respondiendo a @usuario".
- A-10.4 Expandir/contraer respuestas con carga bajo demanda desde `/replies`.
- A-10.5 Editar y borrar los propios; representar los comentarios eliminados con respuestas ("comentario eliminado").
- A-10.6 Menú de denuncia en lugares y comentarios.

### Fase 11 — Perfil, cuenta y pulido (3–4 días)

- A-11.1 Pantalla de perfil: nombre, editar nombre, cerrar sesión.
- A-11.2 Borrado de cuenta con confirmación (el enlace dentro de la app es un requisito de Google Play).
- A-11.3 Perfil público de otro usuario con sus lugares.
- A-11.4 Accesibilidad: descripciones de contenido, tamaños táctiles, contraste, modo oscuro.
- A-11.5 Textos en recursos de cadenas (`strings.xml`) con español e inglés si procede.
- A-11.6 Icono, pantalla de arranque (splash) y nombre final.

### Fase 12 — Despliegue (2–3 días)

- I-12.1 Preparar el mini PC para ejecutar la API con Docker Compose, con usuario no root, firewall y actualizaciones automáticas.
- I-12.2 Configurar el despliegue para integrarse con Nginx; el dominio, DNS y certificado de producción ya están configurados.
- I-12.3 Variables de entorno y secretos fuera del repositorio (fichero `.env` en el servidor con permisos restringidos).
- I-12.4 Despliegue con `docker compose pull && docker compose up -d`, idealmente automatizado desde GitHub Actions.
- I-12.5 **Copias de seguridad:** volcado diario del fichero SQLite (con `sqlite3 .backup`, no copiando el fichero en caliente) y de las fotos a otro destino; probar una restauración.
- I-12.6 Monitorización mínima: `healthz` con un servicio externo de *uptime* y rotación de logs.

### Fase 13 — Pruebas y beta cerrada (1–2 semanas)

- P-13.1 Tests de la capa de datos y de los *handlers* en el backend; objetivo razonable de cobertura en la lógica de negocio (auth, autoría, comentarios).
- P-13.2 Tests unitarios de ViewModels y algunos tests de UI en Android.
- P-13.3 Revisión de seguridad: secretos fuera del código, validación de entradas, límites de tamaño, permisos de autoría en cada endpoint, EXIF eliminado, cabeceras del servidor.
- P-13.4 Prueba en dispositivos reales y distintas versiones de Android; probar con mala conexión.
- P-13.5 Beta cerrada con 10–20 personas (dueños de perros reales) y sembrar lugares de tu zona para que el mapa no esté vacío.
- P-13.6 Recoger feedback, corregir y priorizar el backlog.

### Fase 14 — Publicación en Google Play (1 semana, más tiempos de revisión)

- P-14.1 Firma de la app con Play App Signing; conservar la clave de subida de forma segura.
- P-14.2 Política de privacidad pública (qué datos se recogen: email, ubicación, fotos) y formulario de "Seguridad de los datos".
- P-14.3 Declarar contenido generado por usuarios: mecanismo de denuncia y borrado de cuenta (ya implementados en las fases 6 y 11).
- P-14.4 Ficha de la tienda: capturas, descripción, icono, clasificación de contenido.
- P-14.5 Pasar por las pruebas cerradas que exige Google para cuentas personales nuevas (comprobar los requisitos vigentes antes de planificar fechas).
- P-14.6 Publicar en producción y activar un despliegue escalonado.

## 5. Definición de hecho

Una tarea está terminada cuando:

- El código está en `main` mediante pull request y la CI pasa.
- Tiene tests proporcionados a su riesgo (obligatorios en auth, permisos y lógica de comentarios).
- El comportamiento coincide con `openapi.yaml`; si no, se actualiza el contrato en el mismo PR.
- Los errores se devuelven y se muestran de forma coherente (formato RFC 7807 en el backend, mensajes comprensibles en la app).
- No hay secretos, claves ni datos personales en el repositorio.

## 6. Orden recomendado y paralelización

1. Fase 0 y Fase 1 primero, en orden.
2. Si trabajas solo, avanza por **cortes verticales**: completa auth en backend y luego en la app, después lugares en ambos, y así sucesivamente. Así siempre tienes algo que probar de extremo a extremo.
3. Si prefieres adelantar la app antes de que el backend esté listo, usa el contrato para generar respuestas simuladas (mock) y sustitúyelas después.
4. Despliega en el mini PC pronto (tras la Fase 3) y no esperes al final: detecta problemas de red y permisos cuando aún son baratos de arreglar.

**Resumen de tiempos (orientativo):** backend 3–4 semanas, Android 4–5 semanas (en parte solapadas), despliegue, pruebas y publicación 3–4 semanas. En total, unos 2,5 a 3 meses para una persona con dedicación alta, más si es a tiempo parcial.

## 7. Riesgos y mitigaciones

| Riesgo | Mitigación |
| --- | --- |
| El mapa vacío desanima a los primeros usuarios | Sembrar tú mismo decenas de lugares de tu ciudad antes de la beta |
| Contenido inapropiado o spam | Denuncias desde el día uno, rate limiting, revisión manual de denuncias |
| Fotos con ubicación exacta del usuario | Eliminar EXIF en el servidor, sin excepciones |
| Pérdida de datos (SQLite en un solo servidor) | Copias de seguridad automáticas y restauración probada |
| Concurrencia de escritura en SQLite | WAL, una conexión de escritura, `busy_timeout`; migrar a Postgres si crece |
| Límites de los servidores de teselas y geocodificación | Usar OpenFreeMap y caché; planificar PMTiles propio si hay tráfico |
| Robo de tokens | Access token corto, refresh con rotación, almacenamiento cifrado en el móvil |
| Requisitos de Google Play cambiantes | Revisar la documentación oficial al empezar la Fase 14, no antes |

## 8. Migración futura a Postgres

No es parte del MVP, pero esta lista evita sorpresas:

- Mantener todo el acceso a datos detrás de `sqlc` y de interfaces de repositorio.
- Evitar funciones específicas de SQLite en las consultas.
- Cuando llegue el momento: añadir el servicio Postgres (con PostGIS) a Docker Compose, crear migraciones equivalentes, regenerar `sqlc` con el dialecto de Postgres, migrar los datos con un script y sustituir el filtro por bounding box por consultas espaciales con índice GiST.
- Ensayar la migración con una copia de la base de datos antes de hacerla en producción.

## 9. Backlog posterior al MVP

- Verificación de email y recuperación de contraseña (requiere servicio de correo)
- Inicio de sesión con Google
- Favoritos y listas de lugares
- Perfiles de perro y filtros según tamaño o raza
- Panel de administración y moderación
- Notificaciones de respuestas a tus comentarios
- Búsqueda de direcciones propia (Photon)
- Mapas con PMTiles autoalojados y modo sin conexión
- Almacenamiento de fotos en S3/R2 con CDN
- Versión iOS o multiplataforma
