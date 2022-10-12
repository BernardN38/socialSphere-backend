package application

import (
	"github.com/bernardn38/socialsphere/identity-service/handler"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	dsn string
}
type App struct {
	srv *Server
}
type Server struct {
	router  *chi.Router
	handler handler.Handler
}
