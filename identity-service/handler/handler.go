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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	UserDb       *users.Queries
	TokenManager *token.Manager
	Emitter      *imageServiceBroker.Emitter
}

type Post struct {
	Body       string    `json:"body" validate:"required"`
	Author     uuid.UUID `json:"author" validate:"required"'`
	AuthorName string    `json:"authorName" validate:"required"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserForm struct {
	UserId         uuid.UUID `json:"userId"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	ProfileImageId uuid.UUID `json:"profileImageId"`
}

func (handler *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	userForm := UserForm{}
	json.Unmarshal(body, &userForm)

	handler.UserDb.CreateUser(context.Background(), users.CreateUserParams{ID: userForm.UserId,
		Username: userForm.Username, FirstName: userForm.Username, LastName: userForm.LastName, Email: userForm.Email})
	helpers.ResponseWithPayload(w, 200, []byte("form respose"))
}
func (handler *Handler) CreateUserProfileImage(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedUserId, err := uuid.Parse(userId.(string))
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 500, []byte("user id is invalid"))
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		return
	}
	imageId := uuid.New()
	file, h, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", h.Filename)
	fmt.Printf("File Size: %+v\n", h.Size)
	fmt.Printf("MIME Header: %+v\n", h.Header)

	err = handler.UserDb.CreateUserProfileImage(context.Background(), users.CreateUserProfileImageParams{
		UserID:  parsedUserId,
		ImageID: imageId,
	})
	if err != nil {
		helpers.ResponseWithPayload(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	if file != nil {
		err = SendImageToQueue(file, handler, imageId, h.Header.Get("Content-Type"))
		if err != nil {
			log.Println(err)
		}
	}

	helpers.ResponseNoPayload(w, http.StatusCreated)
}
func (handler *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	parsedId, err := uuid.Parse(userId)
	if err != nil {
		log.Println(err)
		return
	}
	user, err := handler.UserDb.GetUserById(context.Background(), parsedId)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 404, []byte(err.Error()))
		return
	}
	jsonResponse, err := json.Marshal(user)
	if err != nil {
		helpers.ResponseWithPayload(w, 500, []byte(jsonResponse))
	}
	helpers.ResponseWithPayload(w, 200, []byte(jsonResponse))
}

func (handler *Handler) GetUserProfileImage(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	if userId == "" {
		userId, _ = r.Context().Value("userId").(string)
	}
	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 400)
		return
	}
	imageId, err := handler.UserDb.GetUserProfileImage(context.Background(), parsedUserId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 404)
		return
	}
	token, err := r.Cookie("jwtToken")
	if err != nil {
		log.Println(err, "no token found identity service")
		helpers.ResponseNoPayload(w, 401)
		return
	}
	newReq, _ := http.NewRequest("GET", fmt.Sprintf("http://image-service:8080/image/%s", imageId), nil)
	newReq.AddCookie(token)
	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}

	w.Header().Set("Cache-Control", "max-age=2592000") //
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}
