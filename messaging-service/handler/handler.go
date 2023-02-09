package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/bernardn38/socialsphere/messaging-service/models"
	"github.com/bernardn38/socialsphere/messaging-service/token"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type Handler struct {
	TokenManager *token.Manager
	Upgrader     websocket.Upgrader
	UserConns    map[int32]*websocket.Conn
	UserMutex    sync.RWMutex
	RedisClient  *redis.Client
}

func NewHandler(config models.Config) (*Handler, error) {
	//init jwt token manager
	tokenManger := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)

	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool { return true }}

	conns := make(map[int32]*websocket.Conn)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "password", // no password set
		DB:       0,          // use default DB
	})

	handler := Handler{TokenManager: tokenManger, Upgrader: upgrader, UserConns: conns, UserMutex: sync.RWMutex{}, RedisClient: redisClient}
	return &handler, nil
}

func (h *Handler) CheckOnline(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdi64, _ := strconv.ParseInt(userId, 10, 32)

	h.UserMutex.RLock()
	_, connected := h.UserConns[int32(userIdi64)]
	h.UserMutex.RUnlock()
	w.Write([]byte(fmt.Sprintf("%v", connected)))
}

func (h *Handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	username, userId, err := GetUsernameAndId(r.Context())
	if err != nil {
		log.Println(err)
		return
	}
	ws, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	AddUserConnection(h, userId, ws)
	defer RemoveUserConnection(h, userId)

	pubsub := h.RedisClient.Subscribe(context.Background(), string(userId))
	defer pubsub.Close()

	for {

		//parse message
		msg := models.Message{FromUserId: userId, FromUsername: username}
		err = ws.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("%+v", msg)

		//check if target user is online then send message
		if conn, ok := h.UserConns[msg.ToUserId]; ok {
			err = conn.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				break
			}
		}

		//send notification to redis
		err = PublishToRedis(h, msg)
		if err != nil {
			log.Println(err)
		}
	}

}
