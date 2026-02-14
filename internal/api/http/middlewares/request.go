package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger логирует каждый запрос: метод, путь, статус, длительность, client IP.
func RequestLogger(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	clientIP := c.ClientIP()
	method := c.Request.Method

	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()
	if raw != "" {
		path = path + "?" + raw
	}
	slog.Info("request",
		"method", method,
		"path", path,
		"status", status,
		"ip", clientIP,
		"latency_ms", latency.Milliseconds(),
	)
}
