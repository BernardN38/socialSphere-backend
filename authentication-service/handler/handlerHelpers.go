package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"github.com/cristalhq/jwt/v4"
	"gopkg.in/go-playground/validator.v9"
)

func CreateUser(usersDb *users.Queries, form *RegisterForm) (int32, error) {
	user := users.CreateUserParams{
		Username:  form.Username,
		Password:  form.Password,
		Email:     form.Email,
		FirstName: form.FirstName,
		LastName:  form.LastName,
	}
	createdUser, err := usersDb.CreateUser(context.Background(), user)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return createdUser.ID, nil
}

func ValidateRegisterForm(reqBody []byte) (*RegisterForm, error) {
	var form RegisterForm
	err := json.Unmarshal(reqBody, &form)
	if err != nil {
		return nil, err
	}

	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return nil, err
	}

	return &form, nil
}

func ValidateLoginForm(reqBody []byte) (*LoginForm, error) {
	var form LoginForm
	err := json.Unmarshal(reqBody, &form)
	if err != nil {
		return nil, err
	}

	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return nil, err
	}

	return &form, nil
}

func CheckForValidCookie(r *http.Request, handler *Handler) (*jwt.RegisteredClaims, bool) {
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
	claims, ok := handler.TokenManager.VerifyToken(token)
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
		SameSite:   1,
		Raw:        "",
		Unparsed:   nil,
	}
	http.SetCookie(w, cookie)
	log.Println("cookie set:", cookie)
}

func UpdateCookie(w http.ResponseWriter, handler *Handler, userId string, username string) {
	newToken, err := handler.TokenManager.GenerateToken(userId, username, time.Minute*60)
	if err != nil {
		return
	}
	SetCookie(w, newToken)
}
