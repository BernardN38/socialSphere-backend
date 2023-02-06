package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/bernardn38/socialsphere/friend-service/helpers"
	"github.com/bernardn38/socialsphere/friend-service/models"
	rabbitmqBroker "github.com/bernardn38/socialsphere/friend-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/bernardn38/socialsphere/friend-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
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

func NewHandler(config models.Config) (*Handler, error) {
	//open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		return nil, err
	}
	// init sqlc user queries
	queries := users.New(db)

	//init jwt token manager
	tokenManger := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)

	//init rabbitmq message emitter
	rabbitMQConn := rabbitmqBroker.ConnectToRabbitMQ(config.RabbitmqUrl)
	emitter, err := rabbitmqBroker.NewEventEmitter(rabbitMQConn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	handler := Handler{UsersDb: queries, TokenManager: tokenManger, Emitter: &emitter}
	return &handler, nil
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

func (h *Handler) CreateFollow(w http.ResponseWriter, r *http.Request) {
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
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
	}

	err = h.UsersDb.CreateFollow(context.Background(), users.CreateFollowParams{
		FriendA: int32(userIdi64),
		FriendB: int32(friendIdi64),
	})
	var duplicateEntryError = &pq.Error{Code: "23505"}
	if errors.As(err, &duplicateEntryError) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
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
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}
	respPayload, err := json.Marshal(users)
	if err != nil {
		http.Error(w, "error writing json resp", http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, http.StatusOK, respPayload)
}

func (h *Handler) CheckFollow(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	friendB := chi.URLParam(r, "friendId")
	friendBi64, err := strconv.ParseInt(friendB, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	followStatus, err := h.UsersDb.CheckFollow(context.Background(), users.CheckFollowParams{FriendA: int32(userIdi64), FriendB: int32(friendBi64)})
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusNotFound)
		return
	}
	helpers.ResponseWithPayload(w, http.StatusOK, []byte(fmt.Sprintf("%v", followStatus)))
}

func (h *Handler) GetFriendsLastestPhotos(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	userFollows, err := h.UsersDb.GetLatestPhotos(context.Background(), users.GetLatestPhotosParams{
		FriendA: int32(userIdi64),
		Limit:   3,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "no follows found", http.StatusNotFound)
		return
	}
	resp, err := json.Marshal(userFollows)
	if err != nil {
		log.Println(err)
	}
	log.Println(userFollows)
	helpers.ResponseWithPayload(w, http.StatusOK, resp)
}
