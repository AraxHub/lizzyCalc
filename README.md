# lizzyCalc

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
| — | `main` | Текущая основная ветка. |

Кратко: сначала логика и «БД» в памяти → потом HTTP → потом объединение → потом Gin и версионирование → потом реальная БД.

---

## TODO (планируемые ветки)

- [ ] **feature-mvp** — минимальный жизнеспособный продукт (MVP)
- [ ] **feature-frontend** — фронтенд (UI для калькулятора)
- [ ] **feature-redis-cache** — кэширование (Redis)
- [ ] **feature-kafka-broker** — интеграция с Kafka (брокер сообщений)
- [ ] **feature-final-prod-calculator** — финальный production-калькулятор: env + архитектура + API + БД в одном

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
