package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/models"
	rabbitmqBroker "github.com/bernardn38/socialsphere/authentication-service/rabbitmq_broker"
	rpcemitter "github.com/bernardn38/socialsphere/authentication-service/rpc_broker"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
)

func CreateUser(usersDb *users.Queries, form models.RegisterForm) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	user := users.CreateUserParams{
		Username: form.Username,
		Password: form.Password,
		Email:    form.Email,
	}
	createdUser, err := usersDb.CreateUser(ctx, user)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return createdUser.ID, nil
}

func SendRpcCreateUser(rpcEmitter *rpcemitter.RpcClient, rabbitmqEmitter *rabbitmqBroker.RabbitMQEmitter, form models.RegisterForm, userId int32) error {
	user := models.CreateUserParams{
		FirstName: form.FirstName,
		LastName:  form.LastName,
		UserId:    int32(userId),
		Username:  form.Username,
		Email:     form.Email,
	}
	err1 := rpcEmitter.CreateIdentityServiceUser(user)
	err2 := rpcEmitter.CreateFriendServiceUser(user)
	if err1 != nil || err2 != nil {
		return models.RpcCreateUserError{IdentityServiceError: err1, FriendServiceError: err2}
	}
	return nil
}

func SendRabbitMQCreateUser(rabbitMQEMitter *rabbitmqBroker.RabbitMQEmitter, form models.RegisterForm, userId int32) error {
	user := models.CreateUserParams{
		FirstName: form.FirstName,
		LastName:  form.LastName,
		UserId:    int32(userId),
		Username:  form.Username,
		Email:     form.Email,
	}
	jsonUser, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = rabbitMQEMitter.Push(jsonUser, "createUser", "application/json")
	if err != nil {
		return err
	}
	return nil
}
