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
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return nil
}

func NewRabbitEventEmitter(conn *amqp.Connection) (RabbitMQEmitter, error) {
	emitter := RabbitMQEmitter{
		connection: conn,
	}
	err := emitter.setup()
	if err != nil {
		return RabbitMQEmitter{}, err
	}
	return emitter, nil
}
func (e *RabbitMQEmitter) PushPhotoUpdate(event []byte, queue string, routingKey string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "media-service", routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "application/json", Body: event,
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
func (e *RabbitMQEmitter) PushImage(event []byte, queue string, routingKey string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "media-service", routingKey, false, false, amqp.Publishing{
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
func (e *RabbitMQEmitter) PushDelete(key string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	err = channel.PublishWithContext(context.Background(), "media-service", "delete", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "string", Body: []byte{}, Headers: map[string]interface{}{"imageId": key},
	})
	if err != nil {
		return err
	}
	return nil
}

func ConnectToRabbitMQ(rabbitUrl string) *amqp.Connection {
	backOff := time.Second * 5
	for {
		conn, err := amqp.Dial(rabbitUrl)
		if err != nil {
			log.Println("Connection not ready backing off for:", backOff)
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Connected to rabbit ")
			return conn
		}
	}
}
