package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"strconv"

	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
)

func ValidatePostForm(reqBody []byte) (*Post, error) {
	var form Post
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

func SendImageToQueue(file multipart.File, handler *Handler, routingKey string, imageId uuid.UUID, contentType string) error {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Println(err)
	}
	// err := handler.Emitter.Push(buf.Bytes(), "image-service", imageId.String())
	err := handler.Emitter.Push(buf.Bytes(), "image-service", routingKey, imageId.String(), contentType)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SendDeleteToQueue(routingKey string, imageId uuid.UUID, handler *Handler) error {
	err := handler.Emitter.Push(nil, "image-service", routingKey, imageId.String(), "")
	if err != nil {
		return err
	}
	return nil
}
