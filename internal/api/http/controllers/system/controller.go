package system

import (
	"log/slog"
	"net/http"

	"lizzyCalc/internal/ports"

	"github.com/gin-gonic/gin"
)

// Controller — системные маршруты: liveness, readiness, корень API.
type Controller struct {
	repo ports.OperationRepository
	log  *slog.Logger
}

// New создаёт системный контроллер.
func New(repo ports.OperationRepository, log *slog.Logger) *Controller {
	return &Controller{repo: repo, log: log}
}

// RegisterRoutes реализует http.Controller: регистрирует маршруты на роутере.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	r.GET("/liveness", c.live)
	r.GET("/readyness", c.ready)
}

func (c *Controller) live(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func (c *Controller) ready(ctx *gin.Context) {
	if err := c.repo.Ping(ctx.Request.Context()); err != nil {
		c.log.Warn("ready check failed", "error", err)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
}
