package application

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/bernardn38/socialsphere/identity-service/handler"
	"github.com/bernardn38/socialsphere/identity-service/helpers"
	"github.com/bernardn38/socialsphere/identity-service/models"
	"github.com/bernardn38/socialsphere/identity-service/rabbitmq_broker"
	rpcbroker "github.com/bernardn38/socialsphere/identity-service/rpc_broker"
	"github.com/bernardn38/socialsphere/identity-service/service"
	"github.com/bernardn38/socialsphere/identity-service/sql/users"
	"github.com/bernardn38/socialsphere/identity-service/token"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

type App struct {
	srv server
}

type server struct {
	router *chi.Mux
	port   string
}

func New() *App {
	app := App{}

	//get configuration from enviroment and validate
	config := helpers.GetEnvConfig()

	//run app setup
	app.runAppSetup(config)
	return &app
}
func (app *App) Run() {
	//start server
	log.Printf("listening on port %s", app.srv.port)
	log.Fatal(http.ListenAndServe(app.srv.port, app.srv.router))
}

func (app *App) runAppSetup(config models.Config) error { //open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// init sqlc user queries
	queries := users.New(db)

	//init rabbitmq message emitter
	rabbitMQConn := rabbitmq_broker.ConnectToRabbitMQ(config.RabbitmqUrl)
	rabbitBroker, err := rabbitmq_broker.NewRabbitBroker(rabbitMQConn)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Initialize dependencies
	service, err := service.New(queries, rabbitBroker, &rpcbroker.RpcClient{})
	if err != nil {
		log.Fatal(err)
	}

	h := handler.NewHandler(service)

	tokenManager := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)
	app.srv.router = SetupRouter(h, tokenManager)
	app.srv.port = config.Port

	//run async rabbit receiver and rpc receiver
	rabbitmq_broker.RunRabbitBroker(config, queries)
	rpcServer := rpcbroker.NewRpcServer(queries)
	go rpcServer.ListenForRpc()
	return nil
}

func SetupRouter(h *handler.Handler, tm *token.Manager) *chi.Mux {
	router := chi.NewRouter()
	router.Mount("/", ProtectedRoutes(*h, tm))
	return router
}

func ProtectedRoutes(h handler.Handler, tm *token.Manager) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(tm.VerifyJwtToken)
	router.Get("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Server is up and running"))
	})
	router.Get("/api/v1/users/{userId}", h.GetUser)
	router.Get("/api/v1/users/{userId}/profileImage", h.GetUserProfileImage)
	router.Get("/api/v1/users/profileImage", h.GetOwnProfileImage)
	router.Post("/api/v1/users/profileImage", h.CreateUserProfileImage)
	return router
}
