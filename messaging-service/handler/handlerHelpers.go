package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/bernardn38/socialsphere/messaging-service/models"
	"github.com/bernardn38/socialsphere/messaging-service/token"
	"github.com/gorilla/websocket"
)

func PublishToRedis(h *Handler, msg models.Message) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	notification := models.Notification{
		UserId:      msg.ToUserId,
		Payload:     string(jsonMsg),
		MessageType: "newMessage",
	}
	err = notification.Validate()
	if err != nil {
		return err
	}
	notificationJson, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	log.Println(notification)
	err = h.RedisClient.Publish(context.Background(), "notifications", notificationJson).Err()
	if err != nil {
		return err
	}
	return nil
}

func AddUserConnection(h *Handler, userId int32, ws *websocket.Conn) {
	h.UserMutex.RLock()
	h.UserConns[userId] = ws
	h.UserMutex.RUnlock()
}

func RemoveUserConnection(h *Handler, userId int32) {
	h.UserMutex.RLock()
	delete(h.UserConns, userId)
	h.UserMutex.RUnlock()
}

func GetUsernameAndId(ctx context.Context) (string, int32, error) {
	userId := ctx.Value(token.KeyUserid).(string)
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		return "", 0, err
	}
	username, ok := ctx.Value(token.KeyUsername).(string)
	if !ok {
		return "", 0, errors.New("username not valid")
	}
	return username, int32(userIdi64), nil
}
