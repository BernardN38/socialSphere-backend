package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bernardn38/socialsphere/messaging-service/token"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"gopkg.in/go-playground/validator.v9"
)

type Handler struct {
	TokenManager *token.Manager
	Upgrader     websocket.Upgrader
	Conns        map[int32]*websocket.Conn
	UserMutex    sync.RWMutex
	Rdb          *redis.Client
}

var ctx = context.Background()

func (h *Handler) CheckOnline(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdi64, _ := strconv.ParseInt(userId, 10, 32)
	log.Println(h.Conns, userId)

	h.UserMutex.RLock()
	_, connected := h.Conns[int32(userIdi64)]
	h.UserMutex.RUnlock()
	w.Write([]byte(fmt.Sprintf("%v", connected)))
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	username, userId, err := getUsernameAndPassword(r.Context())
	if err != nil {
		log.Println(err)
		return
	}
	ws, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	h.UserMutex.RLock()
	h.Conns[userId] = ws
	h.UserMutex.RUnlock()

	defer func() {
		h.UserMutex.RLock()
		delete(h.Conns, userId)
		h.UserMutex.RUnlock()
	}()

	pubsub := h.Rdb.Subscribe(ctx, string(userId))
	defer pubsub.Close()
	for {
		//parse message
		msg := Message{FromUserId: userId, FromUsername: username, Timestamp: time.Now()}
		err = ws.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("%+v", msg)

		//check if target user is online then send message
		if conn, ok := h.Conns[msg.ToUserId]; ok {
			err = conn.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				break
			}
		}

		//send notification to redis
		err = publishToRedis(h, msg)
		if err != nil {
			log.Println(err)
		}
	}

}

func publishToRedis(h *Handler, msg Message) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	notification := Notification{
		UserId:       msg.ToUserId,
		FromUsername: msg.FromUsername,
		Payload:      string(jsonMsg),
		MessageType:  "newMessage",
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
	err = h.Rdb.Publish(context.Background(), "notifications", notificationJson).Err()
	if err != nil {
		return err
	}
	return nil
}

type Notification struct {
	UserId       int32  `json:"userId" validate:"required"`
	FromUsername string `json:"fromUsername"`
	Payload      string `json:"payload" validate:"required"`
	MessageType  string `json:"type" validate:"required"`
}

func (c *Notification) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

type Message struct {
	FromUserId   int32     `json:"fromUserId"`
	FromUsername string    `json:"fromUsername"`
	ToUserId     int32     `json:"toUserId"`
	Subject      string    `json:"subject"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

func getUsernameAndPassword(ctx context.Context) (string, int32, error) {
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
