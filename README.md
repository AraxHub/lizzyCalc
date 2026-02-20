# Интеграционные тесты с testcontainers

## Что это

**Интеграционные тесты** проверяют реальное взаимодействие кода с инфраструктурой (БД, кэш). В отличие от юнит-тестов с моками, здесь поднимаются **настоящие** PostgreSQL, Redis, MongoDB, ClickHouse в Docker-контейнерах.

**testcontainers-go** — библиотека, которая:
1. Запускает Docker-контейнер перед тестом
2. Ждёт, пока сервис станет готов
3. Отдаёт параметры подключения (host, port)
4. Останавливает контейнер после теста

## Быстрый старт

```bash
# Запустить интеграционные тесты (требуется Docker)
make test-integration

# Или напрямую
go test ./tests/integration/... -v
```

---

## Структура проекта

```
tests/
└── integration/
    ├── main_test.go           ← TestMain: setup/teardown контейнеров
    ├── pg_test.go             ← тесты PostgreSQL репозитория
    ├── redis_test.go          ← тесты Redis кэша
    ├── mongo_test.go          ← тесты MongoDB репозитория
    ├── clickhouse_test.go     ← тесты ClickHouse writer
    └── testutil/
        └── containers.go      ← хелперы для создания контейнеров
```

---

## Как работает testcontainers

### 1. Создание контейнера

```go
// testutil/containers.go

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
    // Запускаем контейнер PostgreSQL
    container, err := postgres.Run(ctx,
        "postgres:16-alpine",                    // образ
        postgres.WithDatabase("testdb"),         // имя БД
        postgres.WithUsername("test"),           // пользователь
        postgres.WithPassword("test"),           // пароль
        testcontainers.WithWaitStrategy(         // ждём готовности
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Получаем параметры подключения
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    return &PostgresContainer{
        PostgresContainer: container,
        Host:              host,
        Port:              port.Port(),  // динамический порт!
        // ...
    }, nil
}
```

### 2. Lifecycle: один контейнер на весь пакет

```go
// main_test.go

var pgContainer *testutil.PostgresContainer
var redisContainer *testutil.RedisContainer
var mongoContainer *testutil.MongoContainer
var clickContainer *testutil.ClickHouseContainer

func TestMain(m *testing.M) {
    ctx := context.Background()

    // === SETUP: поднимаем контейнеры один раз ===
    pgContainer, _ = testutil.NewPostgresContainer(ctx)
    redisContainer, _ = testutil.NewRedisContainer(ctx)
    mongoContainer, _ = testutil.NewMongoContainer(ctx)
    clickContainer, _ = testutil.NewClickHouseContainer(ctx)

    // === ЗАПУСК ТЕСТОВ ===
    code := m.Run()

    // === TEARDOWN: останавливаем контейнеры ===
    pgContainer.Terminate(ctx)
    redisContainer.Terminate(ctx)
    mongoContainer.Terminate(ctx)
    clickContainer.Terminate(ctx)

    os.Exit(code)
}
```

**Почему один контейнер на пакет?**
- Поднятие контейнера занимает 2-5 секунд
- Если поднимать на каждый тест — будет очень медленно
- Вместо этого **очищаем данные** перед каждым тестом

### 3. Очистка данных перед тестом

```go
// pg_test.go

func setupPgDB(t *testing.T) *pg.DB {
    // Подключаемся
    conn, _ := sql.Open("postgres", pgContainer.DSN())

    // Создаём таблицу (миграция)
    conn.Exec(`CREATE TABLE IF NOT EXISTS operations (...)`)

    // ОЧИЩАЕМ таблицу перед каждым тестом
    conn.Exec("TRUNCATE TABLE operations RESTART IDENTITY")

    // ...
}
```

```go
// redis_test.go

func setupRedisCache(t *testing.T) *redis.Cache {
    client, _ := redis.New(...)

    // ОЧИЩАЕМ Redis перед каждым тестом
    client.FlushDB(context.Background())

    // ...
}
```

---

## Какие контейнеры поднимаются

| Сервис | Образ | Порт | Wait Strategy |
|--------|-------|------|---------------|
| PostgreSQL | `postgres:16-alpine` | 5432 | Log: "ready to accept connections" (2x) |
| Redis | `redis:7-alpine` | 6379 | Log: "Ready to accept connections" |
| MongoDB | `mongo:7` | 27017 | Log: "Waiting for connections" |
| ClickHouse | `clickhouse/clickhouse-server:24-alpine` | 9000 | HTTP: `/` на порту 8123 |

---

## Что тестируется

### PostgreSQL (4 теста)

