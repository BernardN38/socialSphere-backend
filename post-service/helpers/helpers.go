package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go"
)

type JsonResponse struct {
	Msg       string      `json:"msg,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
}

type PageResponse struct {
	Page     interface{} `json:"page"`
	PageSize int         `json:"pageSize"`
	PageNo   int32       `json:"pageNo"`
	LastPage bool        `json:"lastPage"`
}

func ResponseWithJson(w http.ResponseWriter, statusCode int, payload JsonResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = w.Write(jsonData)
}
func ResponseWithPayload(w http.ResponseWriter, responseCode int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	_, _ = w.Write(payload)
}
func ResponseNoPayload(w http.ResponseWriter, responseCode int) {
	w.WriteHeader(responseCode)
}

func ConvertPostId(postId string) (int32, error) {
	parsedPostId, err := strconv.ParseInt(postId, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsedPostId), nil
}

func ConvertUserId(userId any) (int32, error) {
	stringUserId, ok := userId.(string)
	if !ok {
		log.Println(stringUserId)
		return 0, errors.New("invalid userId")
	}
	userId64, err := strconv.ParseInt(stringUserId, 10, 32)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return int32(userId64), nil
}

func GetUserIdFromRequest(r *http.Request, checkContext bool) (int32, error) {
	var userId int32

	urlUserId := chi.URLParam(r, "userId")
	if len(urlUserId) > 0 {
		convertedId, err := ConvertUserId(urlUserId)
		if err != nil {
			log.Println(err)
			return 0, errors.New("invalid user id url")
		}
		userId = convertedId
	} else if checkContext {
		contextUserId := r.Context().Value("userId")
		convertedUserId, err := ConvertUserId(contextUserId)
		if err != nil {
			log.Println(err)
			return 0, errors.New("invalid user id context")
		}
		userId = convertedUserId
	}
	return userId, nil

}

func UploadToS3(m *minio.Client, file []byte, imageId string) error {
	fileReader := bytes.NewReader(file)
	info, err := m.PutObject("image-service-socialsphere1", imageId, fileReader, fileReader.Size(), minio.PutObjectOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(info)
	return nil
}
