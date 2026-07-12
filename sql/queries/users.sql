-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;


-- name: UpdateUser :one
UPDATE users SET email = $1, password = $2 WHERE id = $3 RETURNING *;

-- name: UpdateUserChirpyRed :one
UPDATE users SET is_chirpy_red = true WHERE id = $1 RETURNING *;