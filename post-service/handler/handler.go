package handler

import (
	"context"
	"fmt"
	"github.com/bernardn38/socialsphere/post-service/helpers"
	"github.com/bernardn38/socialsphere/post-service/imageServiceBroker"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/bernardn38/socialsphere/post-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	PostDb       *post.Queries
	TokenManager *token.Manager
	Emitter      *imageServiceBroker.Emitter
}

type Post struct {
	Body       string    `json:"body" validate:"required"`
	Author     uuid.UUID `json:"author" validate:"required"'`
	AuthorName string    `json:"authorName" validate:"required"`
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

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		return
	}
	body := r.MultipartForm.Value["body"]
	if len(body) < 1 {
		helpers.ResponseNoPayload(w, 400)
	}
	authorName := r.MultipartForm.Value["authorName"]
	if len(authorName) < 1 {
		helpers.ResponseNoPayload(w, 400)
	}
	file, err := r.MultipartForm.File["image"][0].Open()
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 400)
	}

	imageId := uuid.New()
	createdPost, err := handler.PostDb.CreatePost(context.Background(), post.CreatePostParams{
		Body:       body[0],
		Author:     parsedId,
		AuthorName: authorName[0],
		ImageID: uuid.NullUUID{
			UUID:  imageId,
			Valid: true,
		},
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}
	err = SendImageToQueue(file, handler, imageId)
	if err != nil {
		log.Println(err)
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
