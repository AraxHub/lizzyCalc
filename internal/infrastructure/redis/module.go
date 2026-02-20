package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config — настройки подключения к Redis.
type Config struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     string `envconfig:"PORT" default:"6379"`
	Password string `envconfig:"PASSWORD" default:""`
	DB       int    `envconfig:"DB" default:"0"`
}

// Addr возвращает адрес "host:port".
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Client — обёртка над redis.Client.
type Client struct {
	*redis.Client
}

// New подключается к Redis по конфигу и проверяет пингом.
func New(cfg *Config) (*Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := cli.Ping(context.Background()).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{Client: cli}, nil
}

// Close закрывает соединение.
func (c *Client) Close() error {
	return c.Client.Close()
}

// Ping проверяет соединение (для readiness).
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
