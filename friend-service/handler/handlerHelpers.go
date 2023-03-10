package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/friend-service/models"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/cristalhq/jwt/v4"
	"gopkg.in/go-playground/validator.v9"
)

func CreateUser(usersDb *users.Queries, form *models.UserForm) (int32, error) {
	user := users.CreateUserParams{
		UserID:    form.UserId,
		Username:  form.Username,
		Email:     form.Email,
		FirstName: form.FirstName,
		LastName:  form.LastName,
	}
	createdUserId, err := usersDb.CreateUser(context.Background(), user)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return createdUserId, nil
}

func ValidateUserForm(reqBody []byte) (models.UserForm, error) {
	var form models.UserForm
	err := json.Unmarshal(reqBody, &form)
	if err != nil {
		return models.UserForm{}, err
	}

	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return models.UserForm{}, err
	}

	return form, nil
}
func ValidateFriendshipForm(reqBody []byte) (models.UserFriendshipForm, error) {
	var form models.UserFriendshipForm
	err := json.Unmarshal(reqBody, &form)
	if err != nil {
		return models.UserFriendshipForm{}, err
	}

	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return models.UserFriendshipForm{}, err
	}

	return form, nil
}

func ValidateFindFriendsForm(r *http.Request) (models.FindFriendsForm, error) {
	params := r.URL.Query()
	form := models.FindFriendsForm{
		Username:  params.Get("username"),
		Email:     params.Get("email"),
		FirstName: params.Get("firstName"),
		LastName:  params.Get("lastName"),
	}
	v := validator.New()
	err := v.Struct(form)
	if err != nil {
		return models.FindFriendsForm{}, err
	}
	log.Println(params, form)
	return form, nil
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
