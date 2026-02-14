package pg

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Config — настройки подключения к PostgreSQL.
type Config struct {
	Host     string `env:"HOST" default:"localhost"`
	Port     string `env:"PORT" default:"5433"`
	User     string `env:"USER" default:"postgres"`
	Password string `env:"PASSWORD" default:"postgres"`
	DBName   string `env:"NAME" default:"lizzycalc"`
	SSLMode  string `env:"SSLMODE" default:"disable"`
}

// DSN возвращает строку подключения для lib/pq.
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// DB обёртка над пулом соединений.
type DB struct {
	*sql.DB
}

// New подключается к PostgreSQL по конфигу и проверяет пингом.
func New(cfg *Config) (*DB, error) {
	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("pg open: %w", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}
	return &DB{conn}, nil
}

// Close закрывает пул.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping проверяет соединение с БД (для readiness).
func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
