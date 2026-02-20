# MongoDB в lizzyCalc

## Что такое MongoDB

**MongoDB** — документо-ориентированная NoSQL база данных. В отличие от реляционных БД (PostgreSQL, MySQL), где данные хранятся в таблицах со строгой схемой (строки и колонки), MongoDB хранит данные в виде **документов** — JSON-подобных объектов (формат BSON — Binary JSON).

### Ключевые отличия от PostgreSQL

| Аспект | PostgreSQL (реляционная) | MongoDB (документная) |
|--------|--------------------------|----------------------|
| Единица хранения | Строка в таблице | Документ в коллекции |
| Схема | Строгая: колонки, типы, constraints | Гибкая: документы могут иметь разные поля |
| Связи | JOIN между таблицами | Вложенные документы или ссылки |
| Язык запросов | SQL | MongoDB Query Language (MQL) |
| Транзакции | ACID из коробки | ACID с версии 4.0, но реже используется |
| Масштабирование | Вертикальное (мощнее сервер) | Горизонтальное (шардирование) |

### Когда MongoDB лучше

- **Гибкая схема**: структура данных часто меняется, разные документы имеют разные поля.
- **Иерархические данные**: вложенные объекты, массивы — всё в одном документе.
- **Быстрое прототипирование**: не нужно писать миграции при изменении структуры.
- **Горизонтальное масштабирование**: шардирование «из коробки».
- **Большие объёмы неструктурированных данных**: логи, события, каталоги товаров.

### Когда PostgreSQL лучше

- **Строгая консистентность**: финансы, заказы, где важна целостность.
- **Сложные связи**: много JOIN'ов, нормализованная схема.
- **Аналитика по структурированным данным**: SQL + индексы + оконные функции.

---

## Основные понятия MongoDB

### База данных (Database)

Контейнер для коллекций. Одна MongoDB может содержать много баз. В проекте база называется **`lizzycalc`** (переменная `CALCULATOR_MONGO_DATABASE`).

### Коллекция (Collection)

Аналог таблицы в SQL. Коллекция — это набор документов. В отличие от таблицы, коллекция **не требует схемы**: документы в одной коллекции могут иметь разные поля.

В проекте коллекция называется **`operations`** (переменная `CALCULATOR_MONGO_COLLECTION`).

Пример: коллекция `operations` хранит документы с операциями калькулятора.

### Документ (Document)

Единица данных в MongoDB. Документ — это JSON-объект (хранится как BSON). У каждого документа есть уникальный идентификатор **`_id`** (генерируется автоматически, тип ObjectId).

Пример документа в коллекции `operations`:

```json
{
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "number1": 10.5,
  "number2": 3.0,
  "operation": "+",
  "result": 13.5,
  "message": "",
  "created_at": ISODate("2024-01-15T10:30:00Z")
}
```

### Поле (Field)

Пара «ключ-значение» в документе. Поля могут быть:
- Примитивами: строка, число, boolean, дата.
- Массивами: `"tags": ["math", "calc"]`.
- Вложенными документами: `"user": { "name": "John", "age": 30 }`.

### Индекс (Index)

Структура для ускорения поиска. Без индекса MongoDB сканирует все документы (collection scan). С индексом — быстрый поиск по ключу.

В проекте создаётся индекс по полю `created_at` (метод `EnsureIndexes`) для быстрой сортировки истории операций.

---

## Как MongoDB устроена в lizzyCalc

### Архитектура

```
┌─────────────────────────────────────────────────────────┐
│                        App                              │
│  ┌─────────────┐    ┌─────────────────────────────┐     │
│  │  Use Case   │───▶│  ports.IOperationRepository │     │
│  └─────────────┘    └─────────────────────────────┘     │
│                              │                          │
│         ┌────────────────────┼────────────────────┐     │
│         ▼                    ▼                    ▼     │
│  ┌─────────────┐      ┌─────────────┐      ┌──────────┐ │
│  │  pg.Repo    │      │ mongo.Repo  │      │  (...)   │ │
│  │ (PostgreSQL)│      │ (MongoDB)   │      │          │ │
│  └─────────────┘      └─────────────┘      └──────────┘ │
└─────────────────────────────────────────────────────────┘
```

