// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package users

import ()

type Friendship struct {
	ID      int32
	FriendA int32
	FriendB int32
}

type User struct {
	ID        int32
	UserID    int32
	Username  string
	Email     string
	FirstName string
	LastName  string
}