```go
func TestPgRepo_SaveOperation(t *testing.T)     // INSERT работает
func TestPgRepo_GetHistory(t *testing.T)        // SELECT + сортировка DESC
func TestPgRepo_GetHistory_Empty(t *testing.T)  // Пустая таблица → пустой слайс
func TestPgRepo_Ping(t *testing.T)              // Соединение живое
```

**Пример теста:**

```go
func TestPgRepo_SaveOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("пропускаем интеграционный тест")
    }

    db := setupPgDB(t)  // подключение + очистка
    repo := pg.NewOperationRepo(db, logger)

    op := domain.Operation{
        Number1:   10,
        Number2:   5,
        Operation: "+",
        Result:    15,
        Timestamp: time.Now(),
    }

    // Сохраняем
    err := repo.SaveOperation(ctx, op)
    require.NoError(t, err)

    // Проверяем напрямую в БД
    var count int
    db.QueryRow("SELECT COUNT(*) FROM operations").Scan(&count)
    assert.Equal(t, 1, count)
}
```

### Redis (4 теста)

```go
func TestRedisCache_SetAndGet(t *testing.T)      // Set + Get работают
func TestRedisCache_Get_NotFound(t *testing.T)   // Несуществующий ключ → found=false
func TestRedisCache_Overwrite(t *testing.T)      // Перезапись значения
func TestRedisCache_FloatPrecision(t *testing.T) // Точность float64
```

**Пример теста:**

```go
func TestRedisCache_SetAndGet(t *testing.T) {
    cache := setupRedisCache(t)  // подключение + FlushDB

    // Сохраняем
    err := cache.Set(ctx, "10 + 5", 15.0)
    require.NoError(t, err)

    // Получаем
    value, found, err := cache.Get(ctx, "10 + 5")
    require.NoError(t, err)
    assert.True(t, found)
    assert.Equal(t, 15.0, value)
}
```

### MongoDB (1 тест)

```go
func TestMongoRepo_SaveAndGetHistory(t *testing.T)  // InsertOne + Find
```

### ClickHouse (1 тест)

```go
func TestClickWriter_WriteOperation(t *testing.T)   // EnsureTable + INSERT
```

---

## От чего защищают интеграционные тесты

### Почему моки НЕ защитят от ошибок в SQL

```go
// crud.go — опечатка в SQL
func (r *OperationRepo) SaveOperation(ctx context.Context, op domain.Operation) error {
    _, err := r.db.ExecContext(ctx,
        `INSER INTO operations ...`,  // ← ОПЕЧАТКА!
        op.Number1, ...)
    return err
}
```

```go
// Юнит-тест с моком — ПРОЙДЁТ, хотя SQL сломан!
func TestCalculate(t *testing.T) {
    mockRepo := mocks.NewMockIOperationRepository(ctrl)
    mockRepo.EXPECT().SaveOperation(gomock.Any(), gomock.Any()).Return(nil)  // ← мок просто вернёт nil
    
    uc := New(mockRepo, ...)
    result, err := uc.Calculate(ctx, 10, 5, "+")
    
    assert.NoError(t, err)  // ✅ Тест пройдёт!
}
```

**Мок не выполняет реальный SQL** — он возвращает то, что запрограммировано в `Return()`. Опечатка в SQL обнаружится только в production.

**Интеграционный тест поймает:**

```go
func TestPgRepo_SaveOperation(t *testing.T) {
    repo := pg.NewOperationRepo(realDB, logger)
    
    err := repo.SaveOperation(ctx, op)
    
    // ❌ FAIL: pq: syntax error at or near "INSER"
    require.NoError(t, err)
}
```

### Защищают

| Проблема | Пример |
|----------|--------|
| Ошибки в SQL | `INSER INTO` вместо `INSERT INTO` |
| Неправильные типы | `VARCHAR` вместо `DOUBLE PRECISION` |
| Проблемы сериализации | `float64` → `string` → `float64` теряет точность |
| Неправильная схема | Забыли добавить колонку |
| Ошибки сортировки | `ORDER BY created_at ASC` вместо `DESC` |
| NULL handling | `Scan` в `nil` поле |

### НЕ защищают

| Проблема | Почему |
|----------|--------|
| Производительность под нагрузкой | Тестируем с 1-10 записями |
| Конкурентный доступ | Тесты последовательные |
| Сетевые проблемы в production | Контейнер локальный |
| Большие объёмы данных | Нет нагрузочного тестирования |
| Проблемы с правами доступа | Тестовый пользователь — суперадмин |

---

## Как юнит и интеграционные тесты дополняют друг друга

