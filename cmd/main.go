package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Request структура для входящего запроса
type Request struct {
	Number1   float64 `json:"number1"`
	Number2   float64 `json:"number2"`
	Operation string  `json:"operation"`
}

// Operation структура для хранения истории операций
type Operation struct {
	Number1   float64   `json:"number1"`
	Number2   float64   `json:"number2"`
	Operation string    `json:"operation"`
	Result    float64   `json:"result"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Response структура для ответа
type Response struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

// operations - слайс для хранения истории операций (аналог БД)
var operations []Operation

// success - константа успешного выполнения операции
const success = "Ответ рассчитан"

// calculateUsecase выполняет вычисление на основе запроса и сохраняет в БД
func calculateUsecase(req Request) (string, float64) {
	var msg string
	var result float64

	// Выполняем операцию
	switch req.Operation {
	case "+":
		msg = success
		result = req.Number1 + req.Number2
	case "-":
		msg = success
		result = req.Number1 - req.Number2
	case "*":
		msg = success
		result = req.Number1 * req.Number2
	case "/":
		if req.Number2 == 0 {
			msg = "Деление на ноль невозможно"
			result = 0
		} else {
			msg = success
			result = req.Number1 / req.Number2
		}
	default:
		msg = "Неизвестная операция: " + req.Operation
		result = 0
	}

	// Сохраняем операцию в БД (слайс)
	operation := Operation{
		Number1:   req.Number1,
		Number2:   req.Number2,
		Operation: req.Operation,
		Result:    result,
		Message:   msg,
		Timestamp: time.Now(),
	}
	operations = append(operations, operation)

	return msg, result
}

// calculateHandler обрабатывает запросы на вычисление
func calculateHandler(c *gin.Context) {
	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "Неверный формат JSON: " + err.Error()})
		return
	}

	msg, result := calculateUsecase(req)
	if msg != success {
		c.JSON(http.StatusBadRequest, Response{Error: msg})
		return
	}

	c.JSON(http.StatusOK, Response{Result: result})
}

// historyHandler выводит историю всех операций
func historyHandler(c *gin.Context) {
	if len(operations) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":    "История операций пуста",
			"operations": []Operation{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      len(operations),
		"operations": operations,
	})
}

func main() {
	// Создаём роутер gin
	r := gin.Default()

	// Регистрируем хэндлеры
	r.POST("/calculate", calculateHandler)
	r.GET("/history", historyHandler)

	// Запускаем сервер на порту 8080
	r.Run(":8080")
}