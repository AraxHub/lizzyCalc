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

// SaveOperation сохраняет операцию в БД. Пока без логики.
func (r *OperationRepo) SaveOperation(ctx context.Context, op domain.Operation) error {
	_ = ctx
	r.log.Debug("SaveOperation", "op", op)
	return nil
}

// GetHistory возвращает историю операций из БД. Пока без логики.
func (r *OperationRepo) GetHistory(ctx context.Context) ([]domain.Operation, error) {
	_ = ctx
	r.log.Debug("GetHistory")
	return nil, nil
}

// Ping проверяет доступность БД (readiness).
func (r *OperationRepo) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
