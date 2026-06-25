# GoAcademy

Образовательная платформа для системного изучения языка **Go** — от новичка до уверенного разработчика.
Контент разных форматов в одном месте, прозрачный прогресс и геймификация.

> Полное описание — [DESCRIPTION.md](./DESCRIPTION.md). План — [ROADMAP.md](./ROADMAP.md). Правила — [RULES.md](./RULES.md). Архитектура — [ARCHITECTURE.md](./ARCHITECTURE.md).
> Состояние проекта (источник истины между сессиями) — [STATUS.yaml](./STATUS.yaml).

---

## Стек

- **Backend:** Go · chi · pgx + sqlc · golang-migrate · JWT (argon2id) · slog
- **Database:** PostgreSQL
- **Frontend:** React · TypeScript · Vite · TanStack Query · Tailwind · shadcn/ui · i18next (RU/EN/UZ/JA)
- **Code sandbox / judge:** изолированное исполнение Go в одноразовых Docker-контейнерах (без сети, лимиты CPU/RAM/времени/вывода) — песочница `/sandbox` и онлайн-судья алгозадач
- **Infra:** Docker Compose (`postgres` · `migrate` · `backend` · `frontend`)

## Возможности

Видео (YouTube + прогресс) · статьи (Markdown + подсветка Go) · квизы (single/multiple, разбор) ·
алгозадачи (редактор + онлайн-судья OK/WA/TLE/RE/CE) · учебные треки с прогрессом и сертификатами ·
шпаргалки · глоссарий · мини-проекты с чек-листами · песочница Go · геймификация (XP/уровни/streak/бейджи/
дейли) · лидерборд · заметки · закладки · профиль с аватаром · мультиязычный UI (RU/EN/UZ/JA) ·
тёмная тема · админ-панель (CRUD контента + управление пользователями).

## Требования (для разработки)

| Инструмент | Версия (мин.) | Нужен с |
|------------|---------------|---------|
| Go | 1.25+ | CHAPTER 1 |
| Docker + Docker Compose | актуальная | CHAPTER 0.3 |
| Node.js + npm | 20+ | CHAPTER 14 (frontend) |
| `golang-migrate`, `sqlc` | актуальные | CHAPTER 1 |

> Локальный запуск всего стека — через Docker Compose (см. ниже). Нативные инструменты нужны для разработки backend/frontend.

## Структура репозитория

```
backend/        # Go API (cmd/api, cmd/migrate, internal/*, db/migrations, db/queries, sqlc.yaml, Dockerfile)
frontend/       # React + TS + Vite SPA (src/*, Dockerfile, nginx.conf)
code-runner/    # Модель изоляции исполнения Go (README); раннер — backend/internal/runner
deploy/         # docker-compose.yml (полный стек)
docs/           # Дополнительная документация
.github/        # CI (GitHub Actions)
*.md / *.yaml   # SPMS: DESCRIPTION, ROADMAP, RULES, ARCHITECTURE, STATUS, SYNC-PROTOCOL, EMERGENCY_SYNC
```

## Быстрый старт — весь стек одной командой

```bash
# 1) (опц.) Скопировать пример окружения. Без .env стек тоже поднимется на dev-дефолтах.
cp .env.example .env        # затем ОБЯЗАТЕЛЬНО сменить JWT_SECRET и POSTGRES_PASSWORD

# 2) Собрать и поднять весь стек (postgres → migrate → backend → frontend)
docker compose -f deploy/docker-compose.yml up -d --build

# 3) Открыть приложение
#    Frontend:  http://localhost:3000
#    API:       http://localhost:8080/api/v1   (фронт ходит на него через /api, проксируя nginx-ом)
```

Порядок старта задан зависимостями: `postgres` (healthcheck) → `migrate` (применяет все миграции и
выходит) → `backend` (ждёт успешной миграции, поднимает API) → `frontend` (nginx отдаёт SPA и
**проксирует** `/api` и `/static` на backend, поэтому браузер ходит на один origin — CORS не нужен).

Остановить: `docker compose -f deploy/docker-compose.yml down` (данные БД переживают в томе
`goacademy-postgres-data`; добавьте `-v` чтобы удалить и их).

### Локальный запуск с песочницей (Windows, одним скриптом)

Песочница `/sandbox` и онлайн-судья исполняют Go в Docker, поэтому backend нужно
запускать **нативно** (с Go-тулчейном и доступом к Docker). Скрипт поднимает всё —
Postgres (Docker), миграции, **наполнение контентом**, backend (:8080, песочница
включена) и frontend (:5173):

