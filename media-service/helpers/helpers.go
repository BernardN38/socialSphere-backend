package helpers

import (
	"encoding/json"
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

func DeleteFromS3(m *minio.Client, imageId string) error {
	err := m.RemoveObject("media-service-socialsphere1", imageId)
	if err != nil {
		return err
	}
	return nil
}
