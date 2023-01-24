package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/identity-service/helpers"
	"github.com/bernardn38/socialsphere/identity-service/imageServiceBroker"
	"github.com/bernardn38/socialsphere/identity-service/sql/users"
	"github.com/bernardn38/socialsphere/identity-service/token"
	"github.com/google/uuid"
)

type Handler struct {
	UserDb       *users.Queries
	TokenManager *token.Manager
	Emitter      *imageServiceBroker.Emitter
}

type Post struct {
	Body       string    `json:"body" validate:"required"`
	Author     int       `json:"author" validate:"required"'`
	AuthorName string    `json:"authorName" validate:"required"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserForm struct {
	UserId         int32     `json:"userId"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	ProfileImageId uuid.UUID `json:"profileImageId"`
}

func (handler *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// parse user form from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	userForm := UserForm{}
	json.Unmarshal(body, &userForm)

	//create new user in database
	_, err = handler.UserDb.CreateUser(context.Background(), users.CreateUserParams{ID: userForm.UserId,
		Username: userForm.Username, FirstName: userForm.Username, LastName: userForm.LastName, Email: userForm.Email})
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	helpers.ResponseNoPayload(w, 201)
}
func (handler *Handler) CreateUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadGateway)
		return
	}
	//parse image from form
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, h, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// create profile image association in database
	imageId := uuid.New()
	if err = handler.UserDb.CreateUserProfileImage(context.Background(), users.CreateUserProfileImageParams{
		UserID:  userId,
		ImageID: imageId,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//send image to rabbitmq for processing and upload to s3 bucket
	if file != nil {
		err = SendImageToQueue(file, handler, imageId, h.Header.Get("Content-Type"))
		if err != nil {
			log.Println(err)
		}
	}

	helpers.ResponseNoPayload(w, http.StatusCreated)
}
func (handler *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing return error
	userId, err := helpers.GetUserIdFromRequest(r, false)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	//get from database
	user, err := handler.UserDb.GetUserById(context.Background(), userId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	//respond with json of user data
	jsonResponse, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, 200, []byte(jsonResponse))
}

func (handler *Handler) GetUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	//get image id from datbase for specified user id
	imageId, err := handler.UserDb.GetUserProfileImage(context.Background(), userId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// make request to image service
	token, _ := r.Cookie("jwtToken")
	newReq, _ := http.NewRequest("GET", fmt.Sprintf("http://image-service:8080/image/%s", imageId), nil)
	newReq.AddCookie(token)
	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//read body and send to client
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "max-age=86400") //
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}
