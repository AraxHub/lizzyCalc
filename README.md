# Тестирование в Go: Моки и Coverage

---

# Часть 1: Моки (Mocks)

## Что такое мок

**Мок** (mock) — это объект-заглушка, который притворяется реальной зависимостью. Мок не выполняет настоящую логику, а возвращает заранее запрограммированные ответы.

### Аналогия

Представь актёра в театре. Ты говоришь ему:
- "Когда тебя спросят 'Сколько будет 10 + 5?' — ответь '15'"

Актёр не умеет считать. Он просто говорит заученную реплику. Так и мок — не настоящий кэш/БД/сервис, а актёр, играющий их роль.

## Зачем нужны моки

### Проблема: зависимости

```go
func (u *UseCase) Calculate(ctx context.Context, n1, n2 float64, op string) (*Operation, error) {
    // Зависимость 1: кэш (Redis)
    if cached, found, _ := u.cache.Get(ctx, key); found {
        return &Operation{Result: cached}, nil
    }
    
    // ... расчёт ...
    
    // Зависимость 2: база данных (PostgreSQL)
    u.repo.SaveOperation(ctx, op)
    
    // Зависимость 3: брокер сообщений (Kafka)
    u.broker.Send(ctx, key, value)
    
    return &op, nil
}
```

Чтобы протестировать эту функцию "по-настоящему", нужно:
1. Запустить Redis
2. Запустить PostgreSQL
3. Запустить Kafka

Это **интеграционный тест** — медленный, сложный в настройке, нестабильный.

### Решение: моки

С моками тестируем **только бизнес-логику**, заменяя зависимости на "актёров":

```go
func TestCalculate(t *testing.T) {
    mockCache := mocks.NewMockICache(ctrl)      // "актёр" вместо Redis
    mockRepo := mocks.NewMockIOperationRepository(ctrl)  // "актёр" вместо PostgreSQL
    mockBroker := mocks.NewMockIProducer(ctrl)  // "актёр" вместо Kafka
    
    // Программируем поведение "актёров"
    mockCache.EXPECT().Get(...).Return(0.0, false, nil)  // "в кэше пусто"
    mockRepo.EXPECT().SaveOperation(...).Return(nil)     // "БД сохранила"
    mockBroker.EXPECT().Send(...).Return(nil)            // "Kafka отправила"
    
    uc := New(mockRepo, mockCache, mockBroker, ...)
    result, err := uc.Calculate(...)
    
    // Проверяем результат
}
```

### Преимущества моков

| Без моков (интеграционные) | С моками (юнит-тесты) |
|---------------------------|----------------------|
| Нужна инфраструктура | Ничего не нужно |
| Медленные (секунды) | Быстрые (миллисекунды) |
| Нестабильные (сеть, диск) | Детерминированные |
| Сложно тестировать ошибки | Легко симулировать любые ошибки |

## Как работают моки

### 1. Интерфейс

Моки работают через интерфейсы. У нас есть:

```go
// internal/ports/cache.go
type ICache interface {
    Get(ctx context.Context, key string) (value float64, found bool, err error)
    Set(ctx context.Context, key string, value float64) error
}
```

### 2. Реальная реализация

```go
// internal/infrastructure/redis/cache.go
type Cache struct {
    client *redis.Client
}

func (c *Cache) Get(ctx context.Context, key string) (float64, bool, error) {
    val, err := c.client.Get(ctx, key).Float64()
    if err == redis.Nil {
        return 0, false, nil
    }
    return val, true, err
}
```

### 3. Мок-реализация (генерируется mockgen)

```go
// internal/mocks/cache_mock.go
type MockICache struct {
    ctrl     *gomock.Controller
    recorder *MockICacheMockRecorder
}

func (m *MockICache) Get(ctx context.Context, key string) (float64, bool, error) {
    ret := m.ctrl.Call(m, "Get", ctx, key)
    return ret[0].(float64), ret[1].(bool), ret[2].(error)
}
```

### 4. UseCase не знает разницы

```go
type UseCase struct {
    cache ports.ICache  // интерфейс — может быть Redis или мок
}
```

## mockgen: генерация моков

### Установка

```bash
go install go.uber.org/mock/mockgen@latest
```

### Добавление директивы в интерфейс

