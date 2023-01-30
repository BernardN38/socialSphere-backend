package handler

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

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
	err = h.Emitter.PushImage(jsonMessage, "image-service", routingKey)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
