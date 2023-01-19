package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"

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

// currently unused only support uploading jpeg
func (handler *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		helpers.ResponseWithPayload(w, 413, []byte("image too large"))
		return
	}

	// Get header for filename, size and headers
	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	// compress image
	img, _, err := image.Decode(file)
	opts := jpeg.Options{Quality: 60}
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
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
}

func (handler *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	// get image from s3 bucket
	imageId := chi.URLParam(r, "imageId")
	file, err := helpers.GetImageFromS3(imageId)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
	defer file.Close()
	if err != nil {
		helpers.ResponseNoPayload(w, 404)
		return
	}

	//send image to client; cache image in client
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, file)
	if err != nil {
		log.Println(err)
		helpers.ResponseNoPayload(w, 500)
		return
	}
}