**Use case** работает с интерфейсом `IOperationRepository`. Конкретная реализация (PostgreSQL или MongoDB) выбирается при старте приложения по фича-флагу `CALCULATOR_FEATURE_PG`.

### Файловая структура

```
internal/infrastructure/mongo/
├── module.go      — Config, Client, New(), EnsureIndexes()
└── repository.go  — OperationRepo: SaveOperation, GetHistory, Ping
```

### Конфиг (module.go)

```go
type Config struct {
    URI        string `envconfig:"URI" default:"mongodb://localhost:27017"`
    Database   string `envconfig:"DATABASE" default:"lizzycalc"`
    Collection string `envconfig:"COLLECTION" default:"operations"`
}
```

Переменные окружения:
- `CALCULATOR_MONGO_URI` — строка подключения (например `mongodb://localhost:27017` или `mongodb://user:pass@host:27017`).
- `CALCULATOR_MONGO_DATABASE` — имя базы.
- `CALCULATOR_MONGO_COLLECTION` — имя коллекции.

### Подключение (Client)

```go
func New(ctx context.Context, cfg *Config) (*Client, error)
```

1. Вызывает `mongo.Connect(options.Client().ApplyURI(cfg.URI))`.
2. Проверяет соединение: `client.Ping(ctx, nil)`.
3. Возвращает обёртку `*Client` с методами `DB()`, `Coll()`.

### Создание индексов (EnsureIndexes)

```go
func (c *Client) EnsureIndexes(ctx context.Context) error
```

Создаёт индекс по `created_at` (убывающий) для быстрой сортировки в `GetHistory`. Вызывается один раз при старте приложения (аналог миграций, но в MongoDB коллекция создаётся автоматически при первой вставке).

### Репозиторий (repository.go)

Реализует интерфейс `ports.IOperationRepository`:

```go
type IOperationRepository interface {
    SaveOperation(ctx context.Context, op domain.Operation) error
    GetHistory(ctx context.Context) ([]domain.Operation, error)
    Ping(ctx context.Context) error
}
```

#### SaveOperation

Вставляет документ в коллекцию:

```go
func (r *OperationRepo) SaveOperation(ctx context.Context, op domain.Operation) error {
    doc := operationDoc{
        Number1:   op.Number1,
        Number2:   op.Number2,
        Operation: op.Operation,
        Result:    op.Result,
        Message:   op.Message,
        CreatedAt: op.Timestamp,
    }
    _, err := r.client.Coll().InsertOne(ctx, doc)
    return err
}
```

#### GetHistory

Читает все документы, сортирует по `created_at` (новые первые):

```go
func (r *OperationRepo) GetHistory(ctx context.Context) ([]domain.Operation, error) {
    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
    cursor, err := r.client.Coll().Find(ctx, bson.M{}, opts)
    // ... итерация по cursor, маппинг в domain.Operation
}
```

#### Ping

Проверка доступности (для health-check):

```go
func (r *OperationRepo) Ping(ctx context.Context) error {
    return r.client.Ping(ctx, nil)
}
```

### Структура документа

```go
type operationDoc struct {
    Number1   float64   `bson:"number1"`
    Number2   float64   `bson:"number2"`
    Operation string    `bson:"operation"`
    Result    float64   `bson:"result"`
    Message   string    `bson:"message"`
    CreatedAt time.Time `bson:"created_at"`
}
```

Тег `bson:"..."` задаёт имя поля в документе MongoDB (аналог `json:"..."` для JSON).

---

## Фича-флаг: выбор хранилища

В конфиге приложения есть флаг:

```go
type FeatureFlags struct {
    UsePGStorage bool `envconfig:"PG"`
}
```

Переменная окружения: **`CALCULATOR_FEATURE_PG`**

| Значение | Хранилище |
|----------|-----------|
| `true`, `1`, `yes` | PostgreSQL |
| `false`, `0`, `no` | MongoDB |

При старте приложения (`app.Run`):

```go
if a.cfg.FeatureFlags.UsePGStorage {
    // подключение к PostgreSQL, миграции, pg.NewOperationRepo
    log.Info("storage: PostgreSQL")
} else {
    // подключение к MongoDB, mongo.NewOperationRepo
    log.Info("storage: MongoDB")
}
```

---

## Docker Compose

В `deployment/localCalc/docker-compose.yml` подняты:

