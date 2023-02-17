package application

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/messaging-service/handler"
	"github.com/bernardn38/socialsphere/messaging-service/models"
	"github.com/cristalhq/jwt/v4"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
	jwtSecret := os.Getenv("jwtSecret")
	port := os.Getenv("port")
	MongoUri := os.Getenv("mongoUri")
	config := models.Config{
		JwtSecretKey:     jwtSecret,
		JwtSigningMethod: jwt.Algorithm(jwt.HS256),
		Port:             port,
		MongoUri:         MongoUri,
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
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(h.TokenManager.VerifyJwtToken)
	router.Get("/messaging", h.HandleMessage)
	router.Get("/messages", h.GetAllMessages)
	router.Get("/users/messages", h.GetAllUserMessages)
	router.Get("/users/{userId}/messages", h.GetUserMessages)
	router.Get("/users/{userId}/checkOnline", h.CheckOnline)
	return router
}
