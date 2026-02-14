package calculator

import (
	"log/slog"
	"strconv"

	"lizzyCalc/internal/ports"
)

// cacheKey формирует читаемый ключ операции для кэша, например "1 + 1".
func cacheKey(number1, number2 float64, operation string) string {
	return strconv.FormatFloat(number1, 'f', -1, 64) + " " + operation + " " + strconv.FormatFloat(number2, 'f', -1, 64)
}

// UseCase — бизнес-логика калькулятора.
type UseCase struct {
	repo  ports.OperationRepository
	cache ports.Cache
	log   *slog.Logger
}

// New создаёт юзкейс калькулятора.
func New(repo ports.OperationRepository, cache ports.Cache, log *slog.Logger) *UseCase {
	return &UseCase{repo: repo, cache: cache, log: log}
}
