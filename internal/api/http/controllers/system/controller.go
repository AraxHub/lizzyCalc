package system

import (
	"log/slog"
	"net/http"

	"lizzyCalc/internal/ports"

	"github.com/gin-gonic/gin"
)

// Controller — системные маршруты: liveness, readiness, корень API.
type Controller struct {
	repo ports.IOperationRepository
	log  *slog.Logger
}

// New создаёт системный контроллер.
func New(repo ports.IOperationRepository, log *slog.Logger) *Controller {
	return &Controller{repo: repo, log: log}
}

// RegisterRoutes реализует http.Controller: регистрирует маршруты на роутере.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	r.GET("/liveness", c.live)
	r.GET("/readyness", c.ready)
}

// @Summary Проверка жизнеспособности
// @Description Возвращает статус "alive" если сервер запущен
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string "status: alive"
// @Router /liveness [get]
func (c *Controller) live(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// @Summary Проверка готовности
// @Description Проверяет подключение к БД и возвращает статус готовности
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string "status: ready"
// @Failure 503 {object} map[string]string "status: not ready, error: ..."
// @Router /readyness [get]
func (c *Controller) ready(ctx *gin.Context) {
	if err := c.repo.Ping(ctx.Request.Context()); err != nil {
		c.log.Warn("ready check failed", "error", err)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
}
