package ports

import (
	"context"

	"lizzyCalc/internal/domain"
)

// IOperationRepository — контракт сохранения и чтения операций.
type IOperationRepository interface {
	SaveOperation(ctx context.Context, op domain.Operation) error
	GetHistory(ctx context.Context) ([]domain.Operation, error)
	Ping(ctx context.Context) error
}
