package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/bernardn38/socialsphere/post-service/models"
	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
)

func ValidatePostForm(reqBody []byte) (*models.Post, error) {
	var form models.Post
	err := json.Unmarshal(reqBody, &form)
	if err != nil {
		return nil, err
	}

	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return nil, err
	}

	return &form, nil
}

type PaginationForm struct {
	PageSize int64 `json:"pageSize" validate:"required,min=1,max=20"`
	PageNo   int64 `json:"pageNo" validate:"required,min=1"`
}

func ValidatePagination(pageSize string, pageNo string) (int32, int32, error) {

	parsedPageSize, err := strconv.ParseInt(pageSize, 10, 64)
	if err != nil {
		parsedPageSize = 10
		return 0, 0, errors.New("invalid pageSize")
	}
	parsedPageNo, err := strconv.ParseInt(pageNo, 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid pageNo")
	}
	form := PaginationForm{
		PageSize: parsedPageSize,
		PageNo:   parsedPageNo,
	}
	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return 0, 0, err
	}
	offset := (parsedPageNo - 1) * parsedPageSize
	return int32(parsedPageSize), int32(offset), nil
}

func SendImageToQueue(h *Handler, routingKey string, imageId uuid.UUID, contentType string) error {
	message := map[string]string{
		"imageId":     imageId.String(),
		"contentType": contentType,
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return err
	}
	err = h.RabbitMQEmitter.PushImage(jsonMessage, "image-service", routingKey)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type UserUploadPhotoForm struct {
	UserId  int32     `json:"userId"`
	ImageId uuid.UUID `json:"imageId"`
}

func SendUserPhotoUploadUpdate(h *Handler, routingKey string, imageId uuid.UUID, userId int32) error {
	message := UserUploadPhotoForm{
		UserId:  userId,
		ImageId: imageId,
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return err
	}
	err = h.RabbitMQEmitter.PushPhotoUpdate(jsonMessage, "friend-service", "userPhotoUpload")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
func SendDeleteToQueue(routingKey string, imageId uuid.UUID, h *Handler) error {
	err := h.RabbitMQEmitter.PushImage(nil, "image-service", routingKey)
	if err != nil {
		return err
	}
	return nil
}

func UploadImage(h *Handler, userId int32, imageId uuid.NullUUID, contentType string, file multipart.File) error {
	if file == nil {
		return errors.New("file is nil")
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return err
	}

	uploadImage := models.RpcImageUpload{
		UserId:  userId,
		Image:   buf.Bytes(),
		ImageId: imageId.UUID,
	}
	err := h.RpcClient.UploadImage(uploadImage)
	if err != nil {
		return err
	}
	err = SendImageToQueue(h, "image-proccessing", imageId.UUID, contentType)
	if err != nil {
		return err
	}

	return nil
}

func GetBodyAndImage(r *http.Request) (string, multipart.File, string, error) {
	var body string
	var file multipart.File

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return "", nil, "", err
	}

	bodyArr, ok := r.MultipartForm.Value["body"]
	if ok {
		if len(body) > 0 {
			body = bodyArr[0]
		}
	}
	file, header, fileErr := r.FormFile("image")
	if fileErr != nil {
		return body, nil, "", err
	}

	defer file.Close()
	if body == "" && file == nil {
		return "", nil, "", errors.New("request form empty")
	}
	return body, file, header.Header.Get("Content-Type"), nil
}
