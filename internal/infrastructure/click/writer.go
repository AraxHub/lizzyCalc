package click

import (
	"context"
	"fmt"

	"lizzyCalc/internal/domain"
)

const operationsAnalyticsFull = "default.operations_analytics"

// OperationWriter записывает операции в ClickHouse в формате, удобном для аналитики (GROUP BY operation, по времени и т.д.).
type OperationWriter struct {
	db *Client
}

// NewOperationWriter создаёт писатель операций для аналитики.
func NewOperationWriter(db *Client) *OperationWriter {
	return &OperationWriter{db: db}
}

// EnsureTable создаёт таблицу операций для аналитики в default, если её ещё нет. Вызови один раз при старте приложения.
func (w *OperationWriter) EnsureTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			number1 Float64,
			number2 Float64,
			operation String,
			result Float64,
			message String,
			created_at DateTime64(3)
		) ENGINE = MergeTree()
		ORDER BY (created_at, operation)
		PARTITION BY toYYYYMM(created_at)`,
		operationsAnalyticsFull,
	)
	_, err := w.db.DB().ExecContext(ctx, query)
	return err
}

// WriteOperation реализует ports.IOperationAnalytics: пишет одну операцию в ClickHouse.
func (w *OperationWriter) WriteOperation(ctx context.Context, op domain.Operation) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (number1, number2, operation, result, message, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		operationsAnalyticsFull,
	)
	_, err := w.db.DB().ExecContext(ctx, query,
		op.Number1, op.Number2, op.Operation, op.Result, op.Message, op.Timestamp)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}
	return nil
}
