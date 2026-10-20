-- name: UpsertRating :one
INSERT INTO ratings (
    id, user_id, place_id, rating, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?
)
ON CONFLICT (user_id, place_id) DO UPDATE SET
    rating = EXCLUDED.rating,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetRating :one
SELECT * FROM ratings
WHERE user_id = ? AND place_id = ?;

-- name: DeleteRating :exec
DELETE FROM ratings
WHERE id = ?;
