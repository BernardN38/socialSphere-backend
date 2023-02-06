package application

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/bernardn38/socialsphere/friend-service/handler"
	"github.com/bernardn38/socialsphere/friend-service/models"
	"github.com/bernardn38/socialsphere/friend-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/friend-service/rpc_broker"
	"github.com/bernardn38/socialsphere/friend-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	// init request handler
	h, err := handler.NewHandler(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	//init app router
	app.srv.router = SetupRouter(h)
	app.srv.port = config.Port

	//init async rabbitmq and rpc workers
	rpc_broker.RunRpcServer(h.UsersDb)
	rabbitmq_broker.RunRabbitBroker(config, h.UsersDb)
}
func SetupRouter(h *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "null"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Use(h.TokenManager.VerifyJwtToken)
	router.Get("/friends/find", h.FindFriends)
	router.Get("/friends/{friendId}/follow", h.CheckFollow)
	router.Get("/friends/latestUploads", h.GetFriendsLastestPhotos)
	router.Post("/friends", h.CreateUser)
	router.Post("/friends/{friendId}/follow", h.CreateFollow)
	return router
}