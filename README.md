# Юнит-тесты в Go: подробный гайд

## Что такое юнит-тест

Юнит-тест — это тест **одной изолированной единицы кода**: функции, метода или небольшого модуля. Юнит-тесты проверяют, что конкретная функция при заданных входных данных возвращает ожидаемый результат.

**Ключевые характеристики юнит-тестов:**
- Быстрые (миллисекунды)
- Изолированные (не зависят от БД, сети, файловой системы)
- Детерминированные (всегда один и тот же результат)
- Тестируют одну вещь

## Как Go находит и запускает тесты

### Конвенции именования

Go использует строгие конвенции:

1. **Файлы тестов** — имя заканчивается на `_test.go`:
   ```
   module.go       ← исходный код
   module_test.go  ← тесты для module.go
   ```

2. **Тестовые функции** — имя начинается с `Test` и принимает `*testing.T`:
   ```go
   func TestCacheKey(t *testing.T) { ... }
   func TestCalculate(t *testing.T) { ... }
   ```

3. **Пакет тестов** — тот же пакет, что и тестируемый код:
   ```go
   // module.go
   package calculator
   
   // module_test.go
   package calculator  // тот же пакет — видит приватные функции
   ```

### Почему `_test.go` не компилируется в бинарник

Компилятор Go **игнорирует** файлы `*_test.go` при обычной сборке (`go build`). Они компилируются **только** при запуске `go test`. Это значит:
- Тесты не увеличивают размер production-бинарника
- Тестовые зависимости не попадают в production
- Можно спокойно держать тесты рядом с кодом

### Как работает `go test`

Когда запускаешь `go test ./...`, происходит следующее:

1. Go находит все пакеты в указанном пути
2. Для каждого пакета ищет файлы `*_test.go`
3. Компилирует пакет вместе с тестами во временный бинарник
4. Запускает бинарник
5. Бинарник ищет функции `Test*` и выполняет их
6. Выводит результаты и удаляет временный бинарник

## Запуск тестов

### Базовые команды

```bash
# Запустить все тесты в проекте
go test ./...

# Запустить тесты в конкретном пакете
go test ./internal/usecase/calculator/...

# Запустить с подробным выводом (verbose)
go test ./... -v

# Запустить конкретный тест по имени (регулярное выражение)
go test ./... -run TestCacheKey

# Запустить подтест
go test ./... -run TestCacheKey/сложение_целых

# Запустить тесты с race detector (ищет data races)
go test ./... -race

# Показать покрытие кода
go test ./... -cover

# Сгенерировать HTML-отчёт о покрытии
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Флаги `go test`

| Флаг | Описание |
|------|----------|
| `-v` | Verbose — показывает имя каждого теста |
| `-run <regex>` | Запускает только тесты, соответствующие регулярке |
| `-count N` | Запускает каждый тест N раз (полезно для flaky-тестов) |
| `-timeout 30s` | Таймаут на весь тестовый прогон (по умолчанию 10m) |
| `-short` | Пропускает долгие тесты (если они проверяют `testing.Short()`) |
| `-race` | Включает race detector |
| `-cover` | Показывает процент покрытия |
| `-coverprofile=file` | Записывает данные покрытия в файл |
| `-parallel N` | Максимум N тестов параллельно |

## Структура теста

### Простой тест

```go
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", result)
    }
}
```

### Объект `*testing.T`

`*testing.T` — это объект, через который тест взаимодействует с тестовым фреймворком:

```go
func TestExample(t *testing.T) {
    // Логирование (видно только при -v или при падении теста)
    t.Log("информационное сообщение")
    t.Logf("форматированное: %d", 42)
    
    // Пометить тест как проваленный, но продолжить выполнение
    t.Error("что-то пошло не так")
    t.Errorf("ошибка: %v", err)
    
    // Пометить как проваленный и немедленно остановить
    t.Fatal("критическая ошибка")
    t.Fatalf("критическая: %v", err)
    
    // Пропустить тест
    t.Skip("пропускаем — причина")
    t.Skipf("пропускаем: %s", reason)
    
    // Запустить подтест
    t.Run("subtest name", func(t *testing.T) {
        // ...
    })
    
    // Пометить, что тест можно запускать параллельно
    t.Parallel()
    
    // Получить имя теста
    name := t.Name()
    
    // Создать временную директорию (удалится автоматически)
    dir := t.TempDir()
}
```

## Table-Driven Tests (табличные тесты)

Это **главный паттерн** тестирования в Go. Вместо копипасты создаёшь таблицу тест-кейсов:

```go
func TestCacheKey(t *testing.T) {
    // Таблица тест-кейсов
    tests := []struct {
        name      string  // имя подтеста (для вывода)
        number1   float64 // входные данные
        number2   float64
        operation string
        want      string  // ожидаемый результат
    }{
        {
            name:      "сложение целых",
            number1:   10,
            number2:   5,
            operation: "+",
            want:      "10 + 5",
        },
        {
            name:      "отрицательные числа",
            number1:   -10,
            number2:   -5,
            operation: "+",
            want:      "-10 + -5",
        },
        // ... добавляй сколько угодно кейсов
    }

    // Цикл по всем кейсам
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := cacheKey(tt.number1, tt.number2, tt.operation)
            if got != tt.want {
                t.Errorf("cacheKey(%v, %v, %q) = %q, want %q",
                    tt.number1, tt.number2, tt.operation, got, tt.want)
            }
        })
    }
}
```

### Почему table-driven?

1. **Легко добавлять кейсы** — просто ещё одна строчка в таблице
2. **Нет копипасты** — логика теста в одном месте
3. **Читаемость** — сразу видны все входы и выходы
4. **Подтесты** — `t.Run()` создаёт именованные подтесты, можно запускать по отдельности
5. **Параллельность** — можно добавить `t.Parallel()` в подтест

### Вывод table-driven теста

```
=== RUN   TestCacheKey
=== RUN   TestCacheKey/сложение_целых
=== RUN   TestCacheKey/отрицательные_числа
--- PASS: TestCacheKey (0.00s)
    --- PASS: TestCacheKey/сложение_целых (0.00s)
    --- PASS: TestCacheKey/отрицательные_числа (0.00s)
