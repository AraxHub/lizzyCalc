package click

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Config — настройки подключения к ClickHouse. Переменные: CALCULATOR_CLICKHOUSE_HOST, PORT, DATABASE, USERNAME, PASSWORD.
type Config struct {
	Host     string `env:"HOST" default:"localhost"`
	Port     string `env:"PORT" default:"9000"`
	Database string `env:"DATABASE" default:"default"`
	Username string `env:"USERNAME" default:"default"`
	Password string `env:"PASSWORD" default:""`
}

// Addr возвращает адрес "host:port" для нативного протокола.
func (c *Config) Addr() string {
	if c == nil {
		return "localhost:9000"
	}
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Client — обёртка над sql.DB (драйвер clickhouse). Используй для вставок и запросов аналитики.
type Client struct {
	db *sql.DB
}

// New подключается к ClickHouse по конфигу и проверяет пингом. После использования вызови Close().
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{cfg.Addr()},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
	})
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &Client{db: db}, nil
}

// DB возвращает *sql.DB для выполнения запросов и batch-вставок (PrepareBatch и т.д.).
func (c *Client) DB() *sql.DB {
	return c.db
}

// Close закрывает соединение с ClickHouse.
func (c *Client) Close() error {
	return c.db.Close()
}

// Ping проверяет соединение (для readiness).
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
