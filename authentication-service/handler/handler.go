package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/helpers"
	"github.com/bernardn38/socialsphere/authentication-service/models"
	"github.com/bernardn38/socialsphere/authentication-service/service"
	"github.com/bernardn38/socialsphere/authentication-service/token"
)

type Handler struct {
	AuthService  *service.AuthService
	TokenManager *token.Manager
}

func NewHandler(authService *service.AuthService, tm *token.Manager) *Handler {
	return &Handler{AuthService: authService, TokenManager: tm}
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
	_, err = h.AuthService.RegisterUser(registerForm)
	if err != nil {
		log.Println(err)
	}
	log.Println("Register successful username: ", registerForm.Username)
	helpers.ResponseWithPayload(w, 201, []byte(`Register Success`))
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "login body invalid", http.StatusBadRequest)
		return
	}
	var loginForm models.LoginForm
	err = json.Unmarshal(reqBody, &loginForm)
	if err != nil {
		log.Println(err)
		http.Error(w, "login body invalid", http.StatusBadRequest)
		return
	}
	err = loginForm.Validate()
	if err != nil {
		log.Println(err)
		http.Error(w, "login form invalid", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.LoginUser(loginForm)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}
	newToken, err := h.TokenManager.GenerateToken(fmt.Sprintf("%v", user.ID), user.Username, time.Minute*60)
	if err != nil {
		log.Println(err)
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	log.Println("Log in successful userId: ", user.ID)
	SetCookie(w, newToken)
	helpers.ResponseWithPayload(w, 200, []byte(fmt.Sprintf("%v", user.ID)))
}
