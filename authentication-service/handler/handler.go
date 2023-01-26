package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/helpers"
	rabbitmqBroker "github.com/bernardn38/socialsphere/authentication-service/rabbitmq_broker"
	rpcemitter "github.com/bernardn38/socialsphere/authentication-service/rpc_emitter"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"github.com/bernardn38/socialsphere/authentication-service/token"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	UsersDb         *users.Queries
	TokenManager    *token.Manager
	RabbitMQEmitter *rabbitmqBroker.RabbitMQEmitter
	RpcEmitter      *rpcemitter.RpcEmitter
}
type RegisterForm struct {
	Username  string `json:"username" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}
type LoginForm struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	form, err := ValidateRegisterForm(reqBody)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	form.Password = string(encryptedPassword)
	createdUserId, err := CreateUser(h.UsersDb, form)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	user := rpcemitter.CreateUserParams{
		FirstName: form.FirstName,
		LastName:  form.LastName,
		UserId:    int32(createdUserId),
		Username:  form.Username,
		Email:     form.Email,
	}
	err = h.RpcEmitter.CreateIdentityServiceUser(user)
	if err != nil {
		log.Println(err, "identity")
	}
	err = h.RpcEmitter.CreateFriendServiceUser(user)
	if err != nil {
		log.Println(err, "friends")
	}
	log.Println("Register successful username: ", form.Username)
	helpers.ResponseWithPayload(w, 201, []byte(`Register Success`))
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	form, err := ValidateLoginForm(reqBody)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.UsersDb.GetUserByUsername(context.Background(), form.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	newToken, err := h.TokenManager.GenerateToken(fmt.Sprintf("%v", user.ID), user.Username, time.Minute*60)
	if err != nil {
		return
	}
	log.Println("Log in successful userId: ", user.ID)
	SetCookie(w, newToken)
	helpers.ResponseWithPayload(w, 200, []byte(fmt.Sprintf("%v", user.ID)))
}
