package models

import (
	"github.com/cristalhq/jwt/v4"
	"gopkg.in/go-playground/validator.v9"
)

type Form interface {
	Validate() error
}

type Config struct {
	JwtSecretKey     string        `validate:"required"`
	JwtSigningMethod jwt.Algorithm `validate:"required"`
	MongoUri         string        `validate:"required"`
	Port             string        `validate:"required"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type Notification struct {
	UserId      int32  `json:"userId" validate:"required"`
	Payload     string `json:"payload" validate:"required"`
	MessageType string `json:"type" validate:"required"`
}

func (c *Notification) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type Message struct {
	FromUserId   int32  `json:"fromUserId" bson:"from_user_id"`
	FromUsername string `json:"fromUsername" bson:"from_username"`
	ToUserId     int32  `json:"toUserId" bson:"to_user_id"`
	Subject      string `json:"subject" bson:"subject"`
	Message      string `json:"message" bson:"message"`
}
