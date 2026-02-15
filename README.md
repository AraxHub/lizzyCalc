# lizzyCalc

Калькулятор: Go (Clean Architecture), PostgreSQL, Redis, kafka, React.

**Логика проекта:** материал разложен по веткам. Нужно переходить с ветки на ветку по порядку и читать документацию (README и код) в каждой — так по шагам собирается полная картина.

**Требования:** для поднятия окружения (БД, Redis, Kafka и т.д. по веткам) нужна установка **Docker**. Рекомендуется **Docker Desktop** (Windows/macOS). Полезно понимание **SQL** на базовом уровне (SELECT, INSERT, простые запросы). Для отладки API лучше пользоваться **Postman**; если впадлу — сойдёт и **curl**.

---

## Ветки: порядок прохождения

Изучать по порядку. **Пока досконально не разобрался, как работает текущая ветка — на следующую не переходи.**

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

- [ ] **feature-kafka-broker** — интеграция с Kafka, после расчёта просто пишем в топик, консюмер читает и логирует
- [ ] **feature-clickhouse-analytics** — аналитика в ClickHouse: запись операций/событий в колоночную БД, отдельный handler/use case для агрегатов: сколько по типам операций, по часам, топ выражений. Не замена PG, а доп. хранилище под отчёты.
- [ ] **feature-mongo-storage** — хранилище операций в MongoDB (документная БД по коллекциям). Фича-флаг в конфиге: в зависимости от него используется либо PostgreSQL, либо Mongo (один интерфейс репозитория, два адаптера).
- [ ] **feature-tests** — тесты: юнит-тесты (логика use case, домен), моки для портов (репозиторий, кэш), интеграционные тесты (API или слой инфраструктуры против реальной/тестовой БД).


---

## Docker и контейнеры

**От ветки к ветке** меняется и код, и состав сервисов в compose (на одной ветке только postgres + calculator, на другой добавлены redis, redisinsight и т.д.). Образ **calculator** собирается из кода текущей ветки. Если переключился на другую ветку и не пересобрал образ — в контейнере продолжает крутиться старый бинарник (например, с Redis), и приложение падает или ведёт себя не так, как в коде текущей ветки.

**Почему «пересборка без изменений»:** Docker кэширует слои. Если не сносить старые образы (`--rmi local`) и не собирать без кэша (`--no-cache`), может подтянуться закэшированный слой с другим кодом, и по факту образ остаётся старым. Поэтому при смене ветки лучше делать полную пересборку: остановить контейнеры, удалить локальные образы, собрать заново без кэша, поднять.

**Ниже команды** — их можно выполнять после переключения ветки, чтобы контейнеры соответствовали текущей ветке.

### Бэкенд (postgres + calculator; на части веток ещё redis, redisinsight, кафки и тд)

Из корня репо. После смены ветки выполни:

```bash
cd deployment/localCalc
docker compose -p localcalc down --rmi local 
docker compose -p localcalc build --no-cache
docker compose -p localcalc up -d
```

- `down --rmi local -v` — останавливает контейнеры, удаляет образы, собранные этим compose, и тома. Чтобы **не терять данные БД**, можно добавить `-v`, но проще и данные сносить, чтобы понимать, как происходит наполнение бд.
- `build --no-cache` — сборка образов без кэша (в образ попадёт код текущей ветки).
- Состав сервисов будет таким, какой описан в `deployment/localCalc/docker-compose.yml` на этой ветке (например, на feature-redis-cache поднимется ещё redis и redisinsight).

### Фронт (отдельный compose)

```bash
cd deployment/frontend
docker compose -p lizzycalc-frontend down --rmi local
docker rm -f lizzycalc-frontend 2>/dev/null || true
docker compose -p lizzycalc-frontend build --no-cache
docker compose -p lizzycalc-frontend up -d
```

`docker rm -f lizzycalc-frontend` — на случай, если контейнер с таким именем остался и мешает созданию нового.

### Пример: перешёл на feature-redis-cache

```bash
git checkout feature-redis-cache
cd deployment/localCalc
docker compose -p localcalc down --rmi local 
docker compose -p localcalc build --no-cache
docker compose -p localcalc up -d
```

Потом при необходимости фронт (если на этой ветке он есть и нужен):

```bash
cd deployment/frontend
docker compose -p lizzycalc-frontend down --rmi local
docker rm -f lizzycalc-frontend 2>/dev/null || true
docker compose -p lizzycalc-frontend build --no-cache && docker compose -p lizzycalc-frontend up -d
```

### Пример: перешёл на feature-frontend (движение назад по программе)

На этой ветке в compose только postgres + calculator (без Redis). Чтобы не крутился старый образ с Redis:

```bash
git checkout feature-frontend
cd deployment/localCalc
docker compose -p localcalc down --rmi local -v
docker compose -p localcalc build --no-cache
docker compose -p localcalc up -d
```

Фронт — те же команды, что выше (deployment/frontend).

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