```powershell
powershell -ExecutionPolicy Bypass -File deploy\start-local.ps1
# затем открыть http://localhost:5173
```

Наполнить/обновить контент отдельно: `make seed` (или `go -C backend run ./cmd/seed`) —
учебный роадмап Go на 4 языках (видео/статьи/квизы/задачи/проекты), идемпотентно.

### Песочница и судья (опционально)

Исполнение Go-кода (`/sandbox` и онлайн-судья) **выключено по умолчанию** (`SANDBOX_ENABLED=false`):
оно требует на хосте backend'а Go-тулчейн (кросс-компиляция) **и** доступ к Docker (запуск одноразовых
изолированных контейнеров). Минимальный backend-образ из Compose их не содержит. Для включения запускайте
backend на хосте с Docker (вариант Docker-out-of-Docker — монтирование `/var/run/docker.sock`) и
поставьте `SANDBOX_ENABLED=true`. Без этого задачи/песочница работают в режиме ручной отметки.

### Нативная разработка (без Docker для backend/frontend)

```bash
docker compose -f deploy/docker-compose.yml up -d postgres   # только БД
go -C backend run ./cmd/migrate up                            # миграции (make migrate-up)
go -C backend run ./cmd/api                                   # API на :8080 (make run)
cd frontend && npm install && npm run dev                     # Vite на :5173 (проксирует /api на :8080)
```

## Миграции БД

Миграции — версионированные пары файлов в `backend/db/migrations/` вида
`NNNN_name.up.sql` / `NNNN_name.down.sql` (номер монотонно растёт). Они **встроены** в бинарь
(`go:embed`), поэтому команда работает из любого каталога. Применённая версия отслеживается в
таблице `schema_migrations`. Раннер — `golang-migrate` (см. решение D-011 в ARCHITECTURE.md).

```bash
go -C backend run ./cmd/migrate up        # применить все ожидающие миграции
go -C backend run ./cmd/migrate down      # откатить последнюю
go -C backend run ./cmd/migrate version   # текущая версия схемы
```

`GET /readyz` у API проверяет доступность Postgres и возвращает 503, если БД недоступна;
`GET /healthz` — liveness (200, пока процесс жив).

## Переменные окружения

Полный список с комментариями — в [`.env.example`](./.env.example). Реальные значения только в локальном
`.env` (в `.gitignore`). Дефолты в `.env.example` совпадают с дефолтами в `deploy/docker-compose.yml`,
поэтому стек поднимается и без `.env`. Ключевые:

| Переменная | Назначение | Дефолт | Прод |
|------------|------------|--------|------|
| `JWT_SECRET` | секрет подписи access-токенов (≥32 символов) | dev-заглушка | **сменить** |
| `POSTGRES_PASSWORD` | пароль БД | `goacademy_dev_password` | **сменить** |
| `DATABASE_URL` | DSN Postgres (в Compose host=`postgres`) | — | по окружению |
| `APP_ENV` / `LOG_LEVEL` / `LOG_FORMAT` | окружение и логирование | `production`/`info`/`json` | |
| `COOKIE_SECURE` / `COOKIE_SAMESITE` | флаги refresh-cookie | `false`/`lax` | `true` (HTTPS) |
| `HTTP_PORT` / `FRONTEND_PORT` | публикуемые порты backend/frontend | `8080`/`3000` | |
| `SANDBOX_ENABLED` | песочница/судья (нужен Docker+Go) | `false` | по желанию |
| `RATE_LIMIT_AUTH_PER_MINUTE` / `RATE_LIMIT_SANDBOX_PER_MINUTE` | per-IP лимиты | `10`/`6` | |

## Команды (Makefile)

См. [`Makefile`](./Makefile): `migrate-up/down/version`, `sqlc`, `build`, `vet`, `fmt-check`,
`test` / `test-short`, `ci` (локальный паритет с backend-джобом CI), `run`, `db-up/down`.

## CI

[`.github/workflows/ci.yml`](./.github/workflows/ci.yml): backend (gofmt-check · `go vet` · build ·
`go test -short`), backend-integration (Postgres-сервис · миграции · `go test ./...`), frontend
(`npm ci` · lint · build). Линтеры блокирующие.

## Статус разработки

**Проект завершён: 57/57 этапов (главы 0–18).** Детали по этапам — в [ROADMAP.md](./ROADMAP.md);
журнал состояния и архитектурные решения (D-001…D-027) — в [STATUS.yaml](./STATUS.yaml) и
[ARCHITECTURE.md](./ARCHITECTURE.md).
