# GoAcademy — Makefile
# Цели наполняются по мере прохождения ROADMAP. На этапе bootstrap — заглушки/инфраструктура.

COMPOSE := docker compose -f deploy/docker-compose.yml

.DEFAULT_GOAL := help

.PHONY: help
help: ## Показать список целей
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---- Инфраструктура (CHAPTER 0.3+) ----
.PHONY: db-up db-down db-logs
db-up: ## Поднять Postgres (docker compose)
	$(COMPOSE) up -d postgres

db-down: ## Остановить инфраструктуру
	$(COMPOSE) down

db-logs: ## Логи Postgres
	$(COMPOSE) logs -f postgres

# ---- Backend (CHAPTER 1+) ----
.PHONY: migrate-up migrate-down migrate-version sqlc test test-short vet lint build run fmt-check ci
migrate-up: ## Применить все миграции (нужен DATABASE_URL)
	go -C backend run ./cmd/migrate up

migrate-down: ## Откатить последнюю миграцию
	go -C backend run ./cmd/migrate down

migrate-version: ## Текущая версия схемы
	go -C backend run ./cmd/migrate version

sqlc: ## Сгенерировать типобезопасный код из SQL (sqlc, пин v1.31.1)
	cd backend && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate

test: ## Запустить все тесты backend (интеграционные нужны TEST_DATABASE_URL + Docker)
	go -C backend test ./...

test-short: ## Быстрые тесты backend (без Docker/БД-интеграции)
	go -C backend test -short ./...

vet: ## go vet по backend
	go -C backend vet ./...

fmt-check: ## Проверить форматирование (gofmt), как в CI
	@out="$$(gofmt -l backend)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

lint: ## Линтер backend (golangci-lint добавится позже; пока vet)
	go -C backend vet ./...

build: ## Сборка backend
	go -C backend build ./...

ci: fmt-check vet build test-short ## Локальная проверка как в CI (backend job)

run: ## Локальный запуск API (нужен DATABASE_URL; см. .env.example)
	go -C backend run ./cmd/api
