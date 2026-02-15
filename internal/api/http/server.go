package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
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
	// CORS-мидлварь: зачем и как — см. комментарий в конце файла.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
	}))
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

// --- CORS: полная логика ---
//
// Origin = схема + хост + порт. Страница с http://localhost:3000 и запрос на http://localhost:8080
// — разные origin (порты разные). Такой запрос браузер считает cross-origin и ограничивает.
//
// 1) Preflight (OPTIONS). Перед «непростым» запросом (POST, нестандартные заголовки вроде
//    Content-Type: application/json) браузер сам отправляет OPTIONS на тот же URL и смотрит
//    ответ: есть ли заголовки Access-Control-Allow-Origin, Allow-Methods, Allow-Headers.
//    Если их нет или origin/метод/заголовок не разрешён — основной запрос (POST) не шлёт,
//    в консоли ошибка CORS.
//
// 2) Без мидлвари: на OPTIONS у нас нет маршрута → 404 → браузер считает, что cross-origin
//    запрещён, POST не отправляет.
//
// 3) С мидлварью: запрос перехватывается до роутера. На OPTIONS мидлварь сразу отвечает
//    204 No Content и вешает Access-Control-Allow-Origin (наш origin из списка),
//    Allow-Methods (GET, POST, OPTIONS), Allow-Headers (Origin, Content-Type, Accept).
//    На остальные запросы (GET, POST) мидлварь добавляет к ответу Allow-Origin и при
//    необходимости другие CORS-заголовки. Браузер видит разрешение и пропускает ответ.
//
// 4) AllowOrigins — с каких страниц разрешено слать запросы (фронт на 3000 или Vite 5173).
//    AllowMethods — какие HTTP-методы разрешены. AllowHeaders — какие заголовки запроса
//    разрешены (без этого Content-Type: application/json мог бы быть заблокирован).
