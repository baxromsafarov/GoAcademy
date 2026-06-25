# GoAcademy — RULES (правила разработки)

> Конвенции, обязательные к соблюдению на всех этапах. Нарушение правил = этап не закрыт.
> Связанные файлы: [ARCHITECTURE.md](./ARCHITECTURE.md), [ROADMAP.md](./ROADMAP.md), [SYNC-PROTOCOL.md](./SYNC-PROTOCOL.md).

---

## 1. Структура каталогов (монорепо)

```
go-project/
├── backend/
│   ├── cmd/
│   │   ├── api/            # точка входа HTTP-API
│   │   └── migrate/        # утилита миграций (опц., если не через библиотеку в api)
│   ├── internal/
│   │   ├── config/         # загрузка и валидация конфига из ENV
│   │   ├── httpapi/        # роутер, middleware, handlers (по доменам), httpx-хелперы (имя httpapi во избежание клэша с net/http)
│   │   ├── auth/           # JWT, пароли, сессии
│   │   ├── user/           # профиль
│   │   ├── content/        # videos, articles, quizzes, problems, tracks, ...
│   │   ├── progress/       # прогресс, activity_log
│   │   ├── gamification/   # XP, уровни, бейджи, streaks, daily
│   │   ├── social/         # notes, bookmarks, leaderboard, certificates
│   │   ├── admin/          # админ-обработчики
│   │   ├── mailer/         # интерфейс + log-заглушка
│   │   ├── store/          # сгенерированный sqlc-код + обёртки (db package)
│   │   └── platform/       # инфра-утилиты: logging (slog), postgres (pool + migrate), validator, ids, clock, ...
│   ├── db/
│   │   ├── migrations/     # *.up.sql / *.down.sql (golang-migrate)
│   │   └── queries/        # *.sql для sqlc
│   ├── sqlc.yaml
│   ├── go.mod
│   └── ...
├── frontend/
│   ├── src/
│   │   ├── app/            # роутинг, providers
│   │   ├── pages/          # страницы разделов
│   │   ├── features/       # фиче-модули (auth, videos, quizzes, ...)
│   │   ├── components/     # UI (shadcn) и общие компоненты
│   │   ├── lib/            # api-клиент, query-хелперы, утилиты
│   │   ├── i18n/           # i18next конфиг + локали ru/en/uz/ja
│   │   └── ...
│   └── ...
├── code-runner/            # сервис изолированного исполнения (CHAPTER 16+)
├── deploy/                 # docker-compose, Dockerfiles, конфиги окружения
├── docs/                   # доп. документация
└── SPMS-файлы в корне: DESCRIPTION.md, ROADMAP.md, RULES.md, ARCHITECTURE.md, STATUS.yaml, SYNC-PROTOCOL.md, EMERGENCY_SYNC.yaml
```

Границы: `internal/http` зависит от доменных пакетов, домены — от `store`/`platform`. Домены **не** импортируют `http`. Циклы запрещены.

---

## 2. Стиль кода Go

- Форматирование: `gofmt`/`goimports` обязательны (CI проверяет).
- Линтер: `golangci-lint` (vet, staticcheck, errcheck, ineffassign, revive базово). Предупреждения не игнорируются без обоснования (`//nolint` с причиной).
- Именование: пакеты — короткие, в нижнем регистре, без подчёркиваний; экспортируемые идентификаторы — с doc-комментарием при нетривиальности; ошибки — `ErrXxx`, обёртки через `fmt.Errorf("...: %w", err)`.
- Контекст: `context.Context` — первый аргумент функций, делающих I/O; не хранить в структурах.
- Без глобального состояния: зависимости передаются явно (конструкторы `NewXxx`).
- SQL — только через sqlc; ручные конкатенации запросов запрещены (защита от инъекций).

---

## 3. Обработка ошибок

- Доменные ошибки определяются в доменном пакете (`var ErrNotFound = errors.New(...)`), маппятся в HTTP централизованно (`internal/http/errors.go`).
- Единый формат ответа об ошибке:
  ```json
  { "error": { "code": "validation_error", "message": "human readable", "details": [ ... ] } }
  ```
- Коды: `validation_error` (400/422), `unauthorized` (401), `forbidden` (403), `not_found` (404), `conflict` (409), `rate_limited` (429), `internal` (500).
- 5xx логируются с уровнем error и request-id; клиенту не отдаются внутренние детали/стектрейсы.
- Никаких `panic` в нормальном потоке; recover-middleware ловит непредвиденные паники → 500.

---

## 4. Логирование

- Только `log/slog`, структурированные поля. Формат (JSON/text) — по конфигу.
- Обязательные поля запроса: `request_id`, `method`, `path`, `status`, `duration_ms`.
- **Запрещено** логировать: пароли, токены (любые), хэши паролей, содержимое cookie, секреты конфига.

---

## 5. База данных и миграции

- Каждая миграция — пара `NNNN_name.up.sql` / `NNNN_name.down.sql`, номер монотонно растёт (zero-padded).
- `down` обязана откатывать `up`. Необратимые операции согласуются отдельно.
- Применённые миграции **не редактируются** — только новые.
- Внешние ключи и нужные индексы — в той же миграции, что и таблица.
- Все операции, меняющие XP/streak/прогресс, влияющие на несколько таблиц, — в одной транзакции.
- Денормализация (`user_stats`) допускается ради производительности рейтинга/дашборда, но обновляется консистентно.

