package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"

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

func SendImageToQueue(file multipart.File, h *Handler, imageId uuid.UUID, contentType string) error {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Println(err)
	}
	err := h.Emitter.Push(buf.Bytes(), "image-proccessing", imageId.String(), contentType)
	if err != nil {
		return err
	}
	return nil
}
