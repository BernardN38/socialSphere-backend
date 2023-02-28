package handler

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/image-service/helpers"
	"github.com/bernardn38/socialsphere/image-service/models"
	"github.com/bernardn38/socialsphere/image-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/image-service/sql/userImages"
	"github.com/bernardn38/socialsphere/image-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
)

type Handler struct {
	TokenManager    *token.Manager
	UserImageDB     *userImages.Queries
	RabbitMQEmitter *rabbitmq_broker.RabbitMQEmitter
	MinioClient     *minio.Client
}

func NewHandler(config models.Config) (*Handler, error) {
	//open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		return nil, err
	}
	// init sqlc user queries
	queries := userImages.New(db)

	//init jwt token manager
	tokenManger := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)
	minioClient, err := minio.New("minio:9000", config.MinioKey, config.MinioSecret, false)
	if err != nil {
		return nil, err
	}
	//init rabbitmq message emitter
	rabbitConn := rabbitmq_broker.ConnectToRabbitMQ(config.RabbitmqUrl)
	rabbitMQBroker, err := rabbitmq_broker.NewRabbitEventEmitter(rabbitConn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	handler := Handler{UserImageDB: queries, TokenManager: tokenManger, RabbitMQEmitter: &rabbitMQBroker, MinioClient: minioClient}
	return &handler, nil
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		http.Error(w, "File too Large", http.StatusRequestEntityTooLarge)
		return
	}

	// Get header for filename, size and headers
	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "image missing in body", http.StatusBadRequest)
		return
	}
	defer file.Close()

	_, err = h.MinioClient.PutObject("image-service-socialsphere1", uuid.NewString(), file, header.Size, minio.PutObjectOptions{})
	if err != nil {
		log.Println(err)
		http.Error(w, "error uploading to minio", http.StatusInternalServerError)
		return
	}

	helpers.ResponseNoPayload(w, 201)
}

func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	// get image from s3 bucket
	imageId := chi.URLParam(r, "imageId")
	object, err := h.MinioClient.GetObject("media-service-socialsphere1", imageId, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
		http.Error(w, "image id not found", http.StatusNotFound)
		return
	}
	stat, err := object.Stat()
	if err != nil {
		log.Println(err)
	}
	//send image to client; cache image in client
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Header().Set("Content-Type", stat.ContentType)
	_, err = io.Copy(w, object)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
