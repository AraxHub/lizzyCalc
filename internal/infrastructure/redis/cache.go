package redis

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
	"lizzyCalc/internal/ports"
)

var _ ports.Cache = (*Cache)(nil)

// Cache реализует ports.Cache через Redis. Ключ — строка операции, значение — результат (float64 как строка).
type Cache struct {
	cli *Client
	log *slog.Logger
}

// NewCache возвращает кэш, реализующий ports.Cache.
func NewCache(cli *Client, log *slog.Logger) *Cache {
	return &Cache{cli: cli, log: log}
}

// Get возвращает результат по ключу. Если ключа нет — found == false.
func (c *Cache) Get(ctx context.Context, key string) (value float64, found bool, err error) {
	s, err := c.cli.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil { // ключа нет
			return 0, false, nil
		}
		c.log.Debug("cache get failed", "key", key, "error", err)
		return 0, false, err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		c.log.Debug("cache parse failed", "key", key, "error", err)
		return 0, false, fmt.Errorf("cache parse value: %w", err)
	}
	return v, true, nil
}

// Set сохраняет результат по ключу. Ключи уникальны, дубликаты перезаписываются.
func (c *Cache) Set(ctx context.Context, key string, value float64) error {
	s := strconv.FormatFloat(value, 'g', -1, 64)
	if err := c.cli.Set(ctx, key, s, 0).Err(); err != nil {
		c.log.Debug("cache set failed", "key", key, "error", err)
		return err
	}
	return nil
}
