package rabbitmq_broker

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQEmitter struct {
	connection *amqp.Connection
}

func (e *RabbitMQEmitter) setup() error {
	return nil
}

func (e *RabbitMQEmitter) Push(event []byte, routingKey string, contentType string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "authentication-service", routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: contentType, Body: event,
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
func (e *RabbitMQEmitter) PushDelete(key string) error {
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
func NewEventEmitter(conn *amqp.Connection) *RabbitMQEmitter {
	emitter := RabbitMQEmitter{
		connection: conn,
	}
	return &emitter
}

func ConnectToRabbitMQ(rabbitUrl string) *amqp.Connection {
	backOff := time.Second * 5
	for {
		conn, err := amqp.Dial(rabbitUrl)
		if err != nil {
			log.Println("Connection not ready backing off for ", backOff)
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Connected to rabbit ")
			return conn
		}
	}
}