### Сервис `mongodb`

```yaml
mongodb:
  container_name: lizzycalc-mongodb
  image: mongo:7
  ports:
    - "27017:27017"
  healthcheck:
    test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
    interval: 5s
    timeout: 5s
    retries: 5
```

- **Порт 27017** — стандартный порт MongoDB.
- **Healthcheck** — команда `mongosh` проверяет, что сервер отвечает.

### Сервис `mongo-express`

```yaml
mongo-express:
  container_name: lizzycalc-mongo-express
  image: mongo-express:1.0.2-20-alpine3.19
  ports:
    - "8082:8081"
  environment:
    ME_CONFIG_MONGODB_URL: mongodb://mongodb:27017/
    ME_CONFIG_BASICAUTH: "false"
  depends_on:
    mongodb:
      condition: service_healthy
```

**Mongo Express** — веб-интерфейс для MongoDB (аналог Redis Insight для Redis, Kafka UI для Kafka).

- **URL:** http://localhost:8082
- **Функции:** просмотр баз, коллекций, документов; выполнение запросов; редактирование.

### Calculator: переменные MongoDB

```yaml
calculator:
  environment:
    CALCULATOR_FEATURE_PG: "true"  # или "false" для MongoDB
    CALCULATOR_MONGO_URI: mongodb://mongodb:27017
    CALCULATOR_MONGO_DATABASE: lizzycalc
    CALCULATOR_MONGO_COLLECTION: operations
```

---

## Локальная разработка

### Запуск с MongoDB

1. В `.env` установи:
   ```
   CALCULATOR_FEATURE_PG=false
   ```

2. Подними контейнеры:
   ```bash
   docker compose -f deployment/localCalc/docker-compose.yml up -d
   ```

3. Проверь логи калькулятора:
   ```bash
   docker logs lizzycalc-calculator 2>&1 | grep storage
   # storage: MongoDB
   ```

4. Открой Mongo Express: http://localhost:8082
   - База: `lizzycalc`
   - Коллекция: `operations`

### Запуск Go-приложения локально (без Docker)

1. Убедись, что MongoDB запущена (контейнер или локальная установка).

2. В `.env`:
   ```
   CALCULATOR_FEATURE_PG=false
   CALCULATOR_MONGO_URI=mongodb://localhost:27017
   CALCULATOR_MONGO_DATABASE=lizzycalc
   CALCULATOR_MONGO_COLLECTION=operations
   ```

3. Запусти:
   ```bash
   go run cmd/calculator/main.go
   ```

---

## Работа с MongoDB через mongosh

**mongosh** — официальный CLI-клиент MongoDB.

### Подключение к контейнеру

```bash
docker exec -it lizzycalc-mongodb mongosh
```

### Основные команды

```javascript
// Показать базы
show dbs

// Переключиться на базу
use lizzycalc

// Показать коллекции
show collections

// Показать все документы
db.operations.find()

// Показать последние 5 операций
db.operations.find().sort({ created_at: -1 }).limit(5)

// Найти операции сложения
db.operations.find({ operation: "+" })

// Посчитать количество документов
db.operations.countDocuments()

// Удалить все документы (осторожно!)
db.operations.deleteMany({})

// Показать индексы
db.operations.getIndexes()
```

### Примеры запросов

```javascript
// Все операции с результатом > 100
db.operations.find({ result: { $gt: 100 } })

// Операции за последний час
db.operations.find({
  created_at: { $gte: new Date(Date.now() - 3600000) }
})

// Группировка по типу операции
db.operations.aggregate([
  { $group: { _id: "$operation", count: { $sum: 1 } } }
])

// Средний результат по типу операции
db.operations.aggregate([
  { $group: { _id: "$operation", avgResult: { $avg: "$result" } } }
])
```

---

## Go-драйвер: go.mongodb.org/mongo-driver/v2

### Установка

```bash
go get go.mongodb.org/mongo-driver/v2/mongo
```

### Импорты

```go
import (
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)
```

### Подключение (v2)

```go
client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    return err
}
defer client.Disconnect(context.Background())

// Проверка соединения
err = client.Ping(context.Background(), nil)
```

**Важно:** в v2 функция `Connect` **не принимает context** (в отличие от v1). Context передаётся в `Ping` и другие операции.

