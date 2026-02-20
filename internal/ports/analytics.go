package ports

import (
	"context"

	"lizzyCalc/internal/domain"
)

// IOperationAnalytics — запись операций в хранилище для аналитики (например, ClickHouse).
type IOperationAnalytics interface {
	WriteOperation(ctx context.Context, op domain.Operation) error
}
