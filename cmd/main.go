package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const success string = "Ответ рассчитан"

// Request структура для входящего JSON
type Request struct {
	Number1   float64 `json:"number1"`
	Number2   float64 `json:"number2"`
	Operation string  `json:"operation"`
}

// Response структура для ответа
type Response struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

func main() {
	// Создаём роутер gin
	r := gin.Default()

	// Регистрируем хэндлер
	r.POST("/calculate", calculateHandler)

	// Запускаем сервер на порту 8080
	r.Run(":8080")
}

// calculateHandler обрабатывает запросы на вычисление
func calculateHandler(c *gin.Context) {
	var req Request

	// парсим json в структуру
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "Неверный формат JSON: " + err.Error()})
		return
	}

	msg, result := calculateUsecase(req)

	// Проверяем на ошибку
	if msg != "Ответ рассчитан" {
		c.JSON(http.StatusBadRequest, Response{Error: msg})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, Response{Result: result})
}

// calculateUsecase выполняет вычисление на основе запроса
func calculateUsecase(req Request) (string, float64) {
	switch req.Operation {
	case "+":
		return success, req.Number1 + req.Number2
	case "-":
		return success, req.Number1 - req.Number2
	case "*":
		return success, req.Number1 * req.Number2
	case "/":
		if req.Number2 == 0 {
			return "Деление на ноль невозможно", 0
		}
		return success, req.Number1 / req.Number2
	default:
		return "Неизвестная операция: " + req.Operation, 0
	}
}
