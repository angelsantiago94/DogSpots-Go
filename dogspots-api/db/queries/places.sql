-- name: GetTags :many
SELECT id, slug, name, created_at FROM tags ORDER BY name;

-- name: CreatePlace :exec
INSERT INTO places (id, author_id, title, description, lat, lng, category)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetPlace :one
SELECT id, author_id, title, description, lat, lng, category, created_at, updated_at FROM places
WHERE id = ?;

-- name: GetPlaceAuthor :one
SELECT id, name, email FROM users WHERE id = ?;

-- name: AddPlaceTag :exec
INSERT INTO place_tags (place_id, tag_id)
VALUES (?, ?);

-- name: DeletePlaceTag :exec
DELETE FROM place_tags
WHERE place_id = ? AND tag_id = ?;

-- name: UpdatePlace :exec
UPDATE places
SET title = ?, description = ?, lat = ?, lng = ?, category = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeletePlace :exec
DELETE FROM places WHERE id = ?;

-- name: ListPlacesByBoundingBox :many
SELECT id, author_id, title, description, lat, lng, category, created_at, updated_at FROM places
WHERE lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?
ORDER BY created_at DESC
LIMIT ?;

-- name: ListUserPlaces :many
SELECT id, author_id, title, description, lat, lng, category, created_at, updated_at FROM places
WHERE author_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: GetTagsByPlaceID :many
SELECT t.id, t.slug, t.name, t.created_at
FROM tags t
JOIN place_tags pt ON t.id = pt.tag_id
WHERE pt.place_id = ?;

-- name: GetPhotosByPlaceID :many
SELECT id, place_id, url, thumb_url, caption, width, height, position, uploaded_by, created_at
FROM photos
WHERE place_id = ?;
