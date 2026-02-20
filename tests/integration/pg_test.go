package integration

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"lizzyCalc/internal/domain"
	"lizzyCalc/internal/infrastructure/pg"
	"lizzyCalc/tests/integration/testutil"
)

// pgContainer — контейнер PostgreSQL, поднимается один раз для всех тестов пакета.
// Инициализируется в TestMain (main_test.go).
var pgContainer *testutil.PostgresContainer

// newTestLogger создаёт логгер для тестов.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// setupPgDB подключается к тестовой БД и создаёт таблицу operations.
func setupPgDB(t *testing.T) *pg.DB {
	t.Helper()

	// Подключаемся напрямую через database/sql для создания таблицы
	conn, err := sql.Open("postgres", pgContainer.DSN())
	require.NoError(t, err, "не удалось подключиться к PostgreSQL")

	// Создаём таблицу (миграция)
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS operations (
			id SERIAL PRIMARY KEY,
			number1 DOUBLE PRECISION NOT NULL,
			number2 DOUBLE PRECISION NOT NULL,
			operation VARCHAR(10) NOT NULL,
			result DOUBLE PRECISION NOT NULL,
			message TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "не удалось создать таблицу operations")

	// Очищаем таблицу перед каждым тестом
	_, err = conn.Exec("TRUNCATE TABLE operations RESTART IDENTITY")
	require.NoError(t, err, "не удалось очистить таблицу operations")

	conn.Close()

	// Теперь создаём pg.DB через наш модуль
	db, err := pg.New(&pg.Config{
		Host:     pgContainer.Host,
		Port:     pgContainer.Port,
		User:     pgContainer.User,
		Password: pgContainer.Password,
		DBName:   pgContainer.DBName,
		SSLMode:  "disable",
	})
	require.NoError(t, err, "не удалось создать pg.DB")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// =============================================================================
// Тесты PostgreSQL репозитория
// =============================================================================

func TestPgRepo_SaveOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	db := setupPgDB(t)
	repo := pg.NewOperationRepo(db, newTestLogger())
	ctx := context.Background()

	// Создаём операцию
	op := domain.Operation{
		Number1:   10,
		Number2:   5,
		Operation: "+",
		Result:    15,
		Message:   "test",
		Timestamp: time.Now(),
	}

	// Сохраняем
	err := repo.SaveOperation(ctx, op)
	require.NoError(t, err, "SaveOperation должен успешно сохранить")

	// Проверяем напрямую в БД
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM operations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "в таблице должна быть 1 запись")
}

func TestPgRepo_GetHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	db := setupPgDB(t)
	repo := pg.NewOperationRepo(db, newTestLogger())
	ctx := context.Background()

	// Вставляем несколько операций
	ops := []domain.Operation{
		{Number1: 1, Number2: 1, Operation: "+", Result: 2, Timestamp: time.Now().Add(-2 * time.Second)},
		{Number1: 2, Number2: 2, Operation: "+", Result: 4, Timestamp: time.Now().Add(-1 * time.Second)},
		{Number1: 3, Number2: 3, Operation: "+", Result: 6, Timestamp: time.Now()},
	}

	for _, op := range ops {
		err := repo.SaveOperation(ctx, op)
		require.NoError(t, err)
	}

	// Получаем историю
	history, err := repo.GetHistory(ctx)
	require.NoError(t, err, "GetHistory должен успешно вернуть данные")

	// Проверяем
	assert.Len(t, history, 3, "должно быть 3 записи")

	// Проверяем сортировку (последние сначала)
	assert.Equal(t, 6.0, history[0].Result, "первая запись — самая новая")
	assert.Equal(t, 4.0, history[1].Result)
	assert.Equal(t, 2.0, history[2].Result, "последняя запись — самая старая")

	// Проверяем, что ID назначены
	assert.NotZero(t, history[0].ID, "ID должен быть назначен")
}

func TestPgRepo_GetHistory_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	db := setupPgDB(t)
	repo := pg.NewOperationRepo(db, newTestLogger())
	ctx := context.Background()

	// Получаем историю из пустой таблицы
	history, err := repo.GetHistory(ctx)
	require.NoError(t, err, "GetHistory на пустой таблице не должен возвращать ошибку")
	assert.Empty(t, history, "история должна быть пустой")
}

func TestPgRepo_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	db := setupPgDB(t)
	repo := pg.NewOperationRepo(db, newTestLogger())
	ctx := context.Background()

	err := repo.Ping(ctx)
	assert.NoError(t, err, "Ping должен успешно проверить соединение")
}
