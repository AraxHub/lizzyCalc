package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"lizzyCalc/internal/api/http/middlewares"
)

// ServerConfig — настройки HTTP-сервера. Переменные: APP_SERVER_HOST, APP_SERVER_PORT.
type ServerConfig struct {
	Host string `env:"HOST" default:"0.0.0.0"`
	Port string `env:"PORT" default:"8080"`
}

// Controller — контракт: контроллер регистрирует свои маршруты на роутере.
type Controller interface {
	RegisterRoutes(r *gin.Engine)
}

// Server — API-сервер: конфиг и список контроллеров.
type Server struct {
	cfg         ServerConfig
	controllers []Controller
	srv         *http.Server
}

// NewServer создаёт сервер с конфигом.
func NewServer(cfg ServerConfig) *Server {
	return &Server{cfg: cfg, controllers: nil}
}

// AddController добавляет один или несколько контроллеров.
func (s *Server) AddController(c ...Controller) {
	s.controllers = append(s.controllers, c...)
}

// Start поднимает роутер, запускает сервер и блокируется до отмены ctx (SIGINT/SIGTERM), затем делает graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.RequestLogger)
	for _, c := range s.controllers {
		c.RegisterRoutes(r)
	}

	s.srv = &http.Server{
		Addr:         s.cfg.Host + ":" + s.cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		_ = s.srv.ListenAndServe()
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
