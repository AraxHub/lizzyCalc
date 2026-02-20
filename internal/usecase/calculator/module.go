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
	repo      ports.IOperationRepository
	cache     ports.ICache
	broker    ports.IProducer
	analytics ports.IOperationAnalytics
	log       *slog.Logger
}

// New создаёт юзкейс калькулятора. broker и analytics могут быть nil.
func New(repo ports.IOperationRepository, cache ports.ICache, broker ports.IProducer, analytics ports.IOperationAnalytics, log *slog.Logger) *UseCase {
	return &UseCase{repo: repo, cache: cache, broker: broker, analytics: analytics, log: log}
}
