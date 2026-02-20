package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"lizzyCalc/internal/infrastructure/redis"
	"lizzyCalc/tests/integration/testutil"
)

// redisContainer — контейнер Redis, поднимается один раз для всех тестов пакета.
// Инициализируется в TestMain (main_test.go).
var redisContainer *testutil.RedisContainer

// setupRedisCache подключается к тестовому Redis и очищает его.
func setupRedisCache(t *testing.T) *redis.Cache {
	t.Helper()

	client, err := redis.New(&redis.Config{
		Host:     redisContainer.Host,
		Port:     redisContainer.Port,
		Password: "",
		DB:       0,
	})
	require.NoError(t, err, "не удалось подключиться к Redis")

	// Очищаем Redis перед каждым тестом
	err = client.FlushDB(context.Background()).Err()
	require.NoError(t, err, "не удалось очистить Redis")

	t.Cleanup(func() {
		client.Close()
	})

	return redis.NewCache(client, newTestLogger())
}

// =============================================================================
// Тесты Redis кэша
// =============================================================================

func TestRedisCache_SetAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	cache := setupRedisCache(t)
	ctx := context.Background()

	// Сохраняем значение
	err := cache.Set(ctx, "10 + 5", 15.0)
	require.NoError(t, err, "Set должен успешно сохранить")

	// Получаем значение
	value, found, err := cache.Get(ctx, "10 + 5")
	require.NoError(t, err, "Get должен успешно получить")
	assert.True(t, found, "ключ должен быть найден")
	assert.Equal(t, 15.0, value, "значение должно совпадать")
}

func TestRedisCache_Get_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	cache := setupRedisCache(t)
	ctx := context.Background()

	// Пытаемся получить несуществующий ключ
	value, found, err := cache.Get(ctx, "несуществующий_ключ")

	require.NoError(t, err, "Get несуществующего ключа не должен возвращать ошибку")
	assert.False(t, found, "ключ не должен быть найден")
	assert.Equal(t, 0.0, value, "значение должно быть нулевым")
}

func TestRedisCache_Overwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	cache := setupRedisCache(t)
	ctx := context.Background()

	// Сохраняем первое значение
	err := cache.Set(ctx, "key", 100.0)
	require.NoError(t, err)

	// Перезаписываем
	err = cache.Set(ctx, "key", 200.0)
	require.NoError(t, err)

	// Проверяем, что значение обновилось
	value, found, err := cache.Get(ctx, "key")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 200.0, value, "значение должно быть перезаписано")
}

func TestRedisCache_FloatPrecision(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в short режиме")
	}

	cache := setupRedisCache(t)
	ctx := context.Background()

	// Проверяем точность float64
	testCases := []float64{
		0.1 + 0.2,        // классическая проблема float
		3.14159265358979, // число pi
		1e-10,            // очень маленькое
		1e10,             // очень большое
		-42.5,            // отрицательное
	}

	for _, expected := range testCases {
		key := "precision_test"
		err := cache.Set(ctx, key, expected)
		require.NoError(t, err)

		value, found, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, expected, value, "значение %v должно сохраняться точно", expected)
	}
}
