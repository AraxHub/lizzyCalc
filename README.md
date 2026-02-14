# lizzyCalc

Калькулятор: Go (Clean Architecture), PostgreSQL, Redis, kafka, React.

**Логика проекта:** материал разложен по веткам. Нужно переходить с ветки на ветку по порядку и читать документацию (README и код) в каждой — так по шагам собирается полная картина.

---

## Ветки: порядок прохождения

Изучать по порядку:

| № | Ветка | Что внутри |
|---|-------|------------|
| 1 | `feature-simple-storage` | Бизнес-логика (usecase), слайс как хранилище, ручные запросы в main, вывод истории. Без HTTP. |
| 2 | `feature-simple-handler` | Простейший HTTP-сервер, один хэндлер (JSON → расчёт → ответ). Без Gin. |
| 3 | `feature-simpleHandler-simpleStorage` | Калькулятор по HTTP: хэндлеры + usecase + сохранение в слайс + эндпоинт истории. |
| 4 | `feature-api-server` | API на Gin: роутер, группы, версии (v1/v2), health, production-настройки. |
| 5 | `feature-pg-storage` | PostgreSQL вместо слайса: docker-compose, подключение, save/get в БД, таблица operations. |
| 6 | `feature-env` | Конфиг из переменных окружения (порт, хост, строка подключения к БД и т.д.). |
| 7 | `feature-architecture` | Разделение слоёв: репозиторий (интерфейс), usecase без записи в слайс, чистая архитектура. |
| 8 | `feature-mvp` | Минимальный жизнеспособный продукт (MVP) только бэк: API + БД + базовая логика. |
| 9 | `feature-frontend` | Фронтенд (React): UI калькулятора, история, прокси `/api` на бэк, окружение в контейнерах (`deployment/localCalc`, `deployment/frontend`). и инструкция, как их поднимать |
| 10 | `feature-redis-cache` | Кэширование (Redis): порт Cache, реализация в `internal/infrastructure/redis`, проверка кэша в usecase перед расчётом, запись при промахе. История только из БД. |
| — | `main` | Текущая основная ветка. |

Кратко маршрут: логика и «БД» в памяти → HTTP → объединение → Gin и версионирование → реальная БД → env → архитектура → MVP → фронт → кэш (Redis).

---

## TODO (планируемые ветки)

- [ ] **feature-kafka-broker** — интеграция с Kafka (брокер сообщений)
- [ ] **feature-final-prod-calculator** — финальный production-калькулятор

---

## Git: краткие инструкции

**Клонировать репо**
```bash
git clone <url-репозитория>
cd lizzyCalc
```

**Обновить список веток с удалённого репо**
```bash
git fetch origin
```

**Переключиться на ветку**
```bash
git checkout feature-simple-storage
# или
git switch feature-simple-storage
```

**Узнать текущую ветку**
```bash
git branch
# звёздочка * — текущая
```

**Список всех веток (включая удалённые)**
```bash
git branch -a
```

**Вернуться на main**
```bash
git checkout main
```

**Подтянуть последние изменения в текущей ветке**
```bash
git pull origin <имя-ветки>
# например: git pull origin main
```

**Создать новую ветку и переключиться на неё**
```bash
git checkout -b feature-env
```
