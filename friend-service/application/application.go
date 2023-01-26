package application

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/friend-service/handler"
	rpcreceiver "github.com/bernardn38/socialsphere/friend-service/rpc_receiver"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/bernardn38/socialsphere/friend-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
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
	log.Printf("listening on port %s", "8080")
	log.Fatal(http.ListenAndServe(":8080", app.srv.router))
}

func (app *App) runAppSetup(config Config) {
	db, err := sql.Open("postgres", config.dsn)
	if err != nil {
		log.Fatal(err)
	}

	queries := users.New(db)
	tokenManger := token.NewManager([]byte(config.jwtSecretKey), config.jwtSigningMethod)
	h := &handler.Handler{UsersDb: queries, TokenManager: tokenManger}

	app.srv.router = SetupRouter(h, tokenManger)
	app.pgDb = db
	app.tokenManager = tokenManger
	app.srv.handler = h
	//start workers for recieving rabbitmq messages
	rabbitMQConn := connectToRabbitMQ(config.rabbitmqUrl)
	for i := 0; i < 10; i++ {
		go ListenForMessages(&config, rabbitMQConn)
	}
	rpcReceiver := rpcreceiver.NewRpcReceiver(queries)
	go rpcReceiver.ListenForRpc()
}

func SetupRouter(h *handler.Handler, tm *token.Manager) *chi.Mux {
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
	router.Get("/friends/find", h.FindFriends)
	router.Post("/friends", h.CreateUser)
	router.Post("/friends/friendships/{friendId}", h.CreateFriendship)
	return router
}

func ListenForMessages(config *Config, conn *amqp.Connection) {
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
			log.Println(d.Body)
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
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Successfully connected to rabbitmq")
			return conn
		}
	}
}
