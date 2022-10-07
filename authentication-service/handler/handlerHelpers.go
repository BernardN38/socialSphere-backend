package handler

import (
	"context"
	"encoding/json"
	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
	"github.com/cristalhq/jwt/v4"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/http"
	"strings"
	"time"
)

func CreateUser(usersDb *users.Queries, form *RegisterForm) error {
	user := users.CreateUserParams{
		Username:  form.Username,
		Password:  form.Password,
		Email:     form.Email,
		FirstName: form.FirstName,
		LastName:  form.LastName,
	}
	_, err := usersDb.CreateUser(context.Background(), user)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
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
		Domain:     "localhost",
		Expires:    time.Now().Add(time.Minute * 5),
		RawExpires: "",
		MaxAge:     360,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   1,
		Raw:        "",
		Unparsed:   nil,
	}
	http.SetCookie(w, cookie)
	log.Println("cookie set", cookie)
}

func UpdateCookie(w http.ResponseWriter, handler *Handler, userId string) {
	newToken, err := handler.TokenManager.GenerateToken(userId)
	if err != nil {
		return
	}
	SetCookie(w, newToken)
}
