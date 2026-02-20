package pg

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Config — настройки подключения к PostgreSQL.
type Config struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     string `envconfig:"PORT" default:"5433"`
	User     string `envconfig:"USER" default:"postgres"`
	Password string `envconfig:"PASSWORD" default:"postgres"`
	DBName   string `envconfig:"NAME" default:"lizzycalc"`
	SSLMode  string `envconfig:"SSLMODE" default:"disable"`
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
