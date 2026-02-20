# Бэкенд: deployment/localCalc (postgres, calculator, redis, kafka, zookeeper и т.д.)
# Фронт: deployment/frontend (отдельный compose)
#
# Две основные команды на сторону:
#   — сборка с нуля (down, build без кэша, up)
#   — пересборка только приложения без кэша (build --no-cache приложения, up)

COMPOSE_BACKEND_DIR := deployment/localCalc
COMPOSE_FRONTEND_DIR := deployment/frontend
BACKEND_PROJECT := localcalc
BACKEND_APP_SERVICE := calculator
FRONTEND_PROJECT := lizzycalc-frontend
FRONTEND_CONTAINER := lizzycalc-frontend

# Удалять тома при backend-down (БД и т.д. с нуля). Использование: make backend-from-zero DROP_VOLUMES=1
DROP_VOLUMES ?= 0
DOWN_VOL_ARGS := $(if $(filter 1,$(DROP_VOLUMES)),-v,)

.PHONY: help
.PHONY: backend-from-zero backend-app backend-down backend-build backend-up
.PHONY: frontend-from-zero frontend-app frontend-down frontend-build frontend-up
.PHONY: kafka-create-topic
.PHONY: test test-v test-coverage test-run
.PHONY: mocks

help:
	@echo "Бэкенд:"
	@echo "  make backend-from-zero  — сборка всех контейнеров бэка с нуля (down, build --no-cache, up). DROP_VOLUMES=1 — удалить тома."
	@echo "  make backend-app        — пересборка только приложения (calculator) без кэша и up"
	@echo "  make backend-down       — остановить всё, удалить образы (DROP_VOLUMES=1 — и тома)"
	@echo "  make backend-build      — собрать образ calculator (без down/up)"
	@echo "  make backend-up        — поднять контейнеры без сборки"
	@echo ""
	@echo "Фронт:"
	@echo "  make frontend-from-zero — сборка фронта с нуля (down, build --no-cache, up)"
	@echo "  make frontend-app       — пересборка приложения фронта без кэша и up"
	@echo "  make frontend-down      — остановить и удалить образы"
	@echo "  make frontend-build     — собрать без кэша"
	@echo "  make frontend-up        — поднять контейнер"
	@echo ""
	@echo "Опционально:"
	@echo "  make kafka-create-topic — создать топик operations в Kafka"
	@echo ""
	@echo "Тесты:"
	@echo "  make test               — запустить все тесты"
	@echo "  make test-v             — тесты с verbose"
	@echo "  make test-coverage      — тесты + HTML-отчёт о покрытии (coverage.html)"
	@echo "  make test-run NAME=...  — запустить тест по имени (NAME=TestCacheKey)"
	@echo "  make mocks              — сгенерировать моки (mockgen)"

# --- Backend ---

backend-down:
	cd $(COMPOSE_BACKEND_DIR) && docker compose -p $(BACKEND_PROJECT) down --rmi local $(DOWN_VOL_ARGS)

backend-build:
	@echo "Сборка образа $(BACKEND_APP_SERVICE)..."
	cd $(COMPOSE_BACKEND_DIR) && docker compose --progress=quiet -p $(BACKEND_PROJECT) build $(BACKEND_APP_SERVICE)
	@echo "Готово."

backend-build-all:
	@echo "Сборка всех образов бэка без кэша..."
	cd $(COMPOSE_BACKEND_DIR) && docker compose --progress=quiet -p $(BACKEND_PROJECT) build --no-cache
	@echo "Готово."

backend-up:
	cd $(COMPOSE_BACKEND_DIR) && docker compose -p $(BACKEND_PROJECT) up -d

# Сборка всего бэка с нуля: down, build --no-cache, up
backend-from-zero: backend-down backend-build-all backend-up

# Пересборка только приложения без кэша и up
backend-app:
	@echo "Пересборка приложения $(BACKEND_APP_SERVICE) без кэша..."
	cd $(COMPOSE_BACKEND_DIR) && docker compose --progress=quiet -p $(BACKEND_PROJECT) build --no-cache $(BACKEND_APP_SERVICE)
	@echo "Готово."
	$(MAKE) backend-up

# --- Frontend ---

frontend-down:
	cd $(COMPOSE_FRONTEND_DIR) && docker compose -p $(FRONTEND_PROJECT) down --rmi local
	-docker rm -f $(FRONTEND_CONTAINER) 2>/dev/null || true

frontend-build:
	@echo "Сборка фронта без кэша..."
	cd $(COMPOSE_FRONTEND_DIR) && docker compose --progress=quiet -p $(FRONTEND_PROJECT) build --no-cache
	@echo "Готово."

frontend-up:
	cd $(COMPOSE_FRONTEND_DIR) && docker compose -p $(FRONTEND_PROJECT) up -d

# Сборка фронта с нуля: down, build --no-cache, up
frontend-from-zero: frontend-down frontend-build frontend-up

# Пересборка приложения фронта без кэша и up (без down)
frontend-app: frontend-build frontend-up

# --- Kafka ---

kafka-create-topic:
	docker exec lizzycalc-kafka kafka-topics --create --bootstrap-server localhost:29092 --topic operations --partitions 3 --replication-factor 1 2>/dev/null || true

# --- Tests ---

.PHONY: test test-v test-coverage test-run

# Запустить все тесты
test:
	go test ./...

# Запустить тесты с verbose
test-v:
	go test ./... -v

# Тесты с покрытием + HTML-отчёт
test-coverage:
	go test ./... -coverprofile=coverage.out

# Запустить конкретный тест по имени (usage: make test-run NAME=TestCacheKey)
test-run:
	go test ./... -v -run $(NAME)

# Сгенерировать моки из интерфейсов (требуется mockgen: go install go.uber.org/mock/mockgen@latest)
mocks:
	@echo "Генерация моков..."
	go generate ./internal/ports/...
	@echo "Готово. Моки в internal/mocks/"
