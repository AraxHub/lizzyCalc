package calculator

import (
	"log/slog"

	"lizzyCalc/internal/ports"
)

// UseCase — бизнес-логика калькулятора.
type UseCase struct {
	repo ports.OperationRepository
	log  *slog.Logger
}

// New создаёт юзкейс калькулятора.
func New(repo ports.OperationRepository, log *slog.Logger) *UseCase {
	return &UseCase{repo: repo, log: log}
}