---

## 6. Тесты

- Бизнес-логика покрывается обязательно: scoring квизов, начисление XP, уровни, расчёт streak, прогресс трека, выдача бейджей/сертификатов, ротация refresh-токенов.
- Стиль: табличные тесты; `testify` по необходимости (assert/require).
- Интеграционные тесты с БД — через временную/тестовую базу (или `testcontainers`, решение зафиксировать при CH18).
- Тест считается частью DoD этапа: нет тестов на ключевую логику → этап не закрыт.

---

## 7. Frontend-конвенции

- TypeScript strict; ESLint + Prettier; typecheck в DoD.
- Серверное состояние — TanStack Query (не дублировать в локальном стейте); query-ключи централизованы.
- Все пользовательские строки — через i18next-ключи; хардкод текста запрещён.
- Markdown рендерится только через санитайзер; подсветка Go обязательна для блоков кода.
- Компоненты доступны (a11y): семантика, фокус, alt/aria по необходимости.

---

## 8. Безопасность (чек-лист, проверяется на CH18 и при добавлении соответствующих фич)

- [x] Пароли — argon2id, хранится только хэш. *(CH2.1, internal/auth/password.go)*
- [x] Токены (verify/reset/refresh) — в БД хранится **хэш**, не сам токен; есть TTL и `used_at`/`revoked_at`. *(CH2.2–2.4)*
- [x] Refresh — httpOnly + Secure + SameSite cookie, ротация, reuse-detection. *(CH2.3; Secure через `COOKIE_SECURE`)*
- [x] Все запросы к БД параметризованы (sqlc), без конкатенации. *(sqlc + hand-written методы используют `$N`-плейсхолдеры)*
- [x] Markdown санитизируется на рендере (XSS). *(D-025: react-markdown без rehype-raw — сырой HTML не рендерится; safe URL-transform)*
- [x] CSRF-защита для cookie-сессий. *(refresh-cookie SameSite=Lax → браузер не шлёт его на cross-site POST; все state-changing эндпоинты авторизуются Bearer-заголовком, который нельзя выставить cross-origin — cookie сама по себе ничего не авторизует, кроме /auth/refresh)*
- [x] Rate limiting на auth-эндпоинтах (и на `/sandbox/run`). *(CH2.4 + CH16.2; per-IP token bucket с TTL-эвикцией; IP = реальный TCP-peer, RealIP убран — нельзя подменить через X-Forwarded-For)*
- [x] Валидация всего входа на бэкенде. *(apierr.Validation во всех сервисах/хендлерах)*
- [x] Исполнение пользовательского кода — изолировано (без сети, лимиты CPU/RAM/время/вывод). *(D-027; интеграционные тесты network-none/cap-drop/timeout/OOM/output)*
- [x] Секреты не в логах и не в репозитории. *(request-logger пишет только method/path/status/ip — без Authorization/cookie/body/query; `.env` в `.gitignore`, `.env.example` актуален; LogMailer-токен — только dev-stub)*

**Дополнительно (CH18.1):** security-заголовки на каждом ответе (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `Cross-Origin-Opener-Policy: same-origin`, `Permissions-Policy`, restrictive `Content-Security-Policy`) — middleware `securityHeaders`, покрыт тестом.

---

## 9. Конфигурация

- Только через ENV; `.env` локально (в `.gitignore`), `.env.example` — актуальный и документированный.
- Валидация на старте (fail-fast): отсутствие обязательной переменной = приложение не стартует с понятной ошибкой.
- Секреты по умолчанию отсутствуют (никаких «dev-секретов» в коде).

---

## 10. Git и коммиты

- Conventional Commits: `type(scope): summary`. Типы: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `build`, `ci`.
- `scope` — домен/глава (напр. `auth`, `quizzes`, `frontend`, `spms`).
- Один коммит = одно логически завершённое изменение; зелёный build/lint/test перед коммитом.
- Commit/push выполняются только по явному запросу пользователя.
- Ветку от default создавать перед изменениями, если работаем в git-флоу (уточняется при CH0.2).

---

## 11. Работа с SPMS (когда и как обновлять)

- **`STATUS.yaml` обновляется:**
  - при завершении каждого Stage (новый `status_id`, перенос в `completed`, сдвиг `current`, пересчёт `metrics`, обновление `next_actions`);
  - при принятии архитектурного решения (запись в `decisions` + в `ARCHITECTURE.md`);
  - при блокере (`current.status: blocked`, запись в `open_questions`).
- **`ARCHITECTURE.md`** — при любом значимом архитектурном решении.
- **`ROADMAP.md`** — при изменении плана (добавление/перенос этапов); ID этапов не переиспользуются.
- **`DESCRIPTION.md`** — при изменении продуктовых требований.
- Каждое обновление `STATUS.yaml` проходит валидацию по [SYNC-PROTOCOL.md](./SYNC-PROTOCOL.md).
- Нельзя переходить к следующему Stage, не закрыв DoD текущего и не обновив SPMS.

---

## 12. Принципы взаимодействия

- Не предполагать молча — при неопределённости задать вопрос (и при необходимости занести в `open_questions`).
- Не закрывать этап «на словах»: DoD — это проверяемые артефакты.
- Решения с альтернативами фиксируются с обоснованием (rationale).
