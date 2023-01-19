package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/helpers"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"github.com/bernardn38/socialsphere/authentication-service/token"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	UsersDb      *users.Queries
	TokenManager *token.Manager
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

func (handler *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	form, err := ValidateRegisterForm(reqBody)
	if err != nil {
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	if err != nil {
		helpers.ResponseNoPayload(w, 500)
		return
	}

	form.Password = string(encryptedPassword)
	createdUserId, err := CreateUser(handler.UsersDb, form)
	if err != nil {
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	body := make(map[string]interface{})
	body["firstName"] = form.FirstName
	body["lastName"] = form.LastName
	body["userId"] = createdUserId
	body["username"] = form.Username
	body["email"] = form.Email

	reqData, err := json.Marshal(body)
	if err != nil {
		log.Panicln(err)
	}
	resp, err := http.Post("http://identity-service:8080/users", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		log.Println(err)
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(respBody))
	log.Println("Register successful username: ", form.Username)
	helpers.ResponseWithPayload(w, 200, []byte("Register Success"))
}

func (handler *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	form, err := ValidateLoginForm(reqBody)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	user, err := handler.UsersDb.GetUserByUsername(context.Background(), form.Username)
	if err != nil {
		helpers.ResponseWithPayload(w, 404, []byte("user not found"))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
	if err != nil {
		helpers.ResponseNoPayload(w, 401)
		return
	}

	newToken, err := handler.TokenManager.GenerateToken(fmt.Sprintf("%v", user.ID), user.Username, time.Minute*60)
	if err != nil {
		return
	}
	log.Println("Log in successful userId: ", user.ID)
	SetCookie(w, newToken)
	helpers.ResponseWithPayload(w, 200, []byte(fmt.Sprintf("%v", user.ID)))
}
