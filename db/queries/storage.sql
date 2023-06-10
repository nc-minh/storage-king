-- name: CreateStorage :one
INSERT INTO storage (
    access_token,
    refresh_token,
    email
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: UpdateStorage :one
UPDATE storage
SET
    access_token = COALESCE(sqlc.narg(access_token), access_token),
    refresh_token = COALESCE(sqlc.narg(refresh_token), refresh_token),
    is_refresh_token_expired = COALESCE(sqlc.narg(is_refresh_token_expired), is_refresh_token_expired),
    access_token_expires_in = COALESCE(sqlc.narg(access_token_expires_in), access_token_expires_in)
WHERE id = $1
RETURNING *;

-- name: ListStorage :many
SELECT * FROM storage;

-- name: GetStorage :one
SELECT * FROM storage
WHERE id = $1 OR email = $2
LIMIT 1;