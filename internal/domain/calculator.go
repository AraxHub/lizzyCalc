package domain

import "time"

// Operation — запись об одной операции калькулятора.
type Operation struct {
	ID        int
	Number1   float64
	Number2   float64
	Operation string
	Result    float64
	Message   string
	Timestamp time.Time
}
