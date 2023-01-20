package application

import (
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/messaging-service/handler"
	"github.com/bernardn38/socialsphere/messaging-service/token"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-redis/redis/v8"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type Config struct {
	jwtSecretKey     string
	jwtSigningMethod jwt.Algorithm
}
type App struct {
	srv          server
	tokenManager *token.Manager
}

type server struct {
	router  *chi.Mux
	handler *handler.Handler
}

func New() *App {
	app := App{}
	config := Config{jwtSecretKey: "superSecretKey", jwtSigningMethod: jwt.HS256}
	app.runAppSetup(config)
	return &app
}
func (app *App) Run() {
	log.Printf("listening on port %s", "8081")
	log.Fatal(http.ListenAndServe(":8081", app.srv.router))
}

func (app *App) runAppSetup(config Config) {
	tokenManger := token.NewManager([]byte(config.jwtSecretKey), config.jwtSigningMethod)
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "password",
		DB:       0,
	})
	h := &handler.Handler{TokenManager: tokenManger, Upgrader: upgrader, Conns: make(map[string]*websocket.Conn), Rdb: rdb}

	app.srv.router = SetupRouter(h, tokenManger)
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
	router.Get("/messaging", handler.SendMessage)
	router.Get("/users/{userId}/checkOnline", handler.CheckOnline)
	return router
}
