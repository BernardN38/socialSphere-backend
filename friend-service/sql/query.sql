-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetFollowByFriendA :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByFirstName :one
SELECT *
FROM users
WHERE first_name = $1 LIMIT $2;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUsersByLastName :many
SELECT *
FROM users
WHERE last_name = $1 LIMIT $2;

-- name: GetUsersByFields :many
SELECT user_id, username, first_name, last_name
FROM users
WHERE username = $1 or email = $2 or first_name = $3 or last_name = $4 LIMIT $5;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users(user_id, username, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5) RETURNING user_id;

-- name: CreateFollow :exec
INSERT INTO follow(friend_a, friend_b)
VALUES ($1, $2) RETURNING *;

-- name: CheckFollow :one
SELECT exists( select 1 FROM follow WHERE friend_a = $1 AND friend_b = $2);

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

