package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/post-service/helpers"
	"github.com/bernardn38/socialsphere/post-service/imageServiceBroker"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/bernardn38/socialsphere/post-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	pq "github.com/lib/pq"
)

type Handler struct {
	PostDb       *post.Queries
	TokenManager *token.Manager
	Emitter      *imageServiceBroker.Emitter
}

type Post struct {
	Body       string    `json:"body" validate:"required"`
	Author     int       `json:"author" validate:"required"'`
	AuthorName string    `json:"authorName" validate:"required"`
	CreatedAt  time.Time `json:"created_at"`
}

type CommentsResp struct {
	Body      string    `json:"body"`
	CommentId uuid.UUID `json:"comment_id"`
}
type PostLikes struct {
	PostId    string `json:"postId"`
	LikeCount int64  `json:"likeCount"`
}

func (handler *Handler) GetLikeCount(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
	}
	likeCount, err := handler.PostDb.GetPostLikeCountById(context.TODO(), convertedPostId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 404)
		return
	}

	helpers.ResponseWithJson(w, 200, helpers.JsonResponse{Data: PostLikes{PostId: postId, LikeCount: likeCount}, Timestamp: time.Now()})
}

func (handler *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 400)
		return
	}
	username, ok := r.Context().Value("username").(string)
	if !ok {
		helpers.ResponseNoPayload(w, 400)
		return
	}
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 400)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	var bodyJson map[string]interface{}
	err = json.Unmarshal(bodyBytes, &bodyJson)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}

	commentBody, ok := bodyJson["body"].(string)
	if !ok {
		log.Println("could not parse comment body")
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}

	createdComment, err := handler.PostDb.CreateComment(context.Background(), post.CreateCommentParams{Body: commentBody, UserID: userId, AuthorName: username})
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusInternalServerError)
		return
	}

	_, err = handler.PostDb.CreatePostComment(context.Background(), post.CreatePostCommentParams{
		PostID:    convertedPostId,
		CommentID: createdComment.ID,
	})
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithJson(w, 201, helpers.JsonResponse{Data: createdComment, Timestamp: time.Now()})
}

func (handler *Handler) GetAllPostComments(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "postId")
	parsedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 400)
		return
	}
	postsComments, err := handler.PostDb.GetAllPostCommentsByPostId(context.Background(), parsedPostId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusNotFound)
		return
	}
	jsonResp, err := json.Marshal(postsComments)
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
func (handler *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	username := r.Context().Value("username").(string)
	convertedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
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

	file, h, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	imageId := uuid.New()
	createdPost, err := handler.PostDb.CreatePost(context.Background(), post.CreatePostParams{
		Body:       body[0],
		UserID:     convertedUserId,
		AuthorName: username,
		ImageID: uuid.NullUUID{
			UUID:  imageId,
			Valid: true,
		},
	})
	if err != nil {
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}
	if file != nil {
		err = SendImageToQueue(file, handler, "image-proccessing", imageId, h.Header.Get("Content-Type"))
		if err != nil {
			log.Println(err)
		}
	}

	helpers.ResponseWithPayload(w, http.StatusCreated, []byte(fmt.Sprintf(`{Post created with id: "%v"}`, createdPost.ID)))
}

func (handler *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 400)
		return
	}
	respPost, err := handler.PostDb.GetPostByIdWithLikes(context.TODO(), convertedPostId)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.ResponseNoPayload(w, 404)
			return
		}
		helpers.ResponseNoPayload(w, 500)
		return
	}
	helpers.ResponseWithJson(w, 200, helpers.JsonResponse{
		Data:      respPost,
		Timestamp: time.Now(),
	})
}

func (handler *Handler) GetPostsPageByUserId(w http.ResponseWriter, r *http.Request) {
	pageNo := r.URL.Query().Get("pageNo")
	pageSize := r.URL.Query().Get("pageSize")

	userId, err := helpers.GetUserIdFromRequest(r, false)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 401)
		return
	}
	limit, offset, err := ValidatePagination(pageSize, pageNo)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	posts, err := handler.PostDb.GetPostByUserIdPaged(context.Background(), post.GetPostByUserIdPagedParams{
		UserID: userId,
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		helpers.ResponseNoPayload(w, 400)
	}
	var lastPage bool
	if len(posts) > int(limit) {
		lastPage = false
		posts = posts[:limit]
	} else {
		lastPage = true
	}
	respPage := helpers.PageResponse{
		Page:     posts,
		PageSize: len(posts),
		PageNo:   (offset / limit) + 1,
		LastPage: lastPage,
	}
	jsonResponse := helpers.JsonResponse{
		Data:      respPage,
		Timestamp: time.Now(),
	}
	helpers.ResponseWithJson(w, 200, jsonResponse)
}

func (handler *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
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
		helpers.ResponseNoPayload(w, 400)
		return
	}
	imageId, err := handler.PostDb.DeletePostById(context.Background(), post.DeletePostByIdParams{
		ID:     convertedPostId,
		UserID: parsedUserId,
	})
	if err != nil {
		helpers.ResponseWithJson(w, 500, helpers.JsonResponse{Msg: err.Error()})
		return
	}
	handler.Emitter.PushDelete(imageId.UUID.String())
	helpers.ResponseNoPayload(w, 200)
}
func (handler *Handler) CreatePostLike(w http.ResponseWriter, r *http.Request) {
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
		helpers.ResponseNoPayload(w, 400)
		return
	}
	_, err = handler.PostDb.CreatePostLike(context.Background(), post.CreatePostLikeParams{
		PostID: convertedPostId,
		UserID: parsedUserId,
	})
	var duplicateEntryError = &pq.Error{Code: "23505"}
	if errors.As(err, &duplicateEntryError) {
		helpers.ResponseNoPayload(w, 400)
		return
	}
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
	helpers.ResponseNoPayload(w, 201)
}

func (handler *Handler) DeleteLike(w http.ResponseWriter, r *http.Request) {
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
		helpers.ResponseNoPayload(w, 400)
		return
	}
	err = handler.PostDb.DeletePostLike(context.Background(), post.DeletePostLikeParams{
		UserID: parsedUserId,
		PostID: convertedPostId,
	})
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
	helpers.ResponseNoPayload(w, 200)
}

func (handler *Handler) CheckLike(w http.ResponseWriter, r *http.Request) {
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
		helpers.ResponseNoPayload(w, 400)
		return
	}
	isLiked, err := handler.PostDb.CheckLike(context.Background(), post.CheckLikeParams{
		PostID: convertedPostId,
		UserID: parsedUserId,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	helpers.ResponseWithPayload(w, 200, []byte(fmt.Sprintf("%v", isLiked)))
}
