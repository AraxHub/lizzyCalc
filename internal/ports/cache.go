package ports

import "context"

// ICache — контракт кэша результатов операций. Ключ — операция, значение — результат.
// Ключи уникальны, дубликаты не хранятся.
type ICache interface {
	Get(ctx context.Context, key string) (value float64, found bool, err error)
	Set(ctx context.Context, key string, value float64) error
}
