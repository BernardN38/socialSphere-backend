-- name: GetAllImageIds :many
SELECT *
FROM user_images;

-- name: GetImageById :one
SELECT *
FROM user_images
WHERE id = $1 LIMIT 1;

-- name: GetImagesByUserId :many
SELECT *
FROM user_images
WHERE user_id = $1;

-- name: GetImagesByUserIdPaged :many
SELECT *
FROM user_images
WHERE user_id = $1
ORDER BY created_at limit $2 OFFSET $3;

-- name: CreateImage :one
INSERT INTO user_images(user_id, image_id)
VALUES ($1, $2 ) RETURNING *;

-- name: DeleteImagesByUserId :exec
DELETE
FROM user_images
WHERE user_id = $1;

-- name: DeleteImageById :exec
DELETE
FROM user_images
WHERE id = $1;


-- name: UpdateImage :one
UPDATE user_images
SET image_id = $2
WHERE id = $1 RETURNING *;