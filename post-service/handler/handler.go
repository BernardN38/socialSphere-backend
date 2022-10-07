package handler

import (
	"context"
	"fmt"
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
	Body       string    `json:"body" validate:"required"`
	Author     uuid.UUID `json:"author" validate:"required"'`
	AuthorName string    `json:"authorName" validate:"required"`
	Image      []byte    `json:"image"`
	CreatedAt  time.Time `json:"created_at"`
}

func (handler *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedId, err := uuid.Parse(userId.(string))
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 500, []byte("user id is invalid"))
		return
	}
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
	createdPost, err := handler.PostDb.CreatePost(context.Background(), post.CreatePostParams{
		Body:       formData.Body,
		Author:     parsedId,
		AuthorName: formData.AuthorName,
		ImageID:    uuid.NullUUID{},
		CreatedAt:  time.Now(),
	})
	if err != nil {
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}
	helpers.ResponseWithPayload(w, http.StatusCreated, []byte(fmt.Sprintf(`{Post created with id: "%s"}`, createdPost.ID)))
}

func (handler *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	respPost, err := handler.PostDb.GetPostById(context.Background(), parsedId)
	if err != nil {
		helpers.ResponseNoPayload(w, 500)
		return
	}
	resp := helpers.JsonResponse{
		Data:      respPost,
		Timestamp: time.Now(),
	}
	helpers.ResponseWithJson(w, 200, resp)
}

func (handler *Handler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	pageNo := r.URL.Query().Get("pageNo")
	pageSize := r.URL.Query().Get("pageSize")

	parsedId, err := uuid.Parse(userId.(string))
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 500, []byte("user id is invalid"))
		return
	}
	limit, offset, err := ValidatePagination(pageSize, pageNo)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	posts, err := handler.PostDb.GetPostByUserIdPaged(context.Background(), post.GetPostByUserIdPagedParams{
		Author: parsedId,
		Limit:  limit,
		Offset: offset,
	})
	log.Println(posts)
	respPage := helpers.PageResponse{
		Page:     posts,
		PageSize: len(posts),
		PageNo:   offset + 1,
	}
	jsonResponse := helpers.JsonResponse{
		Msg:       "",
		Data:      respPage,
		Timestamp: time.Now(),
	}
	helpers.ResponseWithJson(w, 200, jsonResponse)
}
