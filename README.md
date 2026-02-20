# LizzyCalc — Swagger-документация: полный гайд

## Содержание
1. [Что такое Swagger](#что-такое-swagger)
2. [Как зайти в Swagger UI](#как-зайти-в-swagger-ui)
3. [Как пользоваться Swagger UI](#как-пользоваться-swagger-ui)
4. [Структура папки docs/](#структура-папки-docs)
5. [Как генерируется документация](#как-генерируется-документация)
6. [Аннотации в коде](#аннотации-в-коде)
7. [Как добавить новый эндпоинт](#как-добавить-новый-эндпоинт)
8. [Swagger в Docker и CI/CD](#swagger-в-docker-и-cicd)
9. [Troubleshooting](#troubleshooting)

---

## Что такое Swagger

**Swagger** (он же OpenAPI) — это стандарт описания REST API. Позволяет:
- Автоматически генерировать интерактивную документацию
- Тестировать эндпоинты прямо из браузера
- Генерировать клиенты на разных языках
- Валидировать запросы/ответы

В проекте используется **swag** — генератор Swagger-документации для Go из аннотаций в комментариях.

### Стек
- **swag** (`github.com/swaggo/swag`) — CLI для генерации `docs/`
- **gin-swagger** (`github.com/swaggo/gin-swagger`) — middleware для Gin, отдаёт Swagger UI
- **swaggo/files** — статические файлы Swagger UI

---

## Как зайти в Swagger UI

После запуска сервера открой в браузере:

```
http://localhost:8080/swagger/index.html
```

Или короткий URL (редирект):
```
http://localhost:8080/swagger/
```

### Что увидишь

Интерактивный интерфейс с:
- Списком всех эндпоинтов, сгруппированных по тегам (`calculator`, `system`)
- Описанием каждого метода (Summary, Description)
- Схемами запросов и ответов
- Кнопкой **Try it out** для тестирования

### Пример использования

1. Открой `http://localhost:8080/swagger/index.html`
2. Найди `POST /api/v1/calculate`
3. Нажми на него → раскроется панель
4. Нажми **Try it out**
5. Отредактируй JSON в поле Request body:
   ```json
   {
     "number1": 10,
     "number2": 5,
     "operation": "+"
   }
   ```
6. Нажми **Execute**
7. Увидишь ответ сервера в секции Responses

---

## Как пользоваться Swagger UI

### Интерфейс

```
┌─────────────────────────────────────────────────────────────┐
│  LizzyCalc API  v1.0                                        │
│  API калькулятора с кэшированием...                         │
├─────────────────────────────────────────────────────────────┤
│  ▼ calculator — Операции калькулятора                       │
│    ┌─────────────────────────────────────────────────────┐  │
│    │ POST   /api/v1/calculate   Выполнить вычисление     │  │
│    └─────────────────────────────────────────────────────┘  │
│    ┌─────────────────────────────────────────────────────┐  │
│    │ GET    /api/v1/history     Получить историю         │  │
│    └─────────────────────────────────────────────────────┘  │
│                                                             │
│  ▼ system — Системные эндпоинты                             │
│    ┌─────────────────────────────────────────────────────┐  │
│    │ GET    /liveness           Проверка жизнеспособности│  │
│    └─────────────────────────────────────────────────────┘  │
│    ┌─────────────────────────────────────────────────────┐  │
│    │ GET    /readyness          Проверка готовности      │  │
│    └─────────────────────────────────────────────────────┘  │
│                                                             │
│  Schemas (внизу страницы)                                   │
│    CalculateRequest, CalculateResponse, HistoryItem...      │
└─────────────────────────────────────────────────────────────┘
```

### Try it out

1. Кликни на эндпоинт — раскроется панель
2. Кнопка **Try it out** переводит в режим редактирования
3. Меняешь параметры / body
4. **Execute** — отправляет реальный запрос на сервер
5. Видишь:
   - **Curl** — готовая команда для терминала
   - **Request URL** — полный URL запроса
   - **Response body** — JSON-ответ
   - **Response headers** — заголовки
   - **HTTP status code** — 200, 400, 500...

### Schemas

Внизу страницы секция **Schemas** — описания всех DTO:
- `CalculateRequest` — что принимает `/calculate`
- `CalculateResponse` — что возвращает
- `HistoryItem` — элемент истории
- `ErrorResponse` — формат ошибки

Кликни на схему — увидишь все поля с типами и примерами.

---

## Структура папки docs/

После генерации в `docs/` появляются 3 файла:

```
docs/
├── docs.go        ← Go-код, регистрирует спецификацию в рантайме
├── swagger.json   ← OpenAPI-спецификация в JSON
└── swagger.yaml   ← OpenAPI-спецификация в YAML
```

### docs.go

```go
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": { ... },
    "paths": { ... },
    "definitions": { ... }
}`

var SwaggerInfo = &swag.Spec{
    Version:     "1.0",
    Host:        "localhost:8080",
    Title:       "LizzyCalc API",
    Description: "API калькулятора...",
    ...
}

func init() {
    swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
```

**Что происходит:**
1. `docTemplate` — JSON-шаблон спецификации
2. `SwaggerInfo` — метаданные API
3. `init()` — регистрирует спецификацию при импорте пакета

**Зачем нужен?** Без этого файла gin-swagger не найдёт спецификацию.

### swagger.json / swagger.yaml

Одна и та же OpenAPI-спецификация в двух форматах:
- **JSON** — используется Swagger UI
- **YAML** — удобнее читать человеку

Эти файлы:
- Можно импортировать в Postman
- Можно загрузить на Swagger Hub
- Можно использовать для генерации клиентов (swagger-codegen)

---

## Как генерируется документация

### Команда генерации

```bash
make swagger
```

Под капотом:
```bash
swag init -g cmd/calculator/main.go -o docs --parseDependency --parseInternal
```

### Флаги swag init

| Флаг | Значение |
|------|----------|
| `-g cmd/calculator/main.go` | Точка входа — файл с main() и общими аннотациями |
| `-o docs` | Папка для вывода |
| `--parseDependency` | Парсить зависимости (для внешних типов) |
| `--parseInternal` | Парсить internal/ пакеты |

### Что делает swag

1. **Читает main.go** — ищет общие аннотации `@title`, `@version`, `@host`
2. **Сканирует проект** — ищет все `@Router`, `@Summary`, `@Param`
3. **Парсит DTO** — находит структуры из `@Param` и `@Success`
4. **Генерирует спецификацию** — создаёт `docs.go`, `swagger.json`, `swagger.yaml`

### Установка swag CLI

```bash
make swagger-install
# или
go install github.com/swaggo/swag/cmd/swag@latest
```

После этого `swag` доступен в `$GOPATH/bin/swag` (обычно `~/go/bin/swag`).

---

## Аннотации в коде

### Общие аннотации (main.go)

```go
// @title LizzyCalc API
// @version 1.0
// @description API калькулятора с кэшированием, хранением истории и аналитикой.
// @description Поддерживает операции: сложение (+), вычитание (-), умножение (*), деление (/).

// @host localhost:8080
// @BasePath /

// @tag.name calculator
// @tag.description Операции калькулятора: вычисления и история

// @tag.name system
// @tag.description Системные эндпоинты: liveness, readiness

func main() { ... }
```

| Аннотация | Назначение |
|-----------|------------|
| `@title` | Название API (заголовок в Swagger UI) |
| `@version` | Версия API |
| `@description` | Описание (можно несколько строк) |
| `@host` | Хост для запросов из Swagger UI |
| `@BasePath` | Базовый путь (обычно `/`) |
| `@tag.name` | Имя тега для группировки эндпоинтов |
| `@tag.description` | Описание тега |

### Аннотации эндпоинтов (controller.go)

```go
// @Summary Выполнить вычисление
// @Description Принимает два числа и операцию (+, -, *, /), возвращает результат. Результат кэшируется и сохраняется в БД.
// @Tags calculator
// @Accept json
// @Produce json
// @Param request body CalculateRequest true "Параметры вычисления"
// @Success 200 {object} CalculateResponse "Результат вычисления"
// @Failure 400 {object} CalculateResponse "Невалидный запрос или неизвестная операция"
// @Failure 500 {object} CalculateResponse "Внутренняя ошибка сервера"
// @Router /api/v1/calculate [post]
func (c *Controller) calculate(ctx *gin.Context) { ... }
```

| Аннотация | Назначение |
|-----------|------------|
| `@Summary` | Краткое описание (заголовок в UI) |
| `@Description` | Подробное описание |
| `@Tags` | Теги для группировки (через запятую) |
| `@Accept` | Content-Type запроса (`json`, `xml`, `multipart/form-data`) |
| `@Produce` | Content-Type ответа |
| `@Param` | Параметр (см. ниже) |
| `@Success` | Успешный ответ |
| `@Failure` | Ответ с ошибкой |
| `@Router` | Путь и метод |

### Формат @Param

```
@Param имя источник тип обязательность "описание"
```

Источники:
- `body` — тело запроса (JSON)
- `query` — query-параметр (`?foo=bar`)
- `path` — часть пути (`/users/{id}`)
- `header` — HTTP-заголовок
- `formData` — form-data

Примеры:
```go
// Body (JSON)
// @Param request body CalculateRequest true "Параметры"

// Query
// @Param limit query int false "Лимит записей" default(10)

// Path
// @Param id path int true "ID операции"

// Header
// @Param Authorization header string true "Bearer token"
```

### Аннотации DTO (dto.go)

```go
// CalculateRequest — запрос на вычисление.
// @Description Параметры для выполнения арифметической операции
type CalculateRequest struct {
    Number1   float64 `json:"number1" binding:"required" example:"10.5"`
    Number2   float64 `json:"number2" binding:"required" example:"5.2"`
    Operation string  `json:"operation" binding:"required" example:"+" enums:"+,-,*,/"`
}
```

Теги полей:
| Тег | Назначение |
|-----|------------|
| `json:"name"` | Имя в JSON |
| `example:"10.5"` | Пример значения (показывается в UI) |
| `enums:"a,b,c"` | Допустимые значения (выпадающий список) |
| `default:"10"` | Значение по умолчанию |
| `minimum:"0"` | Минимальное значение |
| `maximum:"100"` | Максимальное значение |
| `minLength:"1"` | Минимальная длина строки |
| `maxLength:"255"` | Максимальная длина строки |
| `format:"email"` | Формат (`email`, `uri`, `uuid`, `date-time`) |

### Комментарий @Description для типа

```go
// @Description Параметры для выполнения арифметической операции
type CalculateRequest struct { ... }
```

Этот комментарий появится в секции Schemas.

---

## Как добавить новый эндпоинт

### Шаг 1: Добавь handler в контроллер

```go
// @Summary Удалить операцию
// @Description Удаляет операцию из истории по ID
// @Tags calculator
// @Produce json
// @Param id path int true "ID операции"
// @Success 200 {object} map[string]string "Операция удалена"
// @Failure 404 {object} ErrorResponse "Операция не найдена"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка"
// @Router /api/v1/operations/{id} [delete]
func (c *Controller) deleteOperation(ctx *gin.Context) {
    // ...
}
```

### Шаг 2: Зарегистрируй маршрут

```go
func (c *Controller) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api/v1")
    api.POST("/calculate", c.calculate)
    api.GET("/history", c.history)
    api.DELETE("/operations/:id", c.deleteOperation) // новый
}
```

### Шаг 3: Перегенерируй документацию

```bash
make swagger
```

### Шаг 4: Проверь в браузере

Перезапусти сервер и открой `http://localhost:8080/swagger/index.html` — новый эндпоинт появится в списке.

---

## Swagger в Docker и CI/CD

### Почему генерация в Dockerfile

Документация генерируется при сборке образа:

```dockerfile
# Dockerfile (если используется)
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/calculator/main.go -o docs --parseDependency --parseInternal
```

Это гарантирует, что:
- В контейнере актуальная документация
- Не нужно коммитить `docs/` (можно добавить в `.gitignore`)

### Альтернатива: коммитить docs/

В этом проекте `docs/` коммитится. Плюсы:
- Не нужен `swag` при сборке образа
- Быстрее сборка
- Видно diff документации в PR

Минусы:
- Нужно не забывать `make swagger` перед коммитом
- Возможно рассинхронизация кода и документации

### CI-проверка

Добавь в CI/CD:
```bash
# Проверка, что документация актуальна
make swagger
git diff --exit-code docs/
```

Если документация устарела — CI упадёт.

---

## Troubleshooting

### Swagger UI не открывается (404)

**Проверь:**
1. Импорт docs в main.go:
   ```go
   import _ "lizzyCalc/docs"
   ```
   Без этого спецификация не регистрируется.

2. Маршрут в server.go:
   ```go
   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   ```

3. Сервер запущен на правильном порту:
   ```bash
   curl http://localhost:8080/swagger/index.html
   ```

### Эндпоинт не появляется в Swagger

**Проверь:**
1. Есть ли `@Router` аннотация
2. Формат: `// @Router /path [method]` (метод в квадратных скобках)
3. Перегенерировал ли `make swagger`
4. Перезапустил ли сервер

### Схема DTO пустая

**Проверь:**
1. Структура экспортируемая (имя с большой буквы)
2. Поля экспортируемые
3. Есть `json` теги
4. Флаг `--parseInternal` если структура в internal/

### "cannot find type definition"

```
[parser] cannot find type definition: SomeType
```

**Решение:** добавь `--parseDependency`:
```bash
swag init -g cmd/calculator/main.go -o docs --parseDependency --parseInternal
```

### Кириллица крякозябрами

Swagger использует UTF-8. Если проблемы:
1. Проверь `Content-Type: application/json; charset=utf-8` в ответах
2. Файлы `docs/*` должны быть в UTF-8

### Swagger UI показывает старую версию

```bash
# Очистить кэш браузера или
curl http://localhost:8080/swagger/doc.json
```

Если JSON новый, а UI старый — жёсткий refresh (Cmd+Shift+R).

---

## Дополнительные ссылки

- [swag GitHub](https://github.com/swaggo/swag)
- [gin-swagger GitHub](https://github.com/swaggo/gin-swagger)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger Editor](https://editor.swagger.io/) — онлайн-редактор спецификаций
