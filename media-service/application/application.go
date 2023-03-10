package application

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/image-service/handler"
	"github.com/bernardn38/socialsphere/image-service/models"
	"github.com/bernardn38/socialsphere/image-service/rabbitmq_broker"
	"github.com/bernardn38/socialsphere/image-service/rpc_broker"
	"github.com/cristalhq/jwt/v4"
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
	postgresUrl := os.Getenv("DSN")
	jwtSecret := os.Getenv("jwtSecret")
	rabbitMQUrl := os.Getenv("rabbitMQUrl")
	minioKey := os.Getenv("minioKey")
	minioSecret := os.Getenv("minioSecret")
	port := os.Getenv("port")
	config := models.Config{
		JwtSecretKey:     jwtSecret,
		JwtSigningMethod: jwt.Algorithm(jwt.HS256),
		PostgresUrl:      postgresUrl,
		RabbitmqUrl:      rabbitMQUrl,
		MinioKey:         minioKey,
		MinioSecret:      minioSecret,
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
	server := http.Server{
		Addr:        app.srv.port,
		Handler:     app.srv.router,
		ReadTimeout: 30 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
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

	//init async rabbitmq worker
	rabbitmq_broker.RunRabbitBroker(config)
	rpc_broker.RunRpcServer(h.UserImageDB, h.MinioClient)
}

func SetupRouter(h *handler.Handler) *chi.Mux {
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
	router.Use(h.TokenManager.VerifyJwtToken)

	router.Post("/api/v1/images", h.UploadImage)
	router.Get("/api/v1/images/{imageId}", h.GetImage)
	return router
}
