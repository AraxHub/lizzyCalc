# lizzyCalc

Калькулятор: Go (Clean Architecture), PostgreSQL, React. Окружение в `deployment/`: бэкенд — `deployment/localCalc`, фронт — `deployment/frontend`.

---

## 1. Поднятие с полного нуля

Команды выполняй из **корня репо** (или укажи полный путь к нему). Сначала бэк, потом фронт.

**Шаг 1 — бэкенд (postgres + API):**

```bash
cd deployment/localCalc
docker compose -p lizzycalc up -d --build
```

**Шаг 2 — фронт:**

```bash
cd deployment/frontend
docker compose -p lizzycalc-frontend up -d --build
```

**Куда заходить:**

| Что        | URL                                                |
|-----------|-----------------------------------------------------|
| Интерфейс | http://localhost:3000                               |
| API       | http://localhost:8080                              |
| БД        | localhost:5433 (user: postgres, pass: postgres, db: lizzycalc) |

Фронт с 3000 шлёт запросы по `/api` — nginx в контейнере проксирует их на бэк (host:8080). Бэк может быть в контейнере или запущен из отладки на 8080.

**Бэк из отладки вместо контейнера:** поднять только БД — `cd deployment/localCalc && docker compose -p lizzycalc up -d postgres`. В IDE запустить `cmd/calculator` (порт 8080, БД localhost:5433). Конфиг хоста — `deployment/localCalc/.env`. Потом поднять фронт как в шаге 2.

---

## 2. Пересборка с нуля без кэширования

Полная пересборка: контейнеры и локальные образы удаляются, сборка идёт без кэша. Используй после доработок кода или если что-то «залипло».

**Бэкенд (postgres + calculator):**

```bash
cd deployment/localCalc
docker compose -p lizzycalc down --rmi local -v
docker compose -p lizzycalc build --no-cache
docker compose -p lizzycalc up -d
```

- `--rmi local` — удалить образы, собранные этим compose.  
- `-v` — удалить тома (данные БД). Чтобы **сохранить БД**, убери `-v`.

**Фронт:**

```bash
cd deployment/frontend
docker compose -p lizzycalc-frontend down --rmi local
docker compose -p lizzycalc-frontend build --no-cache
docker compose -p lizzycalc-frontend up -d
```

Если при `up -d` ошибка про занятое имя контейнера `lizzycalc-frontend`:

```bash
docker rm -f lizzycalc-frontend
cd deployment/frontend
docker compose -p lizzycalc-frontend up -d
```

**Всё разом (бэк + фронт с нуля, без кэша):**

```bash
cd deployment/localCalc
docker compose -p lizzycalc down --rmi local -v
docker compose -p lizzycalc build --no-cache && docker compose -p lizzycalc up -d

cd ../frontend
docker compose -p lizzycalc-frontend down --rmi local
docker rm -f lizzycalc-frontend 2>/dev/null || true
docker compose -p lizzycalc-frontend build --no-cache && docker compose -p lizzycalc-frontend up -d
```

Чтобы при полной пересборке **не терять данные БД**, в первой команде убери `-v`:  
`docker compose -p lizzycalc down --rmi local` (без `-v`).
