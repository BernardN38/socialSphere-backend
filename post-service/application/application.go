package application

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/post-service/handler"
	"github.com/bernardn38/socialsphere/post-service/models"
	"github.com/bernardn38/socialsphere/post-service/service"
	"github.com/bernardn38/socialsphere/post-service/token"
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

func (app *App) runAppSetup(config models.Config) error {
	// Initialize dependencies
	service, err := service.New(&config)
	if err != nil {
		log.Fatal(err)
	}
	tokenManager := token.NewManager([]byte(config.JwtSecretKey), config.JwtSigningMethod)
	h := handler.NewHandler(service)
	app.srv.router = SetupRouter(h, tokenManager)
	app.srv.port = config.Port
	return nil
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
	// middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(tm.VerifyJwtToken)

	//routes
	router.Post("/api/v1/posts", h.CreatePost)
	router.Get("/api/v1/posts/{postId}", h.GetPost)
	router.Get("/api/v1/posts/users/{userId}", h.GetPostsPageByUserId)
	router.Delete("/api/v1/posts/{postId}", h.DeletePost)
	router.Get("/api/v1/posts/{postId}/likes", h.GetLikeCount)
	router.Post("/api/v1/posts/{postId}/likes", h.CreatePostLike)
	router.Delete("/api/v1/posts/{postId}/likes", h.DeleteLike)
	router.Get("/api/v1/posts/{postId}/likes/check", h.CheckLike)
	router.Post("/api/v1/posts/{postId}/comments", h.CreateComment)
	router.Get("/api/v1/posts/{postId}/comments", h.GetAllPostComments)
	return router
}
