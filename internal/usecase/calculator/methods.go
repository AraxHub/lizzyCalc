package calculator

import (
	"context"

	"lizzyCalc/internal/domain"
)

// Calculate — вычисление операции (пока заглушка: без реализации и без сохранения).
func (u *UseCase) Calculate(ctx context.Context, number1, number2 float64, operation string) (*domain.Operation, error) {
	_ = ctx
	return nil, nil
}

// History — история операций (обвязка над репозиторием).
func (u *UseCase) History(ctx context.Context) ([]domain.Operation, error) {
	return u.repo.GetHistory(ctx)
}
