package rabbitmq_broker

import (
	"context"
	"log"
	"time"

	"github.com/bernardn38/socialsphere/image-service/helpers"
	"github.com/bernardn38/socialsphere/image-service/models"
	"github.com/minio/minio-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQEmitter struct {
	connection *amqp.Connection
}

func RunRabbitBroker(config models.Config) {
	rabbitMQConn := ConnectToRabbitMQ(config.RabbitmqUrl)
	minioClient, err := minio.New("minio:9000", config.MinioKey, config.MinioSecret, false)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		go ListenForMessages(&config, minioClient, rabbitMQConn)
	}
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
func (e *RabbitMQEmitter) PushImage(event []byte, queue string, routingKey string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	err = channel.PublishWithContext(context.Background(), "media-service", routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "multipart", Body: event,
	})
	if err != nil {
		return err
	}

	return nil
}
func ListenForMessages(config *models.Config, m *minio.Client, conn *amqp.Connection) {

	channel, err := conn.Channel()
	if err != nil {
		return
	}
	err = channel.Qos(1, 0, false)
	if err != nil {
		return
	}
	messages, err := channel.Consume("media-service", "", false, false, false, false, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var forever chan struct{}

	for d := range messages {
		switch messageType := d.RoutingKey; messageType {
		case "delete":
			imageId, ok := d.Headers["imageId"].(string)
			if !ok {
				log.Println("image id invalid")
				return
			}
			err := helpers.DeleteFromS3(m, imageId)
			if err != nil {
				log.Println(err)
				d.Ack(false)
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
			log.Println("Connection not ready backing off for ", backOff)
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Connected to rabbit ")
			return conn
		}
	}
}
