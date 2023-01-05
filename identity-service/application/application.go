package application

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/identity-service/handler"
	"github.com/bernardn38/socialsphere/identity-service/imageServiceBroker"
	"github.com/bernardn38/socialsphere/identity-service/sql/users"
	"github.com/bernardn38/socialsphere/identity-service/token"
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
	log.Printf(config.dsn)
	db, err := sql.Open("postgres", config.dsn)
	if err != nil {
		log.Fatal(err)
	}
	conn := connectToRabbitMQ(config.rabbitmqUrl)

	emitter, _ := imageServiceBroker.NewEventEmitter(conn)
	queries := users.New(db)
	tokenManger := token.NewManager([]byte(config.jwtSecretKey), config.jwtSigningMethod)
	h := &handler.Handler{UserDb: queries, TokenManager: tokenManger, Emitter: &emitter}

	app.srv.router = SetupRouter(h, tokenManger)
	app.pgDb = db
	app.tokenManager = tokenManger
	app.srv.handler = h
}

func SetupRouter(handler *handler.Handler, tm *token.Manager) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/users", handler.CreateUser)
	router.Mount("/", ProtectedRoutes(*handler, tm))
	return router
}

func ProtectedRoutes(handler handler.Handler, tm *token.Manager) http.Handler {
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
	router.Get("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Server is up and running"))
	})
	router.Get("/users/{userId}", handler.GetUser)
	router.Get("/users/{userId}/profileImage", handler.GetUserProfileImage)
	router.Post("/users/profileImage", handler.CreateUserProfileImage)
	return router
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
