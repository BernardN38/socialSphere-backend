package rabbitmq_broker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/bernardn38/socialsphere/identity-service/models"
	"github.com/bernardn38/socialsphere/identity-service/sql/users"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitBroker struct {
	connection *amqp.Connection
}

func RunRabbitBroker(config models.Config, userDb *users.Queries) {
	rabbitMQConn := ConnectToRabbitMQ(config.RabbitmqUrl)
	for i := 0; i < 10; i++ {
		go ListenForMessages(&config, rabbitMQConn, userDb)
	}
}
func (e *RabbitBroker) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return nil
}

func (e *RabbitBroker) PushImage(event []byte, queue string, routingKey string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "image-service", routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "multipart", Body: event,
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
func (e *RabbitBroker) PushImageDelete(key string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "image-service", "delete", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "string", Body: []byte{}, Headers: map[string]interface{}{"imageId": key},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewRabbitBroker(conn *amqp.Connection) (*RabbitBroker, error) {
	rabbitBroker := RabbitBroker{
		connection: conn,
	}
	err := rabbitBroker.setup()
	if err != nil {
		return nil, err
	}
	return &rabbitBroker, nil
}

func ListenForMessages(config *models.Config, conn *amqp.Connection, usersDb *users.Queries) {

	channel, err := conn.Channel()
	if err != nil {
		return
	}
	err = channel.Qos(1, 0, false)
	if err != nil {
		return
	}
	messages, err := channel.Consume("identity-service", "", false, false, false, false, nil)
	if err != nil {
		return
	}
	var forever chan struct{}

	for d := range messages {
		switch messageType := d.RoutingKey; messageType {
		case "createUser":
			var user models.AuthServiceCreateUserParams
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
			err = usersDb.CreateUser(context.Background(), users.CreateUserParams{
				ID:        user.UserId,
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
