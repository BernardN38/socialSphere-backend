package handler

import (
	"bytes"
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
	"github.com/bernardn38/socialsphere/post-service/models"
	"github.com/bernardn38/socialsphere/post-service/rabbitmq_broker"
	rpcbroker "github.com/bernardn38/socialsphere/post-service/rpc_broker"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/bernardn38/socialsphere/post-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	pq "github.com/lib/pq"
	"github.com/minio/minio-go"
)

type Handler struct {
	PostDb          *post.Queries
	TokenManager    *token.Manager
	RabbitMQEmitter *rabbitmq_broker.RabbitMQEmitter
	RpcClient       *rpcbroker.RpcClient
	MinioClient     *minio.Client
}

func NewHandler(config models.Config) (*Handler, error) {
	//open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		return nil, err
	}

	// init sqlc post queries
	queries := post.New(db)

	//init jwt token manager
	tokenManger := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)

	//init rabbitmq message emitter
	rabbitConn := rabbitmq_broker.ConnectToRabbitMQ(config.RabbitmqUrl)
	rabbitMQBroker, err := rabbitmq_broker.NewRabbitEventEmitter(rabbitConn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// connect to minio
	minioClient, err := minio.New("minio:9000", config.MinioKey, config.MinioSecret, false)
	if err != nil {
		return nil, err
	}
	handler := Handler{PostDb: queries, TokenManager: tokenManger, RabbitMQEmitter: &rabbitMQBroker, RpcClient: &rpcbroker.RpcClient{}, MinioClient: minioClient}
	return &handler, nil
}

func (h *Handler) GetLikeCount(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "postId")
	convertedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
	}
	likeCount, err := h.PostDb.GetPostLikeCountById(context.Background(), convertedPostId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusNotFound)
		return
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
	//ready comment body
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
		http.Error(w, "comment form invalid", http.StatusBadRequest)
		return
	}

	createdComment, err := h.PostDb.CreateComment(context.Background(), post.CreateCommentParams{Body: commentForm.Body, UserID: userId, AuthorName: username})
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusInternalServerError)
		return
	}

	_, err = h.PostDb.CreatePostComment(context.Background(), post.CreatePostCommentParams{
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

func (h *Handler) GetAllPostComments(w http.ResponseWriter, r *http.Request) {
	//get post id from url and convert to int32
	postId := chi.URLParam(r, "postId")
	parsedPostId, err := helpers.ConvertPostId(postId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
	postsComments, err := h.PostDb.GetAllPostCommentsByPostId(context.Background(), parsedPostId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusNotFound)
		return
	}
	// convert to json then send to client
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
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
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
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
	}
	var imageId uuid.NullUUID
	file, header, fileErr := r.FormFile("image")
	if fileErr != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(fileErr)
		imageId.Valid = false
	} else {
		imageId.UUID = uuid.New()
		imageId.Valid = true
	}
	fmt.Printf("Uploaded File: %+v\n", header.Filename)
	fmt.Printf("File Size: %+v\n", header.Size)
	fmt.Printf("MIME Header: %+v\n", header.Header)
	createdPost, err := h.PostDb.CreatePost(context.Background(), post.CreatePostParams{
		Body:       body[0],
		UserID:     convertedUserId,
		AuthorName: username,
		ImageID:    imageId,
	})
	if err != nil {
		helpers.ResponseWithPayload(w, 500, []byte(err.Error()))
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	uploadImage := models.RpcImageUpload{
		UserId:  convertedUserId,
		Image:   buf.Bytes(),
		ImageId: imageId.UUID,
	}
	if fileErr == nil {
		err := h.RpcClient.UploadImage(uploadImage)
		if err != nil {
			log.Println(err)
		}
		err = SendImageToQueue(h, "image-proccessing", imageId.UUID, header.Header.Get("Content-Type"))
		log.Println(time.Now())
		if err != nil {
			log.Println(err)
		}
	}
	err = file.Close()
	if err != nil {
		log.Println(err)
	}

	helpers.ResponseWithPayload(w, http.StatusCreated, []byte(fmt.Sprintf(`{Post created with id: "%v"}`, createdPost.ID)))
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
	respPost, err := h.PostDb.GetPostByIdWithLikes(context.Background(), convertedPostId)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.ResponseNoPayload(w, http.StatusNotFound)
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

func (h *Handler) GetPostsPageByUserId(w http.ResponseWriter, r *http.Request) {
	pageNo := r.URL.Query().Get("pageNo")
	pageSize := r.URL.Query().Get("pageSize")

	userId, err := helpers.GetUserIdFromRequest(r, false)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, http.StatusUnauthorized)
		return
	}
	limit, offset, err := ValidatePagination(pageSize, pageNo)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 400, []byte(err.Error()))
		return
	}
	posts, err := h.PostDb.GetPostByUserIdPaged(context.Background(), post.GetPostByUserIdPagedParams{
		UserID: userId,
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
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
	imageId, err := h.PostDb.DeletePostById(context.Background(), post.DeletePostByIdParams{
		ID:     convertedPostId,
		UserID: parsedUserId,
	})
	if err != nil {
		helpers.ResponseWithJson(w, 500, helpers.JsonResponse{Msg: err.Error()})
		return
	}
	h.RabbitMQEmitter.PushDelete(imageId.UUID.String())
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
	_, err = h.PostDb.CreatePostLike(context.Background(), post.CreatePostLikeParams{
		PostID: convertedPostId,
		UserID: parsedUserId,
	})
	var duplicateEntryError = &pq.Error{Code: "23505"}
	if errors.As(err, &duplicateEntryError) {
		helpers.ResponseNoPayload(w, http.StatusBadRequest)
		return
	}
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
	err = h.PostDb.DeletePostLike(context.Background(), post.DeletePostLikeParams{
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
	isLiked, err := h.PostDb.CheckLike(context.Background(), post.CheckLikeParams{
		PostID: convertedPostId,
		UserID: parsedUserId,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	helpers.ResponseWithPayload(w, 200, []byte(fmt.Sprintf("%v", isLiked)))
}
