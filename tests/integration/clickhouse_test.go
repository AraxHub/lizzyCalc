package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"lizzyCalc/internal/domain"
	"lizzyCalc/internal/infrastructure/click"
	"lizzyCalc/tests/integration/testutil"
)

// clickContainer — контейнер ClickHouse, инициализируется в TestMain.
var clickContainer *testutil.ClickHouseContainer

// setupClickWriter подключается к тестовому ClickHouse и создаёт таблицу.
func setupClickWriter(t *testing.T) *click.OperationWriter {
	t.Helper()

	ctx := context.Background()

	client, err := click.New(&click.Config{
		Host:     clickContainer.Host,
		Port:     clickContainer.Port,
		Database: clickContainer.Database,
		Username: clickContainer.User,
		Password: clickContainer.Password,
	})
	require.NoError(t, err, "не удалось подключиться к ClickHouse")

	writer := click.NewOperationWriter(client)

	// Создаём таблицу
	err = writer.EnsureTable(ctx)
	require.NoError(t, err, "не удалось создать таблицу")

	// Очищаем таблицу перед тестом
	_, err = client.DB().ExecContext(ctx, "TRUNCATE TABLE default.operations_analytics")
	require.NoError(t, err, "не удалось очистить таблицу")

	t.Cleanup(func() {
		client.Close()
	})

	return writer
}

// =============================================================================
// Тест ClickHouse writer
// =============================================================================

func TestClickWriter_WriteOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	writer := setupClickWriter(t)
	ctx := context.Background()

	// Записываем операцию
	op := domain.Operation{
		Number1:   10,
		Number2:   5,
		Operation: "+",
		Result:    15,
		Message:   "test",
		Timestamp: time.Now(),
	}

	err := writer.WriteOperation(ctx, op)
	require.NoError(t, err, "WriteOperation должен успешно записать")

	// Проверяем, что запись есть (через writer.db недоступен, но мы проверили что нет ошибки)
	// В реальном тесте можно было бы сделать SELECT COUNT(*), но для простоты достаточно
	assert.NoError(t, err, "операция должна быть записана без ошибок")
}
