package service

import (
	"context"
	"log"

	"github.com/bernardn38/socialsphere/authentication-service/models"
	"github.com/bernardn38/socialsphere/authentication-service/rabbitmq_broker"
	rpcbroker "github.com/bernardn38/socialsphere/authentication-service/rpc_broker"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserDb          *users.Queries
	RabbitMQEmitter *rabbitmq_broker.RabbitMQEmitter
	RpcEmitter      *rpcbroker.RpcClient
}

func New(userDb *users.Queries, rabbitBroker *rabbitmq_broker.RabbitMQEmitter, rpcEmitter *rpcbroker.RpcClient) *AuthService {
	return &AuthService{
		UserDb:          userDb,
		RabbitMQEmitter: rabbitBroker,
		RpcEmitter:      rpcEmitter,
	}
}

func (s *AuthService) RegisterUser(registerForm models.RegisterForm) (int32, error) {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(registerForm.Password), 12)
	if err != nil {
		return 0, err
	}

	registerForm.Password = string(encryptedPassword)
	createdUserId, err := CreateUser(s.UserDb, registerForm)
	if err != nil {
		return 0, err
	}
	err = SendRabbitMQCreateUser(s.RabbitMQEmitter, registerForm, createdUserId)
	if err != nil {
		log.Println(err)
		rpcError := SendRpcCreateUser(s.RpcEmitter, s.RabbitMQEmitter, registerForm, createdUserId)
		log.Println(rpcError)
	}
	return createdUserId, nil
}

func (s *AuthService) LoginUser(loginForm models.LoginForm) (users.User, error) {
	user, err := s.UserDb.GetUserByUsername(context.Background(), loginForm.Username)
	if err != nil {
		return users.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginForm.Password))
	if err != nil {
		return users.User{}, err
	}

	return user, nil
}
