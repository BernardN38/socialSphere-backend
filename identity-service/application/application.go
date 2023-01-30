package application

import (
	"log"
	"net/http"
	"os"

	"github.com/bernardn38/socialsphere/identity-service/handler"
	"github.com/bernardn38/socialsphere/identity-service/models"
	imageServiceBroker "github.com/bernardn38/socialsphere/identity-service/rabbitmq_broker"
	rpcreceiver "github.com/bernardn38/socialsphere/identity-service/rpc_broker"
	"github.com/bernardn38/socialsphere/identity-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	rpcreceiver.RunRpcServer(*h.UserDb)
	imageServiceBroker.RunRabbitBroker(config, h.UserDb)
}

func SetupRouter(h *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/users", h.CreateUser)
	router.Mount("/", ProtectedRoutes(*h, h.TokenManager))
	return router
}

func ProtectedRoutes(h handler.Handler, tm *token.Manager) http.Handler {
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
	router.Get("/users/{userId}", h.GetUser)
	router.Get("/users/{userId}/profileImage", h.GetUserProfileImage)
	router.Post("/users/profileImage", h.CreateUserProfileImage)
	return router
}