```go
// internal/ports/cache.go
package ports

//go:generate mockgen -source=cache.go -destination=../mocks/cache_mock.go -package=mocks

type ICache interface {
    Get(ctx context.Context, key string) (value float64, found bool, err error)
    Set(ctx context.Context, key string, value float64) error
}
```

### Генерация

```bash
# Сгенерировать все моки
make mocks

# Или вручную
go generate ./internal/ports/...
```

Результат — файлы в `internal/mocks/`:
```
internal/mocks/
├── cache_mock.go
├── repository_mock.go
├── broker_mock.go
└── analytics_mock.go
```

## Использование моков в тестах

### Базовая структура теста

```go
func TestSomething(t *testing.T) {
    // 1. Создаём контроллер
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()  // проверит, что все EXPECT выполнены
    
    // 2. Создаём моки
    mockCache := mocks.NewMockICache(ctrl)
    
    // 3. Программируем поведение
    mockCache.EXPECT().
        Get(gomock.Any(), "key").
        Return(42.0, true, nil)
    
    // 4. Создаём тестируемый объект
    uc := New(nil, mockCache, nil, nil, nil)
    
    // 5. Вызываем метод
    result, err := uc.SomeMethod(...)
    
    // 6. Проверяем результат
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### EXPECT: программирование поведения

```go
// Базовый вызов
mockCache.EXPECT().Get(gomock.Any(), "10 + 5").Return(15.0, true, nil)

// Любые аргументы
mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(0.0, false, nil)

// Конкретное количество вызовов
mockCache.EXPECT().Get(...).Return(...).Times(3)  // ровно 3 раза
mockCache.EXPECT().Get(...).Return(...).AnyTimes()  // 0 или более раз

// Порядок вызовов
gomock.InOrder(
    mockCache.EXPECT().Get(...).Return(0.0, false, nil),   // сначала это
    mockRepo.EXPECT().SaveOperation(...).Return(nil),       // потом это
    mockCache.EXPECT().Set(...).Return(nil),                // потом это
)

// Симуляция ошибки
mockRepo.EXPECT().SaveOperation(...).Return(errors.New("database error"))
```

### Матчеры gomock

```go
gomock.Any()           // любое значение
gomock.Eq("value")     // точное совпадение
gomock.Nil()           // nil
gomock.Not(gomock.Nil()) // не nil
gomock.Len(5)          // слайс/строка длины 5
```

## Пример: полный тест

```go
func TestCalculate_CacheMiss(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockCache := mocks.NewMockICache(ctrl)
    mockRepo := mocks.NewMockIOperationRepository(ctrl)
    mockBroker := mocks.NewMockIProducer(ctrl)
    mockAnalytics := mocks.NewMockIOperationAnalytics(ctrl)

    // Программируем последовательность вызовов
    gomock.InOrder(
        // 1. Проверяем кэш — там пусто
        mockCache.EXPECT().Get(gomock.Any(), "10 + 5").Return(0.0, false, nil),
        // 2. Сохраняем в БД
        mockRepo.EXPECT().SaveOperation(gomock.Any(), gomock.Any()).Return(nil),
        // 3. Кладём в кэш
        mockCache.EXPECT().Set(gomock.Any(), "10 + 5", 15.0).Return(nil),
        // 4. Отправляем в Kafka
        mockBroker.EXPECT().Send(gomock.Any(), []byte("10 + 5"), gomock.Any()).Return(nil),
    )

    uc := New(mockRepo, mockCache, mockBroker, mockAnalytics, newTestLogger())
    result, err := uc.Calculate(context.Background(), 10, 5, "+")

    require.NoError(t, err)
    assert.Equal(t, 15.0, result.Result)
}
```

## Частые команды

```bash
# Установить mockgen
go install go.uber.org/mock/mockgen@latest

# Сгенерировать все моки
make mocks

# Запустить тесты
make test

# Тесты с verbose
make test-v

# Запустить конкретный тест
make test-run NAME=TestCalculate_CacheHit

