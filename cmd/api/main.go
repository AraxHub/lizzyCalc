package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response структура для успешного ответа v1
type Response struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ResponseV2 расширенная структура для v2 с дополнительными данными
type ResponseV2 struct {
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	Environment string            `json:"environment"`
	Metadata    map[string]string `json:"metadata"`
}

// startTime хранит время запуска сервера для расчёта uptime
var startTime = time.Now()

// healthHandler возвращает статус 200 с информацией о сервисе (v1)
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:    "ok",
		Message:   "Service is healthy",
		Timestamp: time.Now(),
	})
}

// healthHandlerV2 возвращает расширенную информацию о сервисе (v2)
// Версионирование API позволяет развивать API без breaking changes для старых клиентов
// v1 остаётся стабильным, а v2 добавляет новые поля и функциональность
func healthHandlerV2(c *gin.Context) {
	uptime := time.Since(startTime)
	
	c.JSON(http.StatusOK, ResponseV2{
		Status:      "ok",
		Message:     "Service is healthy",
		Timestamp:   time.Now(),
		Version:     "2.0.0",
		Uptime:      uptime.String(),
		Environment: "production",
		Metadata: map[string]string{
			"service_name": "lizzyCalc API",
			"api_version":  "v2",
			"build_date":   "2026-02-08",
			"go_version":   "1.25.5",
		},
	})
}

// setupRouter настраивает роутер и регистрирует маршруты
func setupRouter() *gin.Engine {
	// Устанавливаем режим release для production
	gin.SetMode(gin.ReleaseMode)

	// Создаём роутер
	r := gin.New()

	// Middleware для логирования запросов
	r.Use(gin.Logger())

	// Middleware для восстановления после паники
	r.Use(gin.Recovery())

	// Версионирование API - позволяет поддерживать несколько версий одновременно
	// v1 - базовая версия, стабильная, для старых клиентов
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/health", healthHandler)
	}

	// v2 - расширенная версия с дополнительными данными
	// Клиенты могут выбирать нужную версию через URL
	apiV2 := r.Group("/api/v2")
	{
		apiV2.GET("/health", healthHandlerV2)
	}

	// Корневой маршрут
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "lizzyCalc API",
			"version": "1.0.0",
		})
	})

	return r
}

func main() {
	// Хардкод конфигурации
	serverHost := "0.0.0.0"
	serverPort := "8080"
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second

	// Настраиваем роутер
	router := setupRouter()

	// Создаём HTTP сервер с настройками для production
	server := &http.Server{
		Addr:           serverHost + ":" + serverPort,
		Handler:        router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Запускаем сервер
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
