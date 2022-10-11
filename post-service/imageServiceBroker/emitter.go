package imageServiceBroker

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return nil
}

func (e *Emitter) Push(event []byte, queue string, key string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	err = channel.PublishWithContext(context.Background(), "image-service", queue, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "multipart", Body: event, Headers: map[string]interface{}{"imageId": key},
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
