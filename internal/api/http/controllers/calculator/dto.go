package calculator

import "time"

// CalculateRequest — запрос на вычисление (для POST /api/calculate).
type CalculateRequest struct {
	Number1   float64 `json:"number1" binding:"required"`
	Number2   float64 `json:"number2" binding:"required"`
	Operation string  `json:"operation" binding:"required"`
}

// CalculateResponse — ответ с результатом.
type CalculateResponse struct {
	Result  float64 `json:"result"`
	Message string  `json:"message,omitempty"`
}

// HistoryItem — одна запись в истории (для GET /api/history).
type HistoryItem struct {
	ID        int       `json:"id"`
	Number1   float64   `json:"number1"`
	Number2   float64   `json:"number2"`
	Operation string    `json:"operation"`
	Result    float64   `json:"result"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// HistoryResponse — ответ со списком операций.
type HistoryResponse struct {
	Items []HistoryItem `json:"items"`
}
