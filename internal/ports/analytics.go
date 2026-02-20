package ports

//go:generate mockgen -source=analytics.go -destination=../mocks/analytics_mock.go -package=mocks

import (
	"context"

	"lizzyCalc/internal/domain"
)

// IOperationAnalytics — запись операций в хранилище для аналитики (например, ClickHouse).
type IOperationAnalytics interface {
	WriteOperation(ctx context.Context, op domain.Operation) error
}
