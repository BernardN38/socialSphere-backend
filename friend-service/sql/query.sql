-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users(user_id, username, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: CreateFriendship :one
INSERT INTO friendships(friend_a, friend_b)
VALUES ($1, $2) RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
set username   = $2,
    email      =$3,
    first_name = $4,
    last_name  = $5
WHERE id = $1 RETURNING *;