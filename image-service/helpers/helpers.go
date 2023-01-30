package helpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/image-service/models"
	"github.com/minio/minio-go"
)

func ResponseWithJson(w http.ResponseWriter, statusCode int, payload models.JsonResponse) {
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

func DeleteFromS3(m *minio.Client, imageId string) error {
	err := m.RemoveObject("image-service-socialsphere1", imageId)
	if err != nil {
		return err
	}
	return nil
}

func GetImageFromS3(m *minio.Client, imageId string) (*minio.Object, error) {
	// Get the object
	object, err := m.GetObject("image-service-socialsphere1", imageId, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}
