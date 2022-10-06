
-- name: GetPostById :one
SELECT *
FROM post
WHERE id = $1 LIMIT 1;

-- name: GetPostsByUserId :many
SELECT *
FROM post
WHERE author = $1;

-- name: ListPosts :many
SELECT *
FROM post
ORDER BY id;

-- name: CreatePost :one
INSERT INTO post(body,author,image_id,created_at)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeletePostByUserId :exec
DELETE
FROM post
WHERE author = $1;

-- name: DeletePostById :exec
DELETE
FROM post
WHERE id = $1;


-- name: UpdatePost :one
UPDATE post
SET body = $2
WHERE id = $1 RETURNING *;