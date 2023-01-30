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

type RegisterForm struct {
	Username  string `json:"username" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

func (c *RegisterForm) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type LoginForm struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (c *LoginForm) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}
