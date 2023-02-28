package application

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/friend-service/handler"
	"github.com/bernardn38/socialsphere/friend-service/models"
	"github.com/bernardn38/socialsphere/friend-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/friend-service/rpc_broker"
	"github.com/bernardn38/socialsphere/friend-service/service"
	"github.com/bernardn38/socialsphere/friend-service/sql/users"
	"github.com/bernardn38/socialsphere/friend-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
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
	router *chi.Mux
	port   string
}

func New() *App {
	app := App{}

	//get configuration from enviroment and validate
	postgresUrl := os.Getenv("DSN")
	jwtSecret := os.Getenv("jwtSecret")
	rabbitMQUrl := os.Getenv("rabbitMQUrl")
	port := os.Getenv("port")
	config := models.Config{
		JwtSecretKey:     jwtSecret,
		JwtSigningMethod: jwt.Algorithm(jwt.HS256),
		PostgresUrl:      postgresUrl,
		RabbitmqUrl:      rabbitMQUrl,
		Port:             port,
	}
	err := config.Validate()
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}

	//run app setup
	app.runAppSetup(config)
	return &app
}
func (app *App) Run() {
	//start server
	log.Printf("listening on port %s", app.srv.port)
	log.Fatal(http.ListenAndServe(app.srv.port, app.srv.router))
}

func (app *App) runAppSetup(config models.Config) {
	//open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		log.Fatal(err)
		return
	}
	// init sqlc user queries
	queries := users.New(db)

	//init jwt token manager
	tokenManger := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)

	//init rabbitmq message emitter
	rabbitMQConn := rabbitmq_broker.ConnectToRabbitMQ(config.RabbitmqUrl)
	emitter, err := rabbitmq_broker.NewEventEmitter(rabbitMQConn)
	if err != nil {
		log.Fatal(err)
	}

	//init
	friendService := service.New(queries, emitter)

	// init request handler
	h := handler.NewHandler(friendService)
	//init app router
	app.srv.router = SetupRouter(h, tokenManger)
	app.srv.port = config.Port

	go rpc_broker.NewRpcServer(queries).ListenForRpc()
	rabbitmq_broker.RunRabbitBroker(config, queries)
}
func SetupRouter(h *handler.Handler, tm *token.Manager) *chi.Mux {
	router := chi.NewRouter()
	// router.Use(cors.Handler(cors.Options{
	// 	AllowedOrigins:   []string{"https://*", "http://*", "null"},
	// 	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	// 	ExposedHeaders:   []string{"Link"},
	// 	AllowCredentials: true,
	// 	MaxAge:           300, // Maximum value not ignored by any of major browsers
	// }))
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(tm.VerifyJwtToken)
	router.Get("/api/v1/friends/find", h.FindFriends)
	router.Get("/api/v1/friends/{friendId}/follow", h.CheckFollow)
	router.Get("/api/v1/friends/latestUploads", h.GetFriendsLastestPhotos)
	router.Post("/api/v1/friends", h.CreateUser)
	router.Post("/api/v1/friends/{friendId}/follow", h.CreateFollow)
	router.Get("/api/v1/friends", h.GetFriends)
	return router
}
