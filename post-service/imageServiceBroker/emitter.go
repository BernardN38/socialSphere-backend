package imageServiceBroker

import (
	"context"
	"log"

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
