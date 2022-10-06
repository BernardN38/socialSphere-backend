package handler

import (
	"context"
	"encoding/json"
	"github.com/bernardn38/socialsphere/post-service/helpers"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/bernardn38/socialsphere/post-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	PostDb       *post.Queries
	TokenManager *token.Manager
}

type Post struct {
	Body      string    `json:"body" validate:"required"`
	Author    uuid.UUID `json:"author" validate:"required"'`
	Image     []byte    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
}

func (handler *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("creating post")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		helpers.ResponseNoPayload(w, 500)
		return
	}

	formData, err := ValidatePostForm(body)
	if err != nil {
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	_, err = handler.PostDb.CreatePost(context.Background(), post.CreatePostParams{
		Body:      formData.Body,
		Author:    formData.Author,
		ImageID:   uuid.NullUUID{},
		CreatedAt: time.Now(),
	})
	if err != nil {
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}

	helpers.ResponseWithPayload(w, http.StatusCreated, []byte("Post created"))
}

func (handler *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "id")
	post, err := handler.PostDb.GetPostById(context.Background(), uuid.Must(uuid.Parse(postId)))
	if err != nil {
		helpers.ResponseNoPayload(w, 500)
		return
	}
	jsonPost, err := json.Marshal(post)
	if err != nil {
		helpers.ResponseNoPayload(w, 500)
		return
	}
	helpers.ResponseWithPayload(w, 200, jsonPost)
}
