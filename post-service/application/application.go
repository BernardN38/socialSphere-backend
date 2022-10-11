package application

import (
	"database/sql"
	"github.com/bernardn38/socialsphere/post-service/handler"
	"github.com/bernardn38/socialsphere/post-service/imageServiceBroker"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/bernardn38/socialsphere/post-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	jwtSecretKey     string
	jwtSigningMethod jwt.Algorithm
	dsn              string
	rabbitmqUrl      string
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
	config := Config{jwtSecretKey: "superSecretKey", jwtSigningMethod: jwt.HS256, dsn: dsn, rabbitmqUrl: "amqp://guest:guest@rabbitmq"}
	app.runAppSetup(config)
	return &app
}
func (app *App) Run() {
	log.Printf("listening on port %s", "9001")
	log.Fatal(http.ListenAndServe(":9001", app.srv.router))
}

func (app *App) runAppSetup(config Config) {
	db, err := sql.Open("postgres", config.dsn)
	if err != nil {
		log.Fatal(err)
	}
	queries := post.New(db)
	tokenManger := token.NewManager([]byte(config.jwtSecretKey), config.jwtSigningMethod)

	conn := connectToRabbitMQ(config.rabbitmqUrl)
	err = declareChannelQueue(conn)
	if err != nil {
		log.Fatal(err)
	}

	emitter, _ := imageServiceBroker.NewEventEmitter(conn)
	h := &handler.Handler{PostDb: queries, TokenManager: tokenManger, Emitter: &emitter}

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
	router.Post("/posts", handler.CreatePost)
	router.Get("/posts", handler.GetAllPosts)
	router.Get("/posts/{id}", handler.GetPost)
	return router
}

func declareChannelQueue(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare("image-service", "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}
	q, err := ch.QueueDeclare("image-service", true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.QueueBind(q.Name, "image-service", "image-service", false, nil)
	if err != nil {
		return err
	}
	return nil
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
