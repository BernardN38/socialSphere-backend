package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/models"
	rabbitmqBroker "github.com/bernardn38/socialsphere/authentication-service/rabbitmq_broker"
	rpcemitter "github.com/bernardn38/socialsphere/authentication-service/rpc_broker"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"github.com/cristalhq/jwt/v4"
)

func CreateUser(usersDb *users.Queries, form models.RegisterForm) (int32, error) {
	user := users.CreateUserParams{
		Username: form.Username,
		Password: form.Password,
		Email:    form.Email,
	}
	createdUser, err := usersDb.CreateUser(context.Background(), user)
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

func CheckForValidCookie(r *http.Request, h *Handler) (*jwt.RegisteredClaims, bool) {
	cookie, err := r.Cookie("jwtToken")
	if err != nil {
		log.Println(err)
		return nil, false
	}

	cookieFields := strings.Split(cookie.String(), "=")
	if len(cookieFields) != 2 {
		log.Println("Cookie invalid")
		return nil, false
	}
	token := cookieFields[1]
	claims, ok := h.TokenManager.VerifyToken(token)
	if !ok {
		return nil, false
	}
	return claims, true
}

func SetCookie(w http.ResponseWriter, token *jwt.Token) {
	cookie := &http.Cookie{
		Name:       "jwtToken",
		Value:      token.String(),
		Path:       "/",
		Domain:     "192.168.0.17",
		Expires:    time.Now().Add(time.Minute * 60),
		RawExpires: "",
		MaxAge:     3600,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   http.SameSiteLaxMode,
		Raw:        "",
		Unparsed:   nil,
	}
	http.SetCookie(w, cookie)
	log.Println("cookie set:", cookie)
}

func UpdateCookie(w http.ResponseWriter, h *Handler, userId string, username string) {
	newToken, err := h.TokenManager.GenerateToken(userId, username, time.Minute*60)
	if err != nil {
		return
	}
	SetCookie(w, newToken)
}