# Тесты с покрытием
make test-coverage
```

---

# Часть 2: Покрытие кода тестами (Code Coverage)

## Что такое coverage

Coverage (покрытие) — это метрика, показывающая **какой процент кода был выполнен** во время тестов. Go измеряет покрытие по **блокам кода** (statement blocks), а не по сценариям или тест-кейсам.

## Как Go считает покрытие

### Инструментирование кода

При запуске `go test -cover` компилятор вставляет счётчики в каждый **блок кода**. Блок — это последовательность строк без ветвлений (if, switch, for).

```go
func Calculate(a, b float64, op string) (float64, error) {
    // === Блок 1 ===
    if op == "" {
        // === Блок 2 ===
        return 0, errors.New("empty operation")
    }
    
    // === Блок 3 ===
    switch op {
    case "+":
        // === Блок 4 ===
        return a + b, nil
    case "-":
        // === Блок 5 ===
        return a - b, nil
    case "/":
        // === Блок 6 ===
        if b == 0 {
            // === Блок 7 ===
            return 0, errors.New("division by zero")
        }
        // === Блок 8 ===
        return a / b, nil
    default:
        // === Блок 9 ===
        return 0, errors.New("unknown operation")
    }
}
```

### Формула

```
Coverage = (выполненные блоки / всего блоков) × 100%
```

Если функция имеет 9 блоков и тесты прошли через 7 — покрытие **77.8%**.

## Команды для работы с coverage

### Базовые команды

```bash
# Показать процент покрытия
go test ./... -cover

# Вывод:
# ok  	lizzyCalc/internal/usecase/calculator	0.347s	coverage: 97.1% of statements
```

### Детальный отчёт

```bash
# Сохранить данные покрытия в файл
go test ./... -coverprofile=coverage.out

# Открыть HTML-отчёт в браузере
go tool cover -html=coverage.out

# Или сохранить HTML в файл
go tool cover -html=coverage.out -o coverage.html
```

### Покрытие по функциям

```bash
go tool cover -func=coverage.out

# Вывод:
# lizzyCalc/internal/usecase/calculator/methods.go:14:	Calculate		100.0%
# lizzyCalc/internal/usecase/calculator/methods.go:78:	History			100.0%
# lizzyCalc/internal/usecase/calculator/methods.go:83:	HandleOperationEvent	100.0%
# lizzyCalc/internal/usecase/calculator/module.go:11:	cacheKey		100.0%
# lizzyCalc/internal/usecase/calculator/module.go:25:	New			100.0%
# total:							(statements)		97.1%
```

### Makefile команды

```bash
make test           # запустить все тесты
make test-v         # с verbose
make test-coverage  # тесты + HTML-отчёт
```

## Как читать HTML-отчёт

В HTML-отчёте строки подсвечены цветами:

| Цвет | Значение |
|------|----------|
| **Зелёный** | Строка выполнялась во время тестов |
| **Красный** | Строка НЕ выполнялась (не покрыта) |
| **Серый** | Не инструментируется (объявления типов, комментарии, пустые строки) |

Интенсивность зелёного показывает, сколько раз строка выполнялась (светлее = меньше раз).

## Что coverage измеряет и НЕ измеряет

### Измеряет

- Какие строки/блоки кода выполнились
- Какие ветки `if/else` были пройдены
- Какие `case` в `switch` были выполнены
- Какие итерации циклов произошли

### НЕ измеряет

- Качество проверок (assertions)
- Граничные случаи (boundary conditions)
- Комбинации входных данных
- Логическую корректность тестов

## Важно: 100% coverage ≠ хорошие тесты

### Пример 1: Бесполезный тест с 100% покрытием

```go
func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    Add(1, 2)  // 100% coverage, но ничего не проверяет!
}
```

### Пример 2: Пропущенные сценарии

```go
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func TestDivide(t *testing.T) {
    result, _ := Divide(10, 2)
    assert.Equal(t, 5.0, result)
    
    _, err := Divide(10, 0)
    assert.Error(t, err)
}
// 100% coverage, но не проверены:
// - отрицательные числа
// - очень большие числа (переполнение)
// - деление на очень маленькое число
// - NaN, Inf
```

## Рекомендации по покрытию

### Какой процент нужен?

| Тип кода | Рекомендуемое покрытие |
|----------|----------------------|
| Бизнес-логика (use cases) | 80-90%+ |
| Утилиты, хелперы | 70-80% |
| Инфраструктура (DB, HTTP) | 50-70% (часто интеграционные тесты) |
| Сгенерированный код | 0% (исключить из отчёта) |

### Исключение файлов из покрытия

```bash
# Исключить сгенерированные файлы
go test ./... -coverprofile=coverage.out -coverpkg=./... \
    | grep -v "_mock.go" | grep -v ".pb.go"
```

Или в коде:

```go
//go:build !coverage

