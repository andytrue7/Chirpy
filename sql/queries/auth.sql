-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;


-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1 RETURNING *;