### Вставка документа

```go
coll := client.Database("lizzycalc").Collection("operations")

doc := bson.M{
    "number1":    10.5,
    "number2":    3.0,
    "operation":  "+",
    "result":     13.5,
    "message":    "",
    "created_at": time.Now(),
}

result, err := coll.InsertOne(context.Background(), doc)
fmt.Println("Inserted ID:", result.InsertedID)
```

### Чтение документов

```go
// Все документы
cursor, err := coll.Find(context.Background(), bson.M{})
defer cursor.Close(context.Background())

var results []bson.M
err = cursor.All(context.Background(), &results)

// С фильтром и сортировкой
filter := bson.M{"operation": "+"}
opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(10)
cursor, err := coll.Find(context.Background(), filter, opts)
```

### Создание индекса

```go
indexModel := mongo.IndexModel{
    Keys:    bson.D{{Key: "created_at", Value: -1}},
    Options: options.Index().SetName("created_at_desc"),
}
_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
```

---

## BSON: типы данных

**BSON** (Binary JSON) — формат хранения в MongoDB. Поддерживает больше типов, чем JSON.

| BSON-тип | Go-тип | Пример |
|----------|--------|--------|
| String | `string` | `"hello"` |
| Int32 | `int32` | `42` |
| Int64 | `int64` | `9223372036854775807` |
| Double | `float64` | `3.14` |
| Boolean | `bool` | `true` |
| Date | `time.Time` | `ISODate("2024-01-15T10:30:00Z")` |
| ObjectId | `primitive.ObjectID` | `ObjectId("507f1f77bcf86cd799439011")` |
| Array | `[]any` | `["a", "b", "c"]` |
| Embedded Document | `struct` или `bson.M` | `{ "nested": "value" }` |
| Null | `nil` | `null` |

### bson.M vs bson.D

- **`bson.M`** — map (`map[string]any`), порядок полей **не гарантирован**.
- **`bson.D`** — slice of key-value pairs, порядок полей **сохраняется**.

Используй `bson.D` для:
- Сортировки: `bson.D{{Key: "created_at", Value: -1}}`
- Индексов
- Запросов, где важен порядок

Используй `bson.M` для:
- Простых фильтров: `bson.M{"operation": "+"}`
- Документов для вставки (если порядок не важен)

---

## Сравнение: SQL vs MongoDB Query

| Операция | SQL (PostgreSQL) | MongoDB |
|----------|------------------|---------|
| Выбрать все | `SELECT * FROM operations` | `db.operations.find()` |
| С условием | `SELECT * FROM operations WHERE operation = '+'` | `db.operations.find({ operation: "+" })` |
| Сортировка | `ORDER BY created_at DESC` | `.sort({ created_at: -1 })` |
| Лимит | `LIMIT 10` | `.limit(10)` |
| Подсчёт | `SELECT COUNT(*) FROM operations` | `db.operations.countDocuments()` |
| Вставка | `INSERT INTO operations (...) VALUES (...)` | `db.operations.insertOne({...})` |
| Обновление | `UPDATE operations SET result = 10 WHERE id = 1` | `db.operations.updateOne({ _id: ... }, { $set: { result: 10 } })` |
| Удаление | `DELETE FROM operations WHERE id = 1` | `db.operations.deleteOne({ _id: ... })` |
| Группировка | `SELECT operation, COUNT(*) FROM operations GROUP BY operation` | `db.operations.aggregate([{ $group: { _id: "$operation", count: { $sum: 1 } } }])` |

---

## Резюме

1. **MongoDB** — документная NoSQL БД, хранит JSON-подобные документы в коллекциях.
2. **Коллекция** — аналог таблицы, но без строгой схемы.
3. **Документ** — JSON-объект с полями; у каждого есть `_id`.
4. **В проекте** MongoDB используется как альтернативное хранилище операций (вместо PostgreSQL).
5. **Выбор хранилища** — по фича-флагу `CALCULATOR_FEATURE_PG` (true = PG, false = Mongo).
6. **Интерфейс** `IOperationRepository` един для обоих хранилищ — бизнес-логика не знает, куда пишет.
7. **Mongo Express** (http://localhost:8082) — веб-UI для просмотра и редактирования данных.
8. **mongosh** — CLI для запросов и администрирования.
