package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/bernardn38/socialsphere/messaging-service/token"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type Handler struct {
	TokenManager *token.Manager
	Upgrader     websocket.Upgrader
	Conns        map[string]*websocket.Conn
	UserMutex    sync.RWMutex
	Rdb          *redis.Client
}

var ctx = context.Background()

func (h *Handler) CheckOnline(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	h.UserMutex.RLock()
	_, connected := h.Conns[userId]
	h.UserMutex.RUnlock()
	w.Write([]byte(fmt.Sprintf("%v", connected)))
}
func (handler *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value("userId").(string)
	username := r.Context().Value("username").(string)
	log.Println(userId, "Connected")

	ws, err := handler.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	handler.Conns[userId] = ws
	pubsub := handler.Rdb.Subscribe(ctx, userId)
	defer pubsub.Close()
	for {
		var req Request
		err := ws.ReadJSON(&req)
		if err != nil {
			log.Println(err)
			return
		}
		req.Timestamp = time.Now()
		req.From = userId
		req.FromUserName = username
		message, _ := json.Marshal(req)
		err = handler.Rdb.Publish(ctx, userId, message).Err()
		if err != nil {
			log.Println(err)
		}
		// if userId == req.To {
		// 	continue
		// }
		if conn, ok := handler.Conns[req.To]; ok {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				panic(err)
			}

			fmt.Println(msg.Channel, msg.Payload)
			err = conn.WriteJSON(msg)
			if err != nil {
				log.Panicln(err)
			}
		}
	}
}

type Request struct {
	From         string    `json:"from,omitempty"`
	FromUserName string    `json:"fromUserName,omitempty"`
	To           string    `json:"to,omitempty"`
	Message      string    `json:"message,omitempty"`
	Subject      string    `json:"subject,omitempty"`
	Timestamp    time.Time `json:"time,omitempty"`
}
