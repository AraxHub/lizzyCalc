// Package integration содержит интеграционные тесты с реальной инфраструктурой (PostgreSQL, Redis).
// Тесты используют testcontainers для поднятия Docker-контейнеров.
//
// Запуск:
//
//	go test ./tests/integration/... -v
//
// Пропуск (только юнит-тесты):
//
//	go test ./... -short
package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"lizzyCalc/tests/integration/testutil"
)

// TestMain — точка входа для всех тестов пакета.
// Поднимает контейнеры один раз перед всеми тестами и останавливает после.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// === Setup: поднимаем контейнеры ===
	log.Println("🚀 Поднимаем тестовые контейнеры...")

	var err error

	// PostgreSQL
	pgContainer, err = testutil.NewPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("❌ Не удалось поднять PostgreSQL: %v", err)
	}
	log.Printf("✅ PostgreSQL: %s:%s", pgContainer.Host, pgContainer.Port)

	// Redis
	redisContainer, err = testutil.NewRedisContainer(ctx)
	if err != nil {
		log.Fatalf("❌ Не удалось поднять Redis: %v", err)
	}
	log.Printf("✅ Redis: %s:%s", redisContainer.Host, redisContainer.Port)

	// MongoDB
	mongoContainer, err = testutil.NewMongoContainer(ctx)
	if err != nil {
		log.Fatalf("❌ Не удалось поднять MongoDB: %v", err)
	}
	log.Printf("✅ MongoDB: %s:%s", mongoContainer.Host, mongoContainer.Port)

	// ClickHouse
	clickContainer, err = testutil.NewClickHouseContainer(ctx)
	if err != nil {
		log.Fatalf("❌ Не удалось поднять ClickHouse: %v", err)
	}
	log.Printf("✅ ClickHouse: %s:%s", clickContainer.Host, clickContainer.Port)

	log.Println("🧪 Запускаем тесты...")

	// === Запуск тестов ===
	code := m.Run()

	// === Teardown: останавливаем контейнеры ===
	log.Println("🧹 Останавливаем контейнеры...")

	if pgContainer != nil {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("⚠️  Ошибка остановки PostgreSQL: %v", err)
		}
	}

	if redisContainer != nil {
		if err := redisContainer.Terminate(ctx); err != nil {
			log.Printf("⚠️  Ошибка остановки Redis: %v", err)
		}
	}

	if mongoContainer != nil {
		if err := mongoContainer.Terminate(ctx); err != nil {
			log.Printf("⚠️  Ошибка остановки MongoDB: %v", err)
		}
	}

	if clickContainer != nil {
		if err := clickContainer.Terminate(ctx); err != nil {
			log.Printf("⚠️  Ошибка остановки ClickHouse: %v", err)
		}
	}

	log.Println("✅ Готово")
	os.Exit(code)
}
