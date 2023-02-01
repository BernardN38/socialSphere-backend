package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/helpers"
	"github.com/bernardn38/socialsphere/authentication-service/models"
	rabbitmqBroker "github.com/bernardn38/socialsphere/authentication-service/rabbitmq_broker"
	rpcemitter "github.com/bernardn38/socialsphere/authentication-service/rpc_broker"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"github.com/bernardn38/socialsphere/authentication-service/token"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	UsersDb         *users.Queries
	TokenManager    *token.Manager
	RabbitMQEmitter *rabbitmqBroker.RabbitMQEmitter
	RpcEmitter      *rpcemitter.RpcClient
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
	emitter := rabbitmqBroker.NewEventEmitter(rabbitMQConn)

	handler := Handler{UsersDb: queries, TokenManager: tokenManger, RabbitMQEmitter: emitter}
	return &handler, nil
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "request body invalid", http.StatusBadRequest)
		return
	}
	var registerForm models.RegisterForm
	err = json.Unmarshal(reqBody, &registerForm)
	if err != nil {
		log.Println(err)
		http.Error(w, "request body invalid", http.StatusBadRequest)
		return
	}
	err = registerForm.Validate()
	if err != nil {
		log.Println(err)
		http.Error(w, "register form invalid", http.StatusBadRequest)
		return
	}
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(registerForm.Password), 12)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	registerForm.Password = string(encryptedPassword)
	createdUserId, err := CreateUser(h.UsersDb, registerForm)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	err = SendRabbitMQCreateUser(h.RabbitMQEmitter, registerForm, createdUserId)
	if err != nil {
		log.Println(err)
		rpcError := SendRpcCreateUser(h.RpcEmitter, h.RabbitMQEmitter, registerForm, createdUserId)
		log.Println(rpcError)
	}
	log.Println("Register successful username: ", registerForm.Username)
	helpers.ResponseWithPayload(w, 201, []byte(`Register Success`))
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "request body invalid", http.StatusBadRequest)
		return
	}
	var loginForm models.LoginForm
	err = json.Unmarshal(reqBody, &loginForm)
	if err != nil {
		log.Println(err)
		http.Error(w, "request body invalid", http.StatusBadRequest)
		return
	}
	err = loginForm.Validate()
	if err != nil {
		log.Println(err)
		http.Error(w, "register form invalid", http.StatusBadRequest)
		return
	}
	user, err := h.UsersDb.GetUserByUsername(context.Background(), loginForm.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginForm.Password))
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
