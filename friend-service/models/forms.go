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
	PostgresUrl      string        `validate:"required"`
	RabbitmqUrl      string        `validate:"required"`
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

type UserForm struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	UserId    int32  `json:"userId" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required"`
}

func (u *UserForm) Validate() error {
	validate := validator.New()
	err := validate.Struct(u)
	if err != nil {
		return err
	}
	return nil
}
