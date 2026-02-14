package calculator

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"lizzyCalc/internal/ports"
)

// Controller — маршруты калькулятора: calculate, history.
type Controller struct {
	uc  ports.CalculatorUseCase
	log *slog.Logger
}

// New создаёт контроллер калькулятора.
func New(uc ports.CalculatorUseCase, log *slog.Logger) *Controller {
	return &Controller{uc: uc, log: log}
}

// RegisterRoutes реализует http.Controller: регистрирует маршруты на роутере.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	
	api.POST("/calculate", c.calculate)
	api.GET("/history", c.history)
}

func (c *Controller) calculate(ctx *gin.Context) {
	var req CalculateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("calculate bind failed", "error", err)
		ctx.JSON(http.StatusBadRequest, CalculateResponse{Message: "invalid request: " + err.Error()})
		return
	}

	op, err := c.uc.Calculate(ctx.Request.Context(), req.Number1, req.Number2, req.Operation)
	if err != nil {
		c.log.Error("calculate failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, CalculateResponse{Message: err.Error()})
		return
	}
	if op == nil {
		ctx.JSON(http.StatusOK, CalculateResponse{})
		return
	}
	ctx.JSON(http.StatusOK, CalculateResponse{Result: op.Result, Message: op.Message})
}

func (c *Controller) history(ctx *gin.Context) {
	list, err := c.uc.History(ctx.Request.Context())
	if err != nil {
		c.log.Error("history failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]HistoryItem, len(list))
	for i, op := range list {
		items[i] = HistoryItem{
			ID:        op.ID,
			Number1:   op.Number1,
			Number2:   op.Number2,
			Operation: op.Operation,
			Result:    op.Result,
			Message:   op.Message,
			Timestamp: op.Timestamp,
		}
	}
	ctx.JSON(http.StatusOK, HistoryResponse{Items: items})
}