```

## Тестирование ошибок

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool // ожидаем ошибку?
    }{
        {"нормальное деление", 10, 2, 5, false},
        {"деление на ноль", 10, 0, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            
            // Проверяем, что ошибка есть/нет как ожидалось
            if (err != nil) != tt.wantErr {
                t.Errorf("Divide() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            // Если ошибки не ожидали, проверяем результат
            if !tt.wantErr && got != tt.want {
                t.Errorf("Divide() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Testify — популярная библиотека для ассертов

Стандартная библиотека Go не имеет ассертов — только `t.Error/t.Fatal`. Testify добавляет удобные ассерты.

### Установка

```bash
go get github.com/stretchr/testify
```

### Пакет `assert`

`assert` — при неудаче тест **продолжается** (как `t.Error`):

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestWithAssert(t *testing.T) {
    result := Calculate(10, 5, "+")
    
    // Проверка равенства
    assert.Equal(t, 15.0, result.Result)
    assert.Equal(t, "+", result.Operation)
    
    // Проверка на nil
    assert.Nil(t, err)
    assert.NotNil(t, result)
    
    // Булевы проверки
    assert.True(t, result.Result > 0)
    assert.False(t, result.Result < 0)
    
    // Проверка ошибки
    assert.NoError(t, err)
    assert.Error(t, err) // ожидаем ошибку
    
    // Проверка содержимого строки
    assert.Contains(t, result.Message, "success")
    
    // Проверка типа ошибки
    assert.ErrorIs(t, err, ErrDivisionByZero)
    
    // Сообщение при падении
    assert.Equal(t, 15.0, result.Result, "результат сложения должен быть 15")
}
```

### Пакет `require`

`require` — при неудаче тест **останавливается** (как `t.Fatal`):

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestWithRequire(t *testing.T) {
    result, err := Calculate(10, 5, "+")
    
    // Если err != nil, тест сразу падает — нет смысла проверять result
    require.NoError(t, err)
    require.NotNil(t, result)
    
    // Дальше можно безопасно работать с result
    assert.Equal(t, 15.0, result.Result)
}
```

### Когда что использовать

- **`require`** — для предусловий: если не выполнится, дальнейшие проверки бессмысленны
- **`assert`** — для обычных проверок: хотим увидеть все упавшие ассерты сразу

```go
func TestExample(t *testing.T) {
    // require — если упадёт, дальше проверять нечего
    result, err := DoSomething()
    require.NoError(t, err)
    require.NotNil(t, result)
    
    // assert — проверяем все поля, даже если первое упало
    assert.Equal(t, "expected", result.Field1)
    assert.Equal(t, 42, result.Field2)
    assert.True(t, result.IsValid)
}
```

### Сравнение: стандартный vs testify

**Стандартный Go:**
```go
if result != 15 {
    t.Errorf("got %d, want 15", result)
}
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
```

**С testify:**
```go
assert.Equal(t, 15, result)
require.NoError(t, err)
```

### Полный список ассертов testify

```go
// Равенство
assert.Equal(t, expected, actual)
assert.NotEqual(t, unexpected, actual)
assert.EqualValues(t, expected, actual) // с приведением типов

// Nil
assert.Nil(t, obj)
assert.NotNil(t, obj)

// Булевы
assert.True(t, condition)
assert.False(t, condition)

// Ошибки
assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorIs(t, err, targetErr)
assert.ErrorAs(t, err, &targetType)
assert.ErrorContains(t, err, "substring")

// Строки
assert.Contains(t, str, substring)
assert.NotContains(t, str, substring)
assert.Empty(t, str)
assert.NotEmpty(t, str)
assert.Regexp(t, regexp, str)

// Коллекции
assert.Len(t, slice, expectedLen)
assert.Empty(t, slice)
assert.NotEmpty(t, slice)
assert.Contains(t, slice, element)
assert.ElementsMatch(t, expected, actual) // те же элементы, любой порядок