```
┌─────────────────────────────────────────────────────────────┐
│                      ЮНИТ-ТЕСТЫ (моки)                      │
│                                                             │
│  TestCalculate_CacheHit:                                    │
│    mockCache.EXPECT().Get(...).Return(15.0, true, nil)      │
│    → Проверяем: "при cache hit БД не вызывается"            │
│    → НЕ проверяем: реально ли Redis работает                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              ↓
                    Мок гарантирует, что
                    логика вызывает cache.Get()
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                 ИНТЕГРАЦИОННЫЕ ТЕСТЫ (Docker)               │
│                                                             │
│  TestRedisCache_SetAndGet:                                  │
│    cache.Set(ctx, "10 + 5", 15.0)                           │
│    value, found, _ := cache.Get(ctx, "10 + 5")              │
│    → Проверяем: Redis реально сохраняет и возвращает        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Юнит-тест:** "UseCase вызывает `cache.Get()` в нужный момент"
**Интеграционный:** "`cache.Get()` реально получает данные из Redis"

---

## testutil/containers.go — API хелперов

### PostgresContainer

```go
type PostgresContainer struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
}

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error)
func (c *PostgresContainer) DSN() string  // connection string для lib/pq
func (c *PostgresContainer) Terminate(ctx context.Context) error
```

### RedisContainer

```go
type RedisContainer struct {
    Host string
    Port string
}

func NewRedisContainer(ctx context.Context) (*RedisContainer, error)
func (c *RedisContainer) Addr() string  // "host:port"
func (c *RedisContainer) Terminate(ctx context.Context) error
```

### MongoContainer

```go
type MongoContainer struct {
    Host string
    Port string
}

func NewMongoContainer(ctx context.Context) (*MongoContainer, error)
func (c *MongoContainer) URI() string  // "mongodb://host:port"
func (c *MongoContainer) Terminate(ctx context.Context) error
```

### ClickHouseContainer

```go
type ClickHouseContainer struct {
    Host     string
    Port     string
    User     string
    Password string
    Database string
}

func NewClickHouseContainer(ctx context.Context) (*ClickHouseContainer, error)
func (c *ClickHouseContainer) Terminate(ctx context.Context) error
```

---

## Пропуск интеграционных тестов

В каждом тесте есть проверка:

```go
func TestPgRepo_SaveOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("пропускаем интеграционный тест в short режиме")
    }
    // ...
}
```

Это позволяет:

```bash
# Только юнит-тесты (быстро, без Docker)
go test ./... -short
make test-unit

# Все тесты включая интеграционные
go test ./...
make test
```

---

## Makefile команды

```bash
make test              # все тесты (юнит + интеграционные)
make test-unit         # только юнит-тесты (-short)
make test-integration  # только интеграционные
make test-v            # все с verbose
```

---

## Время выполнения

```
🚀 Поднимаем тестовые контейнеры...
✅ PostgreSQL: localhost:55031     (~2 сек)
✅ Redis: localhost:55032          (~1 сек)
✅ MongoDB: localhost:55033        (~2 сек)
✅ ClickHouse: localhost:55035     (~5 сек)
🧪 Запускаем тесты...
   10 тестов                       (~0.5 сек)
