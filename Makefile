# Пересборка композников после смены ветки.
# Бэкенд: deployment/localCalc (postgres + calculator; на части веток ещё redis, redisinsight и т.д.)
# Фронт: deployment/frontend (отдельный compose)

COMPOSE_BACKEND_DIR := deployment/localCalc
COMPOSE_FRONTEND_DIR := deployment/frontend
BACKEND_PROJECT := localcalc
FRONTEND_PROJECT := lizzycalc-frontend
FRONTEND_CONTAINER := lizzycalc-frontend

# Удалять тома при backend down (чтобы снести данные БД и видеть наполнение с нуля).
# Использование: make backend-down DROP_VOLUMES=1  или  make backend DROP_VOLUMES=1
DROP_VOLUMES ?= 0
DOWN_VOL_ARGS := $(if $(filter 1,$(DROP_VOLUMES)),-v,)

.PHONY: backend-down backend-build backend-up backend frontend-down frontend-build frontend-up frontend help

help:
	@echo "Бэкенд (postgres + calculator; на части веток — redis, redisinsight и т.д.):"
	@echo "  make backend          — down, build --no-cache, up -d"
	@echo "  make backend-down     — остановить и удалить образы (DROP_VOLUMES=1 для удаления томов)"
	@echo "  make backend-build    — собрать образы без кэша"
	@echo "  make backend-up       — поднять контейнеры"
	@echo ""
	@echo "Фронт:"
	@echo "  make frontend         — down, rm контейнер, build --no-cache, up -d"
	@echo "  make frontend-down    — остановить и удалить образы"
	@echo "  make frontend-build   — собрать без кэша"
	@echo "  make frontend-up      — поднять контейнер"
	@echo ""
	@echo "Пример после смены ветки:"
	@echo "  git checkout feature-redis-cache"
	@echo "  make backend"
	@echo "  make frontend   # при необходимости"

# --- Backend (localCalc) ---

backend-down:
	cd $(COMPOSE_BACKEND_DIR) && docker compose -p $(BACKEND_PROJECT) down --rmi local $(DOWN_VOL_ARGS)

backend-build:
	cd $(COMPOSE_BACKEND_DIR) && docker compose -p $(BACKEND_PROJECT) build --no-cache

backend-up:
	cd $(COMPOSE_BACKEND_DIR) && docker compose -p $(BACKEND_PROJECT) up -d

backend: backend-down backend-build backend-up

# --- Frontend ---

frontend-down:
	cd $(COMPOSE_FRONTEND_DIR) && docker compose -p $(FRONTEND_PROJECT) down --rmi local
	-docker rm -f $(FRONTEND_CONTAINER) 2>/dev/null || true

frontend-build:
	cd $(COMPOSE_FRONTEND_DIR) && docker compose -p $(FRONTEND_PROJECT) build --no-cache

frontend-up:
	cd $(COMPOSE_FRONTEND_DIR) && docker compose -p $(FRONTEND_PROJECT) up -d

frontend: frontend-down frontend-build frontend-up
