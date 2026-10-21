-- Revisión manual de denuncias (B-6.3)
--
-- Uso con el CLI:
--   docker compose exec api dogspots-reports -db /data/db/dogspots.db
--
-- Uso con sqlite3 dentro del contenedor:
--   docker compose exec -T api sqlite3 -header -column /data/db/dogspots.db < docs/moderation.sql
--
-- La acción sobre el contenido sigue siendo manual: no hay endpoint de moderación.
-- Borrar un lugar desde la app/API también elimina sus ficheros; hacerlo con SQL deja las fotos huérfanas.

-- 1. Listado de denuncias, de más reciente a más antigua.
SELECT
    r.created_at,
    CASE WHEN r.place_id IS NOT NULL THEN 'lugar' ELSE 'comentario' END AS tipo,
    COALESCE(r.place_id, r.comment_id) AS destino_id,
    COALESCE(p.title, substr(c.body, 1, 80), '(eliminado)') AS destino,
    r.reason,
    r.details,
    u.email AS denunciante
FROM reports r
JOIN users u ON u.id = r.user_id
LEFT JOIN places p ON p.id = r.place_id
LEFT JOIN comments c ON c.id = r.comment_id
ORDER BY r.created_at DESC;

-- 2. Lugares denunciados con su autor.
SELECT r.id AS report_id, r.created_at, p.id AS place_id, p.title, p.author_id, p.lat, p.lng
FROM reports r
JOIN places p ON p.id = r.place_id
ORDER BY r.created_at DESC;

-- 3. Comentarios denunciados con su autor y estado.
SELECT r.id AS report_id, r.created_at, c.id AS comment_id, c.place_id, c.author_id, c.body, c.deleted
FROM reports r
JOIN comments c ON c.id = r.comment_id
ORDER BY r.created_at DESC;
