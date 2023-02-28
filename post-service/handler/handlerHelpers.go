package handler

import (
	"errors"
	"log"
	"mime/multipart"
	"net/http"
)

func GetBodyAndImage(r *http.Request) (string, multipart.File, *multipart.FileHeader, error) {
	var body string
	var file multipart.File

	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		return "", nil, nil, err
	}

	bodyArr, ok := r.MultipartForm.Value["body"]
	if ok {
		if len(bodyArr) > 0 {
			body = bodyArr[0]
		}
	}
	file, header, fileErr := r.FormFile("image")
	if fileErr != nil {
		log.Println(err)
	}
	if body == "" && file == nil {
		return "", nil, nil, errors.New("request form empty")
	}
	return body, file, header, nil
}