🧹 Останавливаем контейнеры...     (~3 сек)
────────────────────────────────────────────
ИТОГО:                             ~13-15 сек
```

---

## Типичные проблемы

### Docker не запущен

```
Cannot connect to the Docker daemon
```

**Решение:** Запустить Docker Desktop

### Порт занят

```
bind: address already in use
```

**Решение:** testcontainers использует динамические порты, эта ошибка редка. Если возникла — перезапустить Docker.

### Таймаут при старте контейнера

```
context deadline exceeded
```

**Решение:** Увеличить `WithStartupTimeout()` в `containers.go`

### Контейнер не останавливается

testcontainers использует **Ryuk** — sidecar-контейнер, который автоматически убивает "осиротевшие" контейнеры. Даже если тест упал — контейнеры будут остановлены.

---

## E2E-тесты (end-to-end)

### Что это

E2E-тест проверяет **весь путь запроса** — от HTTP/gRPC до БД и обратно. Поднимается вся инфраструктура + само приложение, и делаются реальные запросы как от клиента.

### Как выглядит

```go
func TestE2E_Calculate(t *testing.T) {
    // 1. Поднимаем ВСЁ: PG, Redis, Kafka, ClickHouse + само приложение
    containers := setupAllContainers(t)
    app := startApp(t, containers)  // запускаем main() с тестовым конфигом
    
    // 2. Делаем реальный HTTP-запрос
    resp, err := http.Post(
        "http://localhost:8080/api/calculate",
        "application/json",
        strings.NewReader(`{"number1": 10, "number2": 5, "operation": "+"}`),
    )
    
    // 3. Проверяем ответ
    assert.Equal(t, 200, resp.StatusCode)
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, 15.0, result["result"])
    
    // 4. Проверяем, что данные реально сохранились в БД
    var count int
    containers.PgDB.QueryRow("SELECT COUNT(*) FROM operations").Scan(&count)
    assert.Equal(t, 1, count)
}
```

### Что тестируется в E2E

| Слой | Что проверяем |
|------|---------------|
| HTTP/gRPC | Роутинг, middleware, сериализация |
| Контроллеры | Валидация, маппинг |
| UseCase | Бизнес-логика |
| Репозитории | SQL, кэш |
| Интеграция | Всё работает вместе |

### Варианты реализации

**Вариант 1: Приложение в том же процессе**
```go
app := app.New(cfg)
go app.Run()
// делаем запросы к localhost:8080
```

**Вариант 2: Приложение в Docker-контейнере**
```go
appContainer := testcontainers.Run("lizzycalc:test", ...)
// делаем запросы к appContainer.Host():appContainer.Port()
```

**Вариант 3: httptest.Server**
```go
router := http.NewRouter(uc)
server := httptest.NewServer(router)
// делаем запросы к server.URL
```

### Сложности E2E

| Проблема | Описание |
|----------|----------|
| **Долго** | Поднять всю инфраструктуру + приложение |
| **Flaky** | Много точек отказа (сеть, таймауты) |
| **Сложно дебажить** | Где именно сломалось? |
| **Асинхронность** | Kafka consumer — нужно ждать |

### Когда нужны E2E

- Критичные бизнес-флоу (оплата, регистрация)
- Проверка middleware (auth, rate limiting)
- Smoke-тесты перед релизом
- Когда юнит + интеграционные не дают уверенности

---

## Smoke-тесты

### Что это

**Smoke-тест** — быстрая поверхностная проверка, что система вообще работает после деплоя. Название от "smoke test" в электронике: включаешь устройство — если дым не пошёл, значит базово работает.

### Примеры

```bash
# После деплоя проверяем:
curl http://api.example.com/health                    # 200 OK?
curl http://api.example.com/api/calculate \
  -d '{"number1":1,"number2":1,"operation":"+"}'      # возвращает 2?
```

```go
func TestSmoke_HealthEndpoint(t *testing.T) {
    resp, err := http.Get(baseURL + "/health")
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}

func TestSmoke_CalculateWorks(t *testing.T) {
    resp, _ := http.Post(baseURL + "/api/calculate", 
        "application/json",
        strings.NewReader(`{"number1":1,"number2":1,"operation":"+"}`))
    assert.Equal(t, 200, resp.StatusCode)
    // НЕ проверяем все кейсы — только что отвечает
}
```

### Smoke vs E2E vs Интеграционные

| Тип | Цель | Глубина | Когда запускать |
|-----|------|---------|-----------------|
| **Smoke** | "Приложение живое?" | Поверхностно | После каждого деплоя |
| **E2E** | "Флоу работает полностью?" | Глубоко | Перед релизом |
| **Интеграционные** | "Компонент работает с БД?" | Средне | На CI при каждом коммите |

---

## Пирамида тестов

```
            ▲
           /│\        E2E / Smoke
          / │ \       Медленно, дорого, ловит баги интеграции
         /  │  \
        /───┼───\     Интеграционные
       /    │    \    Средне, реальные БД в Docker
      /     │     \
     /──────┼──────\  Юнит-тесты
    /       │       \ Быстро, дёшево, моки
```

| Уровень | Скорость | Стоимость | Что ловит |
|---------|----------|-----------|-----------|
| Юнит | Мс | Дёшево | Баги логики |
| Интеграционные | Сек | Средне | Баги SQL, сериализации |
| E2E | Мин | Дорого | Баги интеграции слоёв |

**Правило:** больше тестов внизу пирамиды, меньше — наверху.

---

## Итог

| Характеристика | Значение |
|----------------|----------|
| **Что тестируем** | Инфраструктурный код (репозитории, кэш) |
| **Как** | Docker-контейнеры через testcontainers |
| **Скорость** | ~15 секунд на 10 тестов |
| **Зависимости** | Docker |
| **Изоляция** | Контейнер на пакет, очистка данных перед тестом |
| **Защищает от** | Ошибок в SQL, сериализации, схемах |
| **Не защищает от** | Нагрузки, конкурентности, сетевых проблем |

### Что реализовано в проекте

| Уровень | Статус | Файлы |
|---------|--------|-------|
| Юнит-тесты с моками | ✅ | `internal/usecase/calculator/*_test.go` |
| Интеграционные | ✅ | `tests/integration/*.go` |
| E2E | ❌ (будущее) | — |
| Smoke | ❌ (будущее) | — |
