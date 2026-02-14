package calculator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"lizzyCalc/internal/domain"
)

// Calculate — проверяет кэш; при промахе считает, сохраняет в БД и в кэш, возвращает результат.
func (u *UseCase) Calculate(ctx context.Context, number1, number2 float64, operation string) (*domain.Operation, error) {
	key := cacheKey(number1, number2, operation)
	if cached, found, err := u.cache.Get(ctx, key); err == nil && found {
		return &domain.Operation{
			Number1:   number1,
			Number2:   number2,
			Operation: operation,
			Result:    cached,
			Message:   "",
			Timestamp: time.Now(),
		}, nil
	}

	var result float64
	var message string
	switch operation {
	case domain.OpAdd:
		result = number1 + number2
	case domain.OpSub:
		result = number1 - number2
	case domain.OpMul:
		result = number1 * number2
	case domain.OpDiv:
		if number2 == 0 {
			return nil, errors.New("division by zero")
		}
		result = number1 / number2
	default:
		return nil, fmt.Errorf("%w: %s", domain.ErrUnknownOperation, operation)
	}
	op := domain.Operation{
		Number1:   number1,
		Number2:   number2,
		Operation: operation,
		Result:    result,
		Message:   message,
		Timestamp: time.Now(),
	}
	if err := u.repo.SaveOperation(ctx, op); err != nil {
		return nil, err
	}
	if err := u.cache.Set(ctx, key, result); err != nil {
		return nil, err
	}
	return &op, nil
}

// History — история операций (обвязка над репозиторием).
func (u *UseCase) History(ctx context.Context) ([]domain.Operation, error) {
	return u.repo.GetHistory(ctx)
}
