// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: query.sql

package users

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(id, username, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5) RETURNING id, username, email, first_name, last_name
`

type CreateUserParams struct {
	ID        int32
	Username  string
	Email     string
	FirstName string
	LastName  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Username,
		arg.Email,
		arg.FirstName,
		arg.LastName,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.FirstName,
		&i.LastName,
	)
	return i, err
}

const createUserProfileImage = `-- name: CreateUserProfileImage :exec
INSERT INTO user_profile_images(user_id, image_id) 
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
SET user_id = $1, image_id = $2
`

type CreateUserProfileImageParams struct {
	UserID  int32
	ImageID uuid.UUID
}

func (q *Queries) CreateUserProfileImage(ctx context.Context, arg CreateUserProfileImageParams) error {
	_, err := q.db.ExecContext(ctx, createUserProfileImage, arg.UserID, arg.ImageID)
	return err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteUser, id)
	return err
}

const getUserById = `-- name: GetUserById :one
SELECT id, username, email, first_name, last_name
FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserById(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.FirstName,
		&i.LastName,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, email, first_name, last_name
FROM users
WHERE username = $1 LIMIT 1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.FirstName,
		&i.LastName,
	)
	return i, err
}

const getUserProfileImage = `-- name: GetUserProfileImage :one
SELECT image_id FROM user_profile_images
WHERE user_id = $1
`

func (q *Queries) GetUserProfileImage(ctx context.Context, userID int32) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getUserProfileImage, userID)
	var image_id uuid.UUID
	err := row.Scan(&image_id)
	return image_id, err
}

const getUserView = `-- name: GetUserView :one
SELECT username, email, first_name, last_name
FROM users
WHERE id = $1 LIMIT 1
`

type GetUserViewRow struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
}

func (q *Queries) GetUserView(ctx context.Context, id int32) (GetUserViewRow, error) {
	row := q.db.QueryRowContext(ctx, getUserView, id)
	var i GetUserViewRow
	err := row.Scan(
		&i.Username,
		&i.Email,
		&i.FirstName,
		&i.LastName,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, username, email, first_name, last_name
FROM users
ORDER BY id
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUser = `-- name: UpdateUser :one
UPDATE users
set username   = $2,
    email      =$3,
    first_name = $4,
    last_name  = $5
WHERE id = $1 RETURNING id, username, email, first_name, last_name
`

type UpdateUserParams struct {
	ID        int32
	Username  string
	Email     string
	FirstName string
	LastName  string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUser,
		arg.ID,
		arg.Username,
		arg.Email,
		arg.FirstName,
		arg.LastName,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.FirstName,
		&i.LastName,
	)
	return i, err
}
