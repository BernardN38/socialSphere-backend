package models

import (
	"time"

	"github.com/google/uuid"
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

type RpcImageUpload struct {
	UserId  int32
	Image   []byte
	ImageId uuid.UUID
}
