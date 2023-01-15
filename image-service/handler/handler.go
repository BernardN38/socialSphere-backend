package handler

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bernardn38/socialsphere/image-service/helpers"
	"github.com/bernardn38/socialsphere/image-service/sql/userImages"
	"github.com/bernardn38/socialsphere/image-service/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	TokenManager *token.Manager
	UserImageDB  *userImages.Queries
	AwsSession   *session.Session
}

func (handler *Handler) GetUserImages(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")
	parsedId, err := uuid.Parse(userId.(string))
	if err != nil {
		helpers.ResponseWithJson(w, 400, helpers.JsonResponse{Msg: "invalid user id"})
		return
	}
	ids, err := handler.UserImageDB.GetImagesByUserId(context.Background(), parsedId)
	if err != nil {
		helpers.ResponseWithJson(w, 500, helpers.JsonResponse{Msg: "error retrieving imageIds"})
		return
	}
	helpers.ResponseWithJson(w, 200, helpers.JsonResponse{Data: ids})
}

func (handler *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	start := time.Now()
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 413, []byte("image too large"))
	}

	// Get header for filename, size and headers
	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	img, format, err := image.Decode(file)
	log.Println(format)
	opts := jpeg.Options{Quality: 60}

	buf := bytes.NewBuffer(nil)

	jpeg.Encode(buf, img, &opts)

	err = helpers.UploadToS3(buf.Bytes(), uuid.New().String())
	if err != nil {
		helpers.ResponseNoPayload(w, 500)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", header.Filename)
	fmt.Printf("File Size: %+v\n", header.Size)
	fmt.Printf("MIME Header: %+v\n", header.Header)

	fmt.Fprintf(w, "Successfully Uploaded File\n")
	end := time.Now()
	log.Println(end.UnixMilli()-start.UnixMilli(), " ms")
}

func (handler *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageId := chi.URLParam(r, "imageId")
	file, err := helpers.GetImageFromS3(imageId)
	defer file.Close()
	if err != nil {
		helpers.ResponseNoPayload(w, 404)
		return
	}
	w.Header().Set("Cache-Control", "max-age=2592000") //
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, file)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
}
