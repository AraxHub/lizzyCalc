package pg

import (
	"context"
	"log/slog"

	"lizzyCalc/internal/domain"
)

// OperationRepo реализует ports.OperationRepository для PostgreSQL.
type OperationRepo struct {
	db  *DB
	log *slog.Logger
}

// NewOperationRepo возвращает репозиторий операций.
func NewOperationRepo(db *DB, log *slog.Logger) *OperationRepo {
	return &OperationRepo{db: db, log: log}
}

// SaveOperation сохраняет операцию в БД.
func (r *OperationRepo) SaveOperation(ctx context.Context, op domain.Operation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO operations (number1, number2, operation, result, message, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		op.Number1, op.Number2, op.Operation, op.Result, op.Message, op.Timestamp)
	if err != nil {
		r.log.Debug("SaveOperation failed", "error", err)
		return err
	}
	return nil
}

// GetHistory возвращает историю операций из БД (последние сначала).
func (r *OperationRepo) GetHistory(ctx context.Context) ([]domain.Operation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, number1, number2, operation, result, message, created_at
		 FROM operations ORDER BY created_at DESC`)
	if err != nil {
		r.log.Debug("GetHistory failed", "error", err)
		return nil, err
	}
	defer rows.Close()
	var list []domain.Operation
	for rows.Next() {
		var op domain.Operation
		err := rows.Scan(&op.ID, &op.Number1, &op.Number2, &op.Operation, &op.Result, &op.Message, &op.Timestamp)
		if err != nil {
			return nil, err
		}
		list = append(list, op)
	}
	return list, rows.Err()
}

// Ping проверяет доступность БД (readiness).
func (r *OperationRepo) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
