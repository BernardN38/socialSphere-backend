package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/identity-service/helpers"
	"github.com/bernardn38/socialsphere/identity-service/models"
	"github.com/bernardn38/socialsphere/identity-service/rabbitmq_broker"
	imageServiceBroker "github.com/bernardn38/socialsphere/identity-service/rabbitmq_broker"
	rpcbroker "github.com/bernardn38/socialsphere/identity-service/rpc_broker"
	"github.com/bernardn38/socialsphere/identity-service/sql/users"
	"github.com/bernardn38/socialsphere/identity-service/token"
	"github.com/google/uuid"
)

type Handler struct {
	UserDb       *users.Queries
	TokenManager *token.Manager
	Emitter      *imageServiceBroker.RabbitBroker
	RpcClient    *rpcbroker.RpcClient
}

func NewHandler(config models.Config) (*Handler, error) {
	//open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		return nil, err
	}

	// init sqlc user queries
	queries := users.New(db)

	//init jwt token manager
	tokenManger := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)

	//init rabbitmq message emitter
	rabbitMQConn := rabbitmq_broker.ConnectToRabbitMQ(config.RabbitmqUrl)
	rabbitBroker, err := imageServiceBroker.NewRabbitBroker(rabbitMQConn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	handler := Handler{UserDb: queries, TokenManager: tokenManger, Emitter: rabbitBroker, RpcClient: &rpcbroker.RpcClient{}}
	return &handler, nil
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// parse user form from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	userForm := models.UserForm{}
	json.Unmarshal(body, &userForm)

	//validate user form
	err = userForm.Validate()
	if err != nil {
		log.Println(err)
		http.Error(w, "user form invalid", http.StatusBadRequest)
		return
	}

	//create new user in database
	err = h.UserDb.CreateUser(context.Background(), users.CreateUserParams{ID: userForm.UserId,
		Username: userForm.Username, FirstName: userForm.Username, LastName: userForm.LastName, Email: userForm.Email})
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	helpers.ResponseNoPayload(w, 201)
}

func (h *Handler) CreateUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	//parse image from form
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// create profile image association in database
	imageId := uuid.New()
	if err = h.UserDb.CreateUserProfileImage(context.Background(), users.CreateUserProfileImageParams{
		UserID:  userId,
		ImageID: imageId,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	//send image to rabbitmq for processing and upload to s3 bucket
	imageUpload := models.RpcImageUpload{
		UserId:  userId,
		Image:   buf.Bytes(),
		ImageId: imageId,
	}
	err = h.RpcClient.UploadImage(imageUpload)
	if err != nil {
		log.Println("rpc error", err)
		return
	}
	if file != nil {
		err = SendImageToQueue(h, "image-proccessing", imageId, header.Header.Get("Content-Type"))
		if err != nil {
			log.Println(err)
			return
		}
	}

	helpers.ResponseNoPayload(w, http.StatusCreated)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing return error
	userId, err := helpers.GetUserIdFromRequest(r, false)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	//get from database
	user, err := h.UserDb.GetUserById(context.Background(), userId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	//respond with json payload of user data
	jsonResponse, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, 200, []byte(jsonResponse))
}

func (h *Handler) GetUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	//get image id from datbase for specified user id
	imageId, err := h.UserDb.GetUserProfileImage(context.Background(), userId)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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

func (h *Handler) GetOwnProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("could not get userId from context")
		http.Error(w, "", http.StatusInternalServerError)
	}
	convertedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
	}
	//get image id from datbase for specified user id
	imageId, err := h.UserDb.GetUserProfileImage(context.Background(), convertedUserId)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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
