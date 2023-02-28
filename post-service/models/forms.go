package models

import (
	"github.com/cristalhq/jwt/v4"
	"github.com/google/uuid"
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
	MinioKey         string        `validate:"required"`
	MinioSecret      string        `validate:"required"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type CreateCommentForm struct {
	Body string `json:"body" validate:"required"`
}

func (c *CreateCommentForm) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type CreatPostForm struct {
	Body             string
	UserID           int32
	AuthorName       string
	ImageID          uuid.NullUUID
	ImageContentType string
}

func (c *CreatPostForm) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}
