package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/image-service/helpers"
	"github.com/bernardn38/socialsphere/image-service/sql/userImages"
	"github.com/bernardn38/socialsphere/image-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
)

type Handler struct {
	TokenManager *token.Manager
	UserImageDB  *userImages.Queries
	MinioClient  *minio.Client
}

// currently unused only support uploading jpeg
func (handler *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
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
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	err = helpers.UploadToS3(handler.MinioClient, buf.Bytes(), uuid.New().String())
	if err != nil {
		http.Error(w, "Error uploading to", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", header.Filename)
	fmt.Printf("File Size: %+v\n", header.Size)
	fmt.Printf("MIME Header: %+v\n", header.Header)

	helpers.ResponseNoPayload(w, 201)
}

func (handler *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	// get image from s3 bucket
	imageId := chi.URLParam(r, "imageId")
	object, err := helpers.GetImageFromS3(handler.MinioClient, imageId)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//send image to client; cache image in client
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, object)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
