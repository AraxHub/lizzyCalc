package ports

import (
	"context"

	"lizzyCalc/internal/domain"
)

// CalculatorUseCase — контракт бизнес-логики калькулятора.
type CalculatorUseCase interface {
	Calculate(ctx context.Context, number1, number2 float64, operation string) (*domain.Operation, error)
	History(ctx context.Context) ([]domain.Operation, error)
}