package mocks
// Этот файл не будет включён при сборке с тегом coverage
```

## Покрытие vs Качество тестов

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   Coverage показывает: "Этот код выполнялся"                │
│   Coverage НЕ показывает: "Этот код работает правильно"     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Пирамида качества тестов

```
        ▲
       /│\        Качество assertions
      / │ \       (что проверяем)
     /  │  \
    /   │   \
   /    │    \    Разнообразие сценариев
  /     │     \   (граничные случаи)
 /      │      \
/───────│───────\ Coverage
        │         (какой код выполнился)
```

Coverage — это **фундамент**, но не гарантия качества.

## Практический пример

### Код

```go
func (u *UseCase) Calculate(ctx context.Context, n1, n2 float64, op string) (*Operation, error) {
    key := cacheKey(n1, n2, op)                          // Блок A
    if cached, found, _ := u.cache.Get(ctx, key); found {
        return &Operation{Result: cached}, nil           // Блок B (cache hit)
    }
    
    var result float64                                   // Блок C
    switch op {
    case "+":
        result = n1 + n2                                 // Блок D
    case "-":
        result = n1 - n2                                 // Блок E
    case "*":
        result = n1 * n2                                 // Блок F
    case "/":
        if n2 == 0 {
            return nil, ErrDivByZero                     // Блок G
        }
        result = n1 / n2                                 // Блок H
    default:
        return nil, ErrUnknownOp                         // Блок I
    }
    
    u.repo.Save(ctx, result)                             // Блок J
    return &Operation{Result: result}, nil
}
```

### Тесты и покрытие

| Тест | Выполненные блоки | Новое покрытие |
|------|------------------|----------------|
| `TestCacheHit` | A, B | 20% |
| `TestAddition` | A, C, D, J | +30% → 50% |
| `TestSubtraction` | A, C, E, J | +10% → 60% |
| `TestMultiplication` | A, C, F, J | +10% → 70% |
| `TestDivision` | A, C, H, J | +10% → 80% |
| `TestDivisionByZero` | A, C, G | +10% → 90% |
| `TestUnknownOp` | A, C, I | +10% → 100% |

## Итог по Coverage

1. **Coverage — полезная метрика**, но не единственная
2. **Красный код** = точно не тестировался
3. **Зелёный код** ≠ хорошо протестирован
4. Стремись к **80%+ для бизнес-логики**
5. Смотри на **качество assertions**, не только на проценты

---

# Часть 3: Шпаргалка по командам

## Тесты

```bash
# Запустить все тесты
make test
go test ./...

# С подробным выводом
make test-v
go test ./... -v

# Конкретный тест
make test-run NAME=TestCalculate
go test ./... -v -run TestCalculate

# Конкретный подтест
go test ./... -v -run TestCacheKey/сложение

# Тесты в конкретном пакете
go test ./internal/usecase/calculator/... -v
```

## Coverage

```bash
# Показать процент
go test ./... -cover

# HTML-отчёт
make test-coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# По функциям
go tool cover -func=coverage.out
```

## Моки

```bash
# Установить mockgen
go install go.uber.org/mock/mockgen@latest

# Сгенерировать моки
make mocks
go generate ./internal/ports/...
```

## Полезные флаги go test

| Флаг | Описание |
|------|----------|
| `-v` | Подробный вывод |
| `-run <regex>` | Фильтр по имени теста |
| `-cover` | Показать покрытие |
| `-coverprofile=file` | Сохранить данные покрытия |
| `-count=N` | Запустить N раз (для flaky-тестов) |
| `-race` | Детектор гонок |
| `-timeout=30s` | Таймаут |
| `-short` | Пропустить долгие тесты |

---

# Итог

## Моки — это

- **Актёры**, играющие роль зависимостей
- Позволяют тестировать **бизнес-логику изолированно**
- Генерируются из **интерфейсов** с помощью mockgen
- Программируются через **EXPECT().Return()**

## Coverage — это

- Метрика того, **какой код выполнился**
- Считается по **блокам кода**, не по сценариям
- **Не гарантирует** качество тестов
- Полезна для поиска **непокрытого кода**

## Workflow

```
1. Пишешь интерфейс (ports/)
2. make mocks — генерируешь моки
3. Пишешь тест с EXPECT()
4. make test — запускаешь
5. make test-coverage — проверяешь покрытие
```
