-- name: GetAllPosts :many
SELECT *
FROM post;

-- name: GetPostById :one
SELECT *
FROM post
WHERE id = $1 LIMIT 1;

-- name: GetPostByIdWithLikes :one
SELECT p.id, p.body, p.user_id, COUNT(pl.user_id)
FROM post p 
LEFT JOIN post_like pl
ON p.id = pl.post_id
WHERE id = $1 GROUP BY p.id LIMIT 1;

-- name: GetPostsByUserId :many
SELECT *
FROM post
WHERE user_id = $1;

-- name: GetPostByUserIdPaged :many
SELECT p.id, p.author_name, p.body, p.user_id, p.created_at, p.image_id, COUNT(pl.user_id) as LikeCount
FROM post p  
LEFT JOIN post_like pl 
ON p.id = pl.post_id
WHERE p.user_id = $1
GROUP BY p.id 
ORDER BY p.created_at 
LIMIT $2 
OFFSET $3;

-- name: CreatePost :one
INSERT INTO post(body,user_id,author_name,image_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeletePostByUserId :exec
DELETE
FROM post
WHERE user_id = $1;

-- name: DeletePostById :one
DELETE
FROM post
WHERE id = $1 AND user_id = $2 RETURNING image_id;

-- name: UpdatePost :one
UPDATE post
SET body = $2
WHERE id = $1 RETURNING *;

-- name: CheckLike :one
select exists(select 1 from post_like where post_id = $1 and user_id = $2) as e;


-- name: GetPostLikeCountById :one
SELECT count(*)
FROM post_like
WHERE post_id = $1;

-- name: CreatePostLike :one
INSERT INTO post_like(post_id,user_id)
VALUES ($1, $2) 
ON CONFLICT (post_id, user_id)  DO
UPDATE SET post_id = $1, user_id = $2 
RETURNING *;


-- name: DeletePostLike :exec
DELETE 
FROM post_like 
WHERE post_id = $1 and user_id = $2;

-- name: GetPostLike :one
SELECT *
FROM post_like
WHERE post_id = $1;

-- name: CreateComment :one
INSERT INTO comment(body, user_id, author_name)
VALUES ($1,$2,$3) RETURNING id, user_id, author_name;

-- name: CreatePostComment :one
INSERT INTO post_comment(post_id, comment_id)
VALUES ($1,$2) RETURNING id;

-- name: GetCommentById :one
SELECT * FROM comment
WHERE id = $1;

-- name: GetAllPostCommentsByPostId :many
SELECT body, comment_id, user_id, author_name FROM post_comment p join comment c on p.comment_id = c.id WHERE post_id = $1;