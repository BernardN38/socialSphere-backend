package handler

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/friend-service/helpers"
	"github.com/bernardn38/socialsphere/friend-service/rabbitmqBroker"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/bernardn38/socialsphere/friend-service/token"
)

type Handler struct {
	UsersDb      *users.Queries
	TokenManager *token.Manager
	Emitter      *rabbitmqBroker.Emitter
}
type UserForm struct {
	UserId    int32  `json:"userId" validate:"required"`
	Username  string `json:"username" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type UserFriendshipForm struct {
	FriendA int32 `json:"friendA" validate:"required"`
	FriendB int32 `json:"friendB" validate:"required"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// read and validate request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}
	userForm, err := ValidateUserForm(body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//create user in datbase
	_, err = h.UsersDb.CreateUser(context.TODO(), users.CreateUserParams{
		UserID:    userForm.UserId,
		Username:  userForm.Username,
		Email:     userForm.Email,
		FirstName: userForm.FirstName,
		LastName:  userForm.LastName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, http.StatusCreated)

}

func (h *Handler) CreateFrinedship(w http.ResponseWriter, r *http.Request) {
	// read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}
	friendshipForm, err := ValidateFriendshipForm(body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = h.UsersDb.CreateFriendship(context.TODO(), users.CreateFriendshipParams{
		FriendA: friendshipForm.FriendA,
		FriendB: friendshipForm.FriendB,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, 201)
}
