package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/bernardn38/socialsphere/friend-service/helpers"
	rabbitmqBroker "github.com/bernardn38/socialsphere/friend-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/bernardn38/socialsphere/friend-service/token"
	"github.com/go-chi/chi/v5"
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

type FindFriendsForm struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
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
	_, err = h.UsersDb.CreateUser(context.Background(), users.CreateUserParams{
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

func (h *Handler) CreateFriendship(w http.ResponseWriter, r *http.Request) {
	friendId := chi.URLParam(r, "friendId")
	if len(friendId) < 1 {
		log.Println("friend id not found")
		http.Error(w, "friend id not provided", http.StatusBadRequest)
		return
	}
	friendIdi64, err := strconv.ParseInt(friendId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to int32")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	_, err = h.UsersDb.CreateFriendship(context.Background(), users.CreateFriendshipParams{
		FriendA: int32(userIdi64),
		FriendB: int32(friendIdi64),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, 201)
}

func (h *Handler) FindFriends(w http.ResponseWriter, r *http.Request) {
	findFriendsForm, err := ValidateFindFriendsForm(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, err := h.UsersDb.GetUsersByFields(context.Background(), users.GetUsersByFieldsParams{
		Username:  findFriendsForm.Username,
		Email:     findFriendsForm.Email,
		FirstName: findFriendsForm.FirstName,
		LastName:  findFriendsForm.LastName,
		Limit:     10,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "No Users Found", http.StatusNotFound)
		return
	}
	if len(users) == 0 {
		http.Error(w, "now users found", http.StatusNotFound)
		return
	}
	respPayload, err := json.Marshal(users)
	if err != nil {
		http.Error(w, "error writing json resp", http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, http.StatusOK, respPayload)
}
