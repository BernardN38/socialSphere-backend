package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/post-service/helpers"
	"github.com/bernardn38/socialsphere/post-service/models"
	"github.com/bernardn38/socialsphere/post-service/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	postService *service.PostService
}

func NewHandler(service *service.PostService) *Handler {
	return &Handler{
		postService: service,
	}
}

func (h *Handler) GetLikeCount(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
	}
	likeCount, err := h.postService.GetLikeCuountbyPostId(convertedPostId)
	if err != nil {
		log.Println(err)
		http.Error(w, "like count could not be calculated", http.StatusInternalServerError)
	}

	helpers.ResponseWithJson(w, 200, helpers.JsonResponse{Data: models.PostLikes{PostId: postId, LikeCount: likeCount}, Timestamp: time.Now()})
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	//get user id from request
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	//get username from jwt token
	username, ok := r.Context().Value("username").(string)
	if !ok {
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	//get post id from url path
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	//read comment body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	//unmarshal body to struct and validate
	var commentForm models.CreateCommentForm
	err = json.Unmarshal(bodyBytes, &commentForm)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	err = commentForm.Validate()
	if err != nil {
		log.Println(err)
		http.Error(w, "create comment form invalid", http.StatusBadRequest)
		return
	}
	err = h.postService.CreateCommentForPostId(commentForm, convertedPostId, userId, username)
	helpers.ResponseNoPayload(w, 201)
}

func (h *Handler) GetAllPostComments(w http.ResponseWriter, r *http.Request) {
	//get post id from url and convert to int32
	postId := chi.URLParam(r, "postId")
	parsedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}

	postComments, err := h.postService.GetCommentByPostId(parsedPostId)
	if err != nil {
		log.Println(err)
		http.Error(w, "could not find post comments", http.StatusInternalServerError)
		return
	}
	// convert to json then send to client
	jsonResp, err := json.Marshal(postComments)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	helpers.ResponseWithJson(w, http.StatusOK, helpers.JsonResponse{
		Data:      string(jsonResp),
		Timestamp: time.Now(),
	})
}
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	username := r.Context().Value("username").(string)
	convertedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
		return
	}
	body, file, header, err := GetBodyAndImage(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid post form", http.StatusBadRequest)
		return
	}

	var imageId = uuid.NullUUID{UUID: uuid.New(), Valid: file != nil}
	createPostForm := models.CreatPostForm{
		Body:       body,
		UserID:     convertedUserId,
		AuthorName: username,
		ImageID:    imageId,
	}
	_, err = h.postService.CreatPost(createPostForm, file, header)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, http.StatusCreated)
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	// get post id from url and convert to int32
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	post, err := h.postService.GetPostWithLikes(convertedPostId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	helpers.ResponseWithJson(w, 200, helpers.JsonResponse{
		Data:      post,
		Timestamp: time.Now(),
	})
}

func (h *Handler) GetPostsPageByUserId(w http.ResponseWriter, r *http.Request) {
	pageNo := r.URL.Query().Get("pageNo")
	pageSize := r.URL.Query().Get("pageSize")

	userId, err := helpers.GetUserIdFromRequest(r, false)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusUnauthorized)
		return
	}

	posts, err := h.postService.GetPostsByUserIdPaginated(userId, pageNo, pageSize)
	jsonResponse := helpers.JsonResponse{
		Data:      posts,
		Timestamp: time.Now(),
	}

	helpers.ResponseWithJson(w, 200, jsonResponse)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
		return
	}

	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	err = h.postService.DeletePost(convertedPostId, parsedUserId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.ResponseNoPayload(w, 200)
}
func (h *Handler) CreatePostLike(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
		return
	}
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	err = h.postService.CreatePostLike(convertedPostId, parsedUserId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
	helpers.ResponseNoPayload(w, 201)
}

func (h *Handler) DeleteLike(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
		return
	}
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	err = h.postService.DeletePostLike(convertedPostId, parsedUserId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, http.StatusOK)
}

func (h *Handler) CheckLike(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
		return
	}
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	isLiked, err := h.postService.CheckLike(convertedPostId, parsedUserId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, 200, []byte(fmt.Sprintf("%v", isLiked)))
}
