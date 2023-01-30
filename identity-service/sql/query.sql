-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
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

-- name: CreateUser :exec
INSERT INTO users(id, username, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5);

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

-- name: CreateUserProfileImage :exec
INSERT INTO user_profile_images(user_id, image_id) 
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
SET user_id = $1, image_id = $2;


-- name: GetUserProfileImage :one
SELECT image_id FROM user_profile_images
WHERE user_id = $1;