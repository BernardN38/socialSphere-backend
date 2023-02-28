package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/friend-service/models"
	"github.com/bernardn38/socialsphere/friend-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/lib/pq"
)

type FriendService struct {
	UsersDb *users.Queries
	Emitter *rabbitmq_broker.Emitter
}

func New(userDb *users.Queries, rabbitBroker *rabbitmq_broker.Emitter) *FriendService {
	return &FriendService{UsersDb: userDb, Emitter: rabbitBroker}
}

func (s *FriendService) CreateUser(userForm models.CreateUserForm) (int32, error) {
	// create user in datbase
	userId, err := s.UsersDb.CreateUser(context.Background(), users.CreateUserParams{
		UserID:    userForm.UserId,
		Username:  userForm.Username,
		Email:     userForm.Email,
		FirstName: userForm.FirstName,
		LastName:  userForm.LastName,
	})
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *FriendService) CreateUserFollow(userId int32, username string, friendId int32, cookie string) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.UsersDb.CreateFollow(ctx, users.CreateFollowParams{
		FriendA: userId,
		FriendB: friendId,
	})
	var duplicateEntryError = &pq.Error{Code: "23505"}
	if errors.As(err, &duplicateEntryError) {
		err := s.UsersDb.DeleteFollow(ctx, users.DeleteFollowParams{
			FriendA: userId,
			FriendB: friendId,
		})
		if err != nil {
			return 0, err
		}
	}
	payload := models.FollowNotificaitonPayload{Follower: userId, FollowerUsername: username, Followed: friendId, MessageType: "newFollow"}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	req, _ := http.NewRequest("POST", "http://notification-service:8080/api/v1/notifications/follow", bytes.NewBuffer(jsonBytes))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", cookie)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *FriendService) FindFriends(findFriendsForm models.FindFriendsForm) ([]users.GetUsersByFieldsRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	users, err := s.UsersDb.GetUsersByFields(ctx, users.GetUsersByFieldsParams{
		Username:  findFriendsForm.Username,
		Email:     findFriendsForm.Email,
		FirstName: findFriendsForm.FirstName,
		LastName:  findFriendsForm.LastName,
		Limit:     10,
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *FriendService) CheckFollow(userId int32, friendId int32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	followStatus, err := s.UsersDb.CheckFollow(ctx, users.CheckFollowParams{FriendA: userId, FriendB: friendId})
	if err != nil {
		return false, err
	}
	return followStatus, nil
}

func (s *FriendService) GetLatestFriendPhotos(userId int32) ([]users.GetLatestPhotosRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	userFollows, err := s.UsersDb.GetLatestPhotos(ctx, users.GetLatestPhotosParams{
		FriendA: userId,
		Limit:   3,
	})
	if err != nil {
		return nil, err
	}
	return userFollows, nil
}

func (s *FriendService) GetFriends(userId int32) ([]int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	friendsIds, err := s.UsersDb.GetUserFriendById(ctx, userId)
	if err != nil {
		return nil, err
	}
	return friendsIds, nil
}
