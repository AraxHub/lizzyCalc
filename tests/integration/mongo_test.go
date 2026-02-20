package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"lizzyCalc/internal/domain"
	"lizzyCalc/internal/infrastructure/mongo"
	"lizzyCalc/tests/integration/testutil"
)

// mongoContainer — контейнер MongoDB, инициализируется в TestMain.
var mongoContainer *testutil.MongoContainer

// setupMongoRepo подключается к тестовой MongoDB и очищает коллекцию.
func setupMongoRepo(t *testing.T) *mongo.OperationRepo {
	t.Helper()

	ctx := context.Background()

	client, err := mongo.New(ctx, &mongo.Config{
		URI:        mongoContainer.URI(),
		Database:   "testdb",
		Collection: "operations",
	})
	require.NoError(t, err, "не удалось подключиться к MongoDB")

	// Очищаем коллекцию перед тестом
	err = client.Coll().Drop(ctx)
	if err != nil {
		// Игнорируем ошибку, если коллекции не было
		t.Logf("drop collection: %v (игнорируем)", err)
	}

	t.Cleanup(func() {
		client.Disconnect(context.Background())
	})

	return mongo.NewOperationRepo(client, newTestLogger())
}

// =============================================================================
// Тест MongoDB репозитория
// =============================================================================

func TestMongoRepo_SaveAndGetHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	repo := setupMongoRepo(t)
	ctx := context.Background()

	// Сохраняем операцию
	op := domain.Operation{
		Number1:   10,
		Number2:   5,
		Operation: "+",
		Result:    15,
		Message:   "test",
		Timestamp: time.Now(),
	}

	err := repo.SaveOperation(ctx, op)
	require.NoError(t, err, "SaveOperation должен успешно сохранить")

	// Получаем историю
	history, err := repo.GetHistory(ctx)
	require.NoError(t, err, "GetHistory должен успешно вернуть данные")

	assert.Len(t, history, 1, "должна быть 1 запись")
	assert.Equal(t, 15.0, history[0].Result, "результат должен совпадать")
	assert.Equal(t, "+", history[0].Operation, "операция должна совпадать")
}
