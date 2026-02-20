package calculator

import (
	"context"
	"encoding/json"
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
	} else {
		u.log.Info("operation saved", "key", key, "result", result)
	}
	if err := u.cache.Set(ctx, key, result); err != nil {
		return nil, err
	}

	value, err := json.Marshal(op)
	if err != nil {
		return nil, err
	}

	if err := u.broker.Send(ctx, []byte(key), value); err != nil {
		u.log.Warn("broker send", "key", key, "error", err)
	} else {
		u.log.Info("operation published", "key", key, "result", result)
	}

	return &op, nil
}

// History — история операций (обвязка над репозиторием).
func (u *UseCase) History(ctx context.Context) ([]domain.Operation, error) {
	return u.repo.GetHistory(ctx)
}

// HandleOperationEvent вызывается консьюмером при получении сообщения из топика operations (часть ICalculatorUseCase).
func (u *UseCase) HandleOperationEvent(ctx context.Context, op domain.Operation) error {
	if err := u.analytics.WriteOperation(ctx, op); err != nil {
		u.log.Warn("analytics write", "error", err)
		return err
	}
	u.log.Info("operation stored to click", "number1", op.Number1, "operation", op.Operation, "number2", op.Number2, "result", op.Result)

	return nil
}
