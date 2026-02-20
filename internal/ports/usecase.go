package ports

//go:generate mockgen -source=usecase.go -destination=../mocks/usecase_mock.go -package=mocks

import (
	"context"

	"lizzyCalc/internal/domain"
)

// ICalculatorUseCase — контракт бизнес-логики калькулятора (расчёт, история, обработка событий из Kafka и т.п.).
type ICalculatorUseCase interface {
	Calculate(ctx context.Context, number1, number2 float64, operation string) (*domain.Operation, error)
	History(ctx context.Context) ([]domain.Operation, error)
	HandleOperationEvent(ctx context.Context, op domain.Operation) error
}
