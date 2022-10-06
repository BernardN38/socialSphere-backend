-- noinspection SqlDialectInspectionForFile

-- noinspection SqlNoDataSourceInspectionForFile

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserPassword :one
SELECT password
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserView :one
SELECT username, email, first_name, last_name
FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users(username, password, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;


-- name: UpdateUser :one
UPDATE users
set username   = $2,
    password   = $3,
    email      =$4,
    first_name = $5,
    last_name  = $6
WHERE id = $1 RETURNING *;