// Числа
assert.Greater(t, a, b)
assert.GreaterOrEqual(t, a, b)
assert.Less(t, a, b)
assert.LessOrEqual(t, a, b)
assert.InDelta(t, expected, actual, delta) // для float с погрешностью

// Типы
assert.IsType(t, expectedType, actual)
assert.Implements(t, (*Interface)(nil), obj)

// Паника
assert.Panics(t, func() { panicFunc() })
assert.NotPanics(t, func() { safeFunc() })
assert.PanicsWithValue(t, "panic message", func() { ... })
```

## Пример: переписываем тест с testify

**Было (стандартная библиотека):**

```go
func TestCacheKey(t *testing.T) {
    tests := []struct {
        name      string
        number1   float64
        number2   float64
        operation string
        want      string
    }{
        {"сложение", 10, 5, "+", "10 + 5"},
        {"вычитание", 10, 5, "-", "10 - 5"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := cacheKey(tt.number1, tt.number2, tt.operation)
            if got != tt.want {
                t.Errorf("cacheKey(%v, %v, %q) = %q, want %q",
                    tt.number1, tt.number2, tt.operation, got, tt.want)
            }
        })
    }
}
```

**Стало (с testify):**

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCacheKey(t *testing.T) {
    tests := []struct {
        name      string
        number1   float64
        number2   float64
        operation string
        want      string
    }{
        {"сложение", 10, 5, "+", "10 + 5"},
        {"вычитание", 10, 5, "-", "10 - 5"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := cacheKey(tt.number1, tt.number2, tt.operation)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Хелперы: `t.Helper()`

Если пишешь вспомогательную функцию для тестов, добавь `t.Helper()` — тогда при падении Go покажет строку вызова хелпера, а не строку внутри хелпера:

```go
func assertOperation(t *testing.T, op *Operation, expectedResult float64) {
    t.Helper() // без этого ошибка покажет строку ниже, а не место вызова
    
    if op.Result != expectedResult {
        t.Errorf("Result = %v, want %v", op.Result, expectedResult)
    }
}

func TestCalculate(t *testing.T) {
    op := Calculate(10, 5, "+")
    assertOperation(t, op, 15) // ← ошибка покажет эту строку
}
```

## Параллельные тесты

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name string
        // ...
    }{
        {"test1", ...},
        {"test2", ...},
    }

    for _, tt := range tests {
        tt := tt // ВАЖНО: захват переменной для замыкания
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // этот подтест может выполняться параллельно с другими
            // ...
        })
    }
}
```

## Setup и Teardown

### Для одного теста

```go
func TestWithSetup(t *testing.T) {
    // Setup
    tempFile := createTempFile(t)
    
    // Teardown (выполнится в конце теста или при t.Fatal)
    t.Cleanup(func() {
        os.Remove(tempFile)
    })
    
    // Тест
    // ...
}
```

### TestMain — для всего пакета

```go
func TestMain(m *testing.M) {
    // Setup для всех тестов пакета
    setupDB()
    
    // Запуск всех тестов
    code := m.Run()
    
    // Teardown
    teardownDB()
    
    os.Exit(code)
}
```

## Покрытие кода (Coverage)

```bash
# Показать процент покрытия
go test ./... -cover

# Сохранить в файл и сгенерировать HTML
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Показать покрытие по функциям
go tool cover -func=coverage.out
```

**Что значит процент покрытия:**
- 80%+ — хорошо для бизнес-логики
- 100% — не всегда нужно, иногда избыточно
- Покрытие не гарантирует качество тестов — можно иметь 100% покрытия с бесполезными тестами

## Частые ошибки

### 1. Не захватил переменную в параллельном тесте

```go
// НЕПРАВИЛЬНО — все подтесты используют последний tt
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // tt здесь — всегда последний элемент!
    })
}

// ПРАВИЛЬНО
for _, tt := range tests {
    tt := tt // захватываем
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // tt — правильный
    })
}
```

### 2. Используешь `t.Error` вместо `t.Fatal` для критичных проверок

```go
// НЕПРАВИЛЬНО — если result == nil, следующая строка — паника
result, err := DoSomething()
if err != nil {
    t.Errorf("error: %v", err) // тест продолжится!
}
fmt.Println(result.Value) // ПАНИКА если result == nil

// ПРАВИЛЬНО
if err != nil {
    t.Fatalf("error: %v", err) // тест остановится
}
```

### 3. Тестируешь приватные функции через публичный API

Иногда лучше тестировать приватную функцию напрямую (если она в том же пакете), а не через сложный публичный API.

## Структура проекта с тестами

```
internal/
└── usecase/
    └── calculator/
        ├── module.go           ← код
        ├── module_test.go      ← тесты для module.go
        ├── methods.go          ← код
        └── methods_test.go     ← тесты для methods.go
```

**Правило:** тест-файл лежит рядом с исходником и имеет то же имя + `_test.go`.

## Полезные команды

```bash
# Запустить все тесты
make test

# Запустить с verbose
make test-v

# Покрытие
make test-coverage
```
