package application

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/image-service/handler"
	"github.com/bernardn38/socialsphere/image-service/helpers"
	"github.com/bernardn38/socialsphere/image-service/sql/userImages"
	"github.com/bernardn38/socialsphere/image-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	jwtSecretKey     string
	jwtSigningMethod jwt.Algorithm
	dsn              string
	rabbitmqUrl      string
	minioKey         string
	minioSecret      string
}
type App struct {
	srv          server
	pgDb         *sql.DB
	tokenManager *token.Manager
}

type server struct {
	router  *chi.Mux
	handler *handler.Handler
}

func New() *App {
	app := App{}
	dsn := os.Getenv("DSN")
	config := Config{jwtSecretKey: "superSecretKey", jwtSigningMethod: jwt.HS256, dsn: dsn, rabbitmqUrl: "amqp://guest:guest@rabbitmq",
		minioKey: "minio", minioSecret: "minio123"}
	app.runAppSetup(config)
	return &app
}
func (app *App) Run() {
	log.Printf("listening on port %s", "8080")
	log.Fatal(http.ListenAndServe(":8080", app.srv.router))
}

func (app *App) runAppSetup(config Config) {
	db, err := sql.Open("postgres", config.dsn)
	if err != nil {
		log.Fatal(err)
	}
	minioClient, err := minio.New("minio:9000", config.minioKey, config.minioSecret, false)
	if err != nil {
		log.Fatal(err)
	}
	queries := userImages.New(db)
	tokenManger := token.NewManager([]byte(config.jwtSecretKey), config.jwtSigningMethod)
	h := &handler.Handler{TokenManager: tokenManger, UserImageDB: queries, MinioClient: minioClient}
	for i := 0; i < 10; i++ {
		go ListenForMessages(&config, minioClient)
	}
	app.srv.router = SetupRouter(h, tokenManger)
	app.pgDb = db
	app.tokenManager = tokenManger
	app.srv.handler = h
}

func SetupRouter(handler *handler.Handler, tm *token.Manager) *chi.Mux {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "null"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Use(tm.VerifyJwtToken)
	router.Post("/image", handler.UploadImage)
	router.Get("/image/{imageId}", handler.GetImage)
	return router
}
func ListenForMessages(config *Config, m *minio.Client) {
	conn := connectToRabbitMQ(config.rabbitmqUrl)

	channel, err := conn.Channel()
	if err != nil {
		return
	}
	err = channel.Qos(1, 0, false)
	if err != nil {
		return
	}
	messages, err := channel.Consume("image-service", "", false, false, false, false, nil)
	if err != nil {
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

func connectToRabbitMQ(rabbitUrl string) *amqp.Connection {
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
