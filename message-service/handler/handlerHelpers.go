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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	log.Println("removing connection", userId)
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

func GetMessagePage(client *mongo.Client, ctx context.Context, page int64, pageSize int64, userId int32, targetId int32) ([]models.Message, error) {
	collection := client.Database("message-service").Collection("message")
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created_at": -1})
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip((page - 1) * pageSize)
	cursor, err := collection.Find(ctx, bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"to_user_id": userId},
					bson.M{"from_user_id": userId},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"to_user_id": targetId},
					bson.M{"from_user_id": targetId},
				},
			},
		},
	}, findOptions)
	if err != nil {
		return nil, err
	}
	messages := make([]models.Message, page)
	err = cursor.All(ctx, &messages)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return messages, nil
}

func GetUserMessagePages(client *mongo.Client, ctx context.Context, page int64, pageSize int64, userId int32) (map[int32][]models.Message, error) {
	collection := client.Database("message-service").Collection("message")
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created_at": -1})
	findOptions.SetLimit(pageSize)
	findOptions.SetSkip((page - 1) * pageSize)
	filter := bson.M{
		"$or": []bson.M{
			bson.M{"from_user_id": userId, "to_user_id": bson.M{"$ne": userId}},
			bson.M{"from_user_id": bson.M{"$ne": userId}, "to_user_id": userId},
		},
	}
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	// Create a map to store the messages
	messagesMap := make(map[int32][]models.Message)
	for cursor.Next(ctx) {
		var message models.Message
		err := cursor.Decode(&message)
		if err != nil {
			// Handle error
			log.Println(err)
		}

		// Determine the other user involved in the conversation
		var otherUserID int32
		if message.FromUserId == userId {
			otherUserID = message.ToUserId
		} else {
			otherUserID = message.FromUserId
		}

		// Add the message to the map for the other user
		messagesMap[otherUserID] = append(messagesMap[otherUserID], message)
	}

	return messagesMap, nil
}
