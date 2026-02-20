package calculator

import (
	"fmt"
	"time"

	"lizzyCalc/internal/domain"
)

// CalculateRequest — запрос на вычисление.
// @Description Параметры для выполнения арифметической операции
type CalculateRequest struct {
	Number1   float64 `json:"number1" binding:"required" example:"10.5"`
	Number2   float64 `json:"number2" binding:"required" example:"5.2"`
	Operation string  `json:"operation" binding:"required" example:"+" enums:"+,-,*,/"`
}

// Validate проверяет запрос: операция должна быть одной из +, -, *, /.
func (r *CalculateRequest) Validate() error {
	switch r.Operation {
	case domain.OpAdd, domain.OpSub, domain.OpMul, domain.OpDiv:
		return nil
	default:
		return fmt.Errorf("invalid operation: %s", r.Operation)
	}
}

// CalculateResponse — ответ с результатом вычисления.
// @Description Результат арифметической операции
type CalculateResponse struct {
	Result  float64 `json:"result" example:"15.7"`
	Message string  `json:"message,omitempty" example:"division by zero"`
}

// HistoryItem — одна запись в истории.
// @Description Информация об одной выполненной операции
type HistoryItem struct {
	ID        int       `json:"id" example:"1"`
	Number1   float64   `json:"number1" example:"10.5"`
	Number2   float64   `json:"number2" example:"5.2"`
	Operation string    `json:"operation" example:"+"`
	Result    float64   `json:"result" example:"15.7"`
	Message   string    `json:"message,omitempty" example:""`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// HistoryResponse — ответ со списком операций.
// @Description Список всех выполненных операций
type HistoryResponse struct {
	Items []HistoryItem `json:"items"`
}

// ErrorResponse — ответ с ошибкой.
// @Description Сообщение об ошибке
type ErrorResponse struct {
	Error string `json:"error" example:"internal server error"`
}
