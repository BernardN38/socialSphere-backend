package models

import "gopkg.in/go-playground/validator.v9"

type CreateUserForm struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserId    int32  `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

type Notification struct {
	UserId  int32  `json:"userId" validate:"required"`
	Payload string `json:"payload" validate:"required"`
}

func (c *Notification) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type FollowNotificaitonPayload struct {
	Follower         int32  `json:"follower"`
	FollowerUsername string `json:"followerUsername"`
	Followed         int32  `json:"followed"`
	MessageType      string `json:"type" validate:"required"`
}

type UserFriendshipForm struct {
	FriendA int32 `json:"friendA" validate:"required"`
	FriendB int32 `json:"friendB" validate:"required"`
}

type FindFriendsForm struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
