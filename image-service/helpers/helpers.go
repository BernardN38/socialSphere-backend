package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"time"
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

func UploadToS3(file []byte, key string) error {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession())
	region := "us-east-1"
	sess.Config.Region = &region
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("image-service-socialsphere"),
		Key:    aws.String(key),
		Body:   bytes.NewBuffer(file),
	})
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
	return nil
}

func GetImageFromS3(key string) (*os.File, error) {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession())
	region := "us-east-1"
	sess.Config.Region = &region

	downloader := s3manager.NewDownloader(sess)
	fileId := uuid.New()
	file, err := os.Create(fileId.String())
	if err != nil {
		return nil, err
	}
	download, err := downloader.Download(file, &s3.GetObjectInput{Bucket: aws.String("image-service-socialsphere"), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	log.Println("Download: ", download)
	return file, err
}
