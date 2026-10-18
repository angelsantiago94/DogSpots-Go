-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, name, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id, user_id, token_hash, family_id, expires_at, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE id = ? LIMIT 1;

-- name: ListRefreshTokensByFamily :many
SELECT * FROM refresh_tokens
WHERE family_id = ?;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET is_revoked = TRUE
WHERE id = ?;

-- name: RevokeAllTokensByFamily :exec
UPDATE refresh_tokens
SET is_revoked = TRUE
WHERE family_id = ?;
