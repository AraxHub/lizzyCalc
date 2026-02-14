# lizzyCalc

Учебный микросервис-калькулятор на Go: REST API, PostgreSQL, слоистая архитектура, graceful shutdown.

---

## Архитектура

### 1. Основные принципы чистой архитектуры

- **Правило зависимостей**: зависимости направлены только внутрь, к домену. Внешние слои (HTTP, БД) не знают друг о друге, они зависят от интерфейсов, заданных ядром.
- **Независимость ядра**: домен и бизнес-правила (use case) не зависят от фреймворков, БД, HTTP. Их можно переиспользовать и тестировать без инфраструктуры.
- **Порты и адаптеры**: ядро объявляет интерфейсы (порты); конкретные реализации (адаптеры) — HTTP-контроллеры, репозиторий на PostgreSQL — живут снаружи и подставляются при сборке приложения.

### 2. Слои в приложении

- **cmd/calculator** — точка входа: загрузка конфига, создание App, вызов Run().
- **internal/app** — сборка приложения: подключение к БД, миграции, логгер, создание репозитория, use case, HTTP-сервера с контроллерами, запуск с graceful shutdown.
- **internal/domain** — сущности (например, Operation). Без зависимостей.
- **internal/ports** — интерфейсы: OperationRepository (сохранение/чтение операций, Ping), CalculatorUseCase (Calculate, History).
- **internal/usecase/calculator** — бизнес-логика калькулятора; зависит только от порта репозитория и домена.
- **internal/api/http** — входной адаптер: сервер (Gin), контроллеры (system — liveness/readyness, calculator — calculate/history), middlewares (RequestLogger).
- **internal/repository/pg** — выходной адаптер: реализация OperationRepository на PostgreSQL, миграции (таблица operations).
- **internal/pkg/logger** — общая утилита логирования (файл + консоль).

### 3. Как реализована чистая архитектура

Контроллеры принимают интерфейс **CalculatorUseCase** и вызывают Calculate/History; не знают про БД. Use case принимает интерфейс **OperationRepository** и сохраняет/читает операции; не знает про Gin и PostgreSQL. Репозиторий **pg.OperationRepo** реализует OperationRepository и работает с БД. Домен **Operation** — общая структура для всех слоёв. В **app.Run()** создаётся одна реализация репозитория и use case, они передаются в контроллеры; замена БД или способа ввода (например, другой транспорт) делается без изменения домена и use case — только новыми адаптерами и сборкой в app.

---

## Конфигурация

Переменные окружения (префикс **CALCULATOR_**):

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| CALCULATOR_SERVER_HOST | Хост HTTP-сервера | 0.0.0.0 |
| CALCULATOR_SERVER_PORT | Порт HTTP-сервера | 8080 |
| CALCULATOR_DB_HOST | Хост PostgreSQL | localhost |
| CALCULATOR_DB_PORT | Порт PostgreSQL | 5433 |
| CALCULATOR_DB_USER | Пользователь БД | postgres |
| CALCULATOR_DB_PASSWORD | Пароль БД | postgres |
| CALCULATOR_DB_NAME | Имя БД | lizzycalc |
| CALCULATOR_DB_SSLMODE | SSL режим | disable |

Конфиг загружается из `.env` (godotenv) и из окружения (envconfig). Пример: `deployment/localCalc/.env`.

---

## Запуск

**Локально** (нужен запущенный PostgreSQL на порту 5433 или свои CALCULATOR_DB_*):

```bash
go run ./cmd/calculator
```

**Docker Compose** (из каталога с docker-compose.yml):

```bash
cd deployment/localCalc && docker compose up --build
```

API: `http://localhost:8080`. БД с хоста: `localhost:5433`.

---

## API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /liveness | Liveness-пробник |
| GET | /readyness | Readiness (проверка БД) |
| POST | /api/v1/calculate | Вычисление: JSON `{ "number1", "number2", "operation" }` → `{ "result", "message" }` |
| GET | /api/v1/history | История операций → `{ "items": [ HistoryItem, ... ] }` |

---

## Логирование и завершение

- Логгер: `internal/pkg/logger` — пишет в **app.log** и в **stderr**.
- При старте логируется `application started` с адресом сервера.
- Middleware логирует каждый запрос (method, path, status, ip, latency_ms).
- **Graceful shutdown**: по SIGINT/SIGTERM сервер перестаёт принимать новые запросы и даёт до 10 секунд на завершение текущих, затем процесс выходит.
