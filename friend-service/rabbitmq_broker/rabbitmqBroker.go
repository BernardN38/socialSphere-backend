package rabbitmq_broker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/bernardn38/socialsphere/friend-service/models"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	e.channel = channel
	defer channel.Close()
	return nil
}
func RunRabbitBroker(config models.Config, userDb *users.Queries) {
	rabbitMQConn := ConnectToRabbitMQ(config.RabbitmqUrl)
	for i := 0; i < 10; i++ {
		go ListenForMessages(config, rabbitMQConn, userDb)
	}
}
func (e *Emitter) Push(event []byte, queue string, routingKey string, imageId string, contentType string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "image-service", routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "multipart", Body: event, Headers: map[string]interface{}{"imageId": imageId, "contentType": contentType},
	})
	if err != nil {
		return err
	}
	err = channel.Close()
	if err != nil {
		log.Println(err)
	}
	return nil
}
func (e *Emitter) PushDelete(key string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	err = channel.PublishWithContext(context.Background(), "image-service", "delete", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "string", Body: []byte{}, Headers: map[string]interface{}{"imageId": key},
	})
	if err != nil {
		return err
	}
	return nil
}
func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	emitter := Emitter{
		connection: conn,
	}
	err := emitter.setup()
	if err != nil {
		return Emitter{}, err
	}
	return emitter, nil
}

func ListenForMessages(config models.Config, conn *amqp.Connection, userDb *users.Queries) {
	channel, err := conn.Channel()
	if err != nil {
		return
	}
	err = channel.Qos(1, 0, false)
	if err != nil {
		return
	}
	messages, err := channel.Consume("friend-service", "", false, false, false, false, nil)
	if err != nil {
		return
	}
	var forever chan struct{}

	for d := range messages {
		switch messageType := d.RoutingKey; messageType {
		case "createUser":
			var user models.UserForm
			err := json.Unmarshal(d.Body, &user)
			if err != nil {
				log.Println(err)
				d.Reject(false)
				return
			}
			err = user.Validate()
			if err != nil {
				log.Println("user form not valid")
				d.Reject(false)
				return
			}
			_, err = userDb.CreateUser(context.Background(), users.CreateUserParams{
				UserID:    user.UserId,
				Username:  user.Username,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			})
			if err != nil {
				log.Println(err)
				d.Reject(false)
				return
			}
		case "userPhotoUpload":
			var userUploadPhotoForm models.UserUploadPhotoForm
			err := json.Unmarshal(d.Body, &userUploadPhotoForm)
			if err != nil {
				log.Println(err)
				d.Reject(false)
				return
			}
			err = userUploadPhotoForm.Validate()
			if err != nil {
				log.Println("user form not valid")
				d.Reject(false)
				return
			}
			err = userDb.UpdateUserLastUpload(context.Background(), users.UpdateUserLastUploadParams{
				LastUpload:  sql.NullTime{Time: time.Now(), Valid: true},
				LastImageID: uuid.NullUUID{UUID: userUploadPhotoForm.ImageId, Valid: true},
				UserID:      userUploadPhotoForm.UserId,
			})
			if err != nil {
				log.Println(err)
				d.Reject(false)
				return
			}
		default:
			log.Println("no case" + messageType)
		}
		d.Ack(true)
	}
	<-forever
}

func ConnectToRabbitMQ(rabbitUrl string) *amqp.Connection {
	backOff := time.Second * 5
	for {
		conn, err := amqp.Dial(rabbitUrl)
		if err != nil {
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Successfully connected to rabbitmq")
			return conn
		}
	}
}
