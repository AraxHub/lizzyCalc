// Package testutil содержит хелперы для интеграционных тестов.
package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer — обёртка над testcontainers PostgreSQL.
type PostgresContainer struct {
	*postgres.PostgresContainer
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewPostgresContainer поднимает PostgreSQL в Docker и возвращает параметры подключения.
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	const (
		user     = "test"
		password = "test"
		dbName   = "testdb"
	)

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("postgres port: %w", err)
	}

	return &PostgresContainer{
		PostgresContainer: container,
		Host:              host,
		Port:              port.Port(),
		User:              user,
		Password:          password,
		DBName:            dbName,
	}, nil
}

// DSN возвращает строку подключения для lib/pq.
func (c *PostgresContainer) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

// RedisContainer — обёртка над testcontainers Redis.
type RedisContainer struct {
	*redis.RedisContainer
	Host string
	Port string
}

// NewRedisContainer поднимает Redis в Docker и возвращает параметры подключения.
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	container, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("redis container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("redis port: %w", err)
	}

	return &RedisContainer{
		RedisContainer: container,
		Host:           host,
		Port:           port.Port(),
	}, nil
}

// Addr возвращает адрес "host:port" для подключения.
func (c *RedisContainer) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// =============================================================================
// MongoDB
// =============================================================================

// MongoContainer — обёртка над testcontainers MongoDB.
type MongoContainer struct {
	*mongodb.MongoDBContainer
	Host string
	Port string
}

// NewMongoContainer поднимает MongoDB в Docker и возвращает параметры подключения.
func NewMongoContainer(ctx context.Context) (*MongoContainer, error) {
	container, err := mongodb.Run(ctx,
		"mongo:7",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("mongo container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("mongo host: %w", err)
	}

	port, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, fmt.Errorf("mongo port: %w", err)
	}

	return &MongoContainer{
		MongoDBContainer: container,
		Host:             host,
		Port:             port.Port(),
	}, nil
}

// URI возвращает строку подключения для mongo-driver.
func (c *MongoContainer) URI() string {
	return fmt.Sprintf("mongodb://%s:%s", c.Host, c.Port)
}

// =============================================================================
// ClickHouse
// =============================================================================

// ClickHouseContainer — обёртка над testcontainers ClickHouse.
type ClickHouseContainer struct {
	*clickhouse.ClickHouseContainer
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewClickHouseContainer поднимает ClickHouse в Docker и возвращает параметры подключения.
func NewClickHouseContainer(ctx context.Context) (*ClickHouseContainer, error) {
	const (
		user     = "default"
		password = ""
		database = "default"
	)

	container, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:24-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(database),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("clickhouse host: %w", err)
	}

	// Нативный порт ClickHouse
	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, fmt.Errorf("clickhouse port: %w", err)
	}

	return &ClickHouseContainer{
		ClickHouseContainer: container,
		Host:                host,
		Port:                port.Port(),
		User:                user,
		Password:            password,
		Database:            database,
	}, nil
}
