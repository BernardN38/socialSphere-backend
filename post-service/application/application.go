package application

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/post-service/handler"
	"github.com/bernardn38/socialsphere/post-service/models"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/middleware"
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
	// middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(h.TokenManager.VerifyJwtToken)

	//routes
	router.Post("/posts", h.CreatePost)
	router.Get("/users/{userId}/posts", h.GetPostsPageByUserId)
	router.Get("/posts/{postId}", h.GetPost)
	router.Delete("/posts/{postId}", h.DeletePost)
	router.Get("/posts/{postId}/likes", h.GetLikeCount)
	router.Post("/posts/{postId}/likes", h.CreatePostLike)
	router.Delete("/posts/{postId}/likes", h.DeleteLike)
	router.Get("/posts/{postId}/likes/check", h.CheckLike)
	router.Post("/posts/{postId}/comments", h.CreateComment)
	router.Get("/posts/{postId}/comments", h.GetAllPostComments)
	return router
}
