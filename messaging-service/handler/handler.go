package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bernardn38/socialsphere/messaging-service/models"
	"github.com/bernardn38/socialsphere/messaging-service/token"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	TokenManager *token.Manager
	Upgrader     websocket.Upgrader
	UserConns    map[int32]*websocket.Conn
	UserMutex    sync.RWMutex
	RedisClient  *redis.Client
	MongoClient  *mongo.Client
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
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		panic(err)
	}

	handler := Handler{TokenManager: tokenManger, Upgrader: upgrader, UserConns: conns, UserMutex: sync.RWMutex{}, RedisClient: redisClient, MongoClient: mongoClient}
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
func (h *Handler) GetAllMessages(w http.ResponseWriter, r *http.Request) {
	collection := h.MongoClient.Database("message-service").Collection("message")
	cursor, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusNotFound)
		return
	}
	messages := []models.Message{}
	err = cursor.All(context.Background(), &messages)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
func (h *Handler) GetAllUserMessages(w http.ResponseWriter, r *http.Request) {
	_, userId, err := GetUsernameAndId(r.Context())
	if err != nil {
		log.Println(err)
		return
	}
	messages, err := GetUserMessagePages(h.MongoClient, r.Context(), 1, 10, userId)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
func (h *Handler) GetUserMessages(w http.ResponseWriter, r *http.Request) {
	pageNo := r.URL.Query().Get("pageNo")
	convPageNo, err := strconv.ParseInt(pageNo, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid page number", http.StatusBadRequest)
		return
	}
	pageSize := r.URL.Query().Get("pageSize")
	convPageSize, err := strconv.ParseInt(pageSize, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid page size", http.StatusBadRequest)
	}
	_, userId, err := GetUsernameAndId(r.Context())
	if err != nil {
		log.Println(err)
		return
	}
	targetUserId := chi.URLParam(r, "userId")
	convTargetId, err := strconv.ParseInt(targetUserId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	page, err := GetMessagePage(h.MongoClient, r.Context(), convPageNo, convPageSize, userId, int32(convTargetId))
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(page)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
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

	collection := h.MongoClient.Database("message-service").Collection("message")
	for {

		//parse message
		msg := models.Message{FromUserId: userId, FromUsername: username, CreatedAt: time.Now()}
		err = ws.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			break
		}
		result, err := collection.InsertOne(context.Background(), msg)
		if err != nil {
			log.Println(err)
		}
		log.Printf("%+v, %+v", msg, result)

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
