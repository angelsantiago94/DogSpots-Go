-- name: CreateComment :one
INSERT INTO comments (
    id, place_id, author_id, body, parent_id, reply_to, edited, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetComment :one
SELECT * FROM comments WHERE id = ?;

-- name: GetCommentsByPlace :many
SELECT 
    c.*,
    u.id as author_id,
    u.name as author_name,
    u.email as author_email
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.place_id = ? AND c.parent_id IS NULL
ORDER BY c.created_at DESC;

-- name: GetReplies :many
SELECT 
    c.*,
    u.id as author_id,
    u.name as author_name,
    u.email as author_email
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.parent_id = ?
ORDER BY c.created_at ASC;

-- name: UpdateComment :exec
UPDATE comments
SET body = ?, edited = TRUE, updated_at = ?
WHERE id = ? AND author_id = ?;

-- name: DeleteCommentLogical :exec
UPDATE comments
SET deleted = TRUE, updated_at = ?
WHERE id = ? AND author_id = ?;

-- name: DeleteCommentPhysical :exec
DELETE FROM comments
WHERE id = ?;

-- name: IncrementReplyCount :exec
UPDATE comments
SET reply_count = reply_count + 1
WHERE id = ?;

-- name: DecrementReplyCount :exec
UPDATE comments
SET reply_count = reply_count - 1
WHERE id = ?;

-- name: IncrementCommentCount :exec
UPDATE places
SET comment_count = comment_count + 1
WHERE id = ?;

-- name: DecrementCommentCount :exec
UPDATE places
SET comment_count = comment_count - 1
WHERE id = ?;
