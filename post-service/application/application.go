package application

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bernardn38/socialsphere/post-service/handler"
	"github.com/bernardn38/socialsphere/post-service/imageServiceBroker"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/bernardn38/socialsphere/post-service/token"
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
	queries := post.New(db)
	tokenManger := token.NewManager([]byte(config.jwtSecretKey), config.jwtSigningMethod)

	conn := connectToRabbitMQ(config.rabbitmqUrl)

	emitter, _ := imageServiceBroker.NewEventEmitter(conn)
	h := &handler.Handler{PostDb: queries, TokenManager: tokenManger, Emitter: &emitter}

	app.srv.router = SetupRouter(h, tokenManger)
	app.pgDb = db
	app.tokenManager = tokenManger
	app.srv.handler = h
}

func SetupRouter(handler *handler.Handler, tm *token.Manager) *chi.Mux {
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
	router.Post("/posts", handler.CreatePost)
	router.Get("/users/{userId}/posts", handler.GetPostsPageByUserId)
	router.Get("/posts/{postId}", handler.GetPost)
	router.Delete("/posts/{postId}", handler.DeletePost)
	router.Get("/posts/{postId}/likes", handler.GetLikeCount)
	router.Post("/posts/{postId}/likes", handler.CreatePostLike)
	router.Delete("/posts/{postId}/likes", handler.DeleteLike)
	router.Get("/posts/{postId}/likes/check", handler.CheckLike)
	router.Post("/posts/{postId}/comments", handler.CreateComment)
	router.Get("/posts/{postId}/comments", handler.GetAllPostComments)
	return router
}

func connectToRabbitMQ(rabbitUrl string) *amqp.Connection {
	backOff := time.Second * 5
	for {
		conn, err := amqp.Dial(rabbitUrl)
		if err != nil {
			log.Println("Connection not ready backing off for:", backOff)
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Connected to rabbit ")
			return conn
		}
	}
}
