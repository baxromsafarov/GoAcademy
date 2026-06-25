# GoAcademy — ARCHITECTURE

> Архитектура системы, модель данных и ключевые решения. Обновляется при принятии важных решений (синхронно с `decisions` в [STATUS.yaml](./STATUS.yaml)).
> Связанные файлы: [DESCRIPTION.md](./DESCRIPTION.md), [ROADMAP.md](./ROADMAP.md), [RULES.md](./RULES.md).

---

## 1. Компоненты системы

```
                         ┌────────────────────────────┐
                         │   Frontend (React/Vite)     │
                         │  TanStack Query · i18next   │
                         │  Tailwind · shadcn/ui       │
                         └─────────────┬──────────────┘
                                       │ HTTPS /api/v1 (JSON)
                                       │ refresh: httpOnly cookie
                         ┌─────────────▼──────────────┐
                         │      Backend (Go, chi)      │
                         │  http → domains → store     │
                         │  JWT · slog · sqlc/pgx      │
                         └──────┬───────────────┬─────┘
                                │               │ (CHAPTER 16+)
                   ┌────────────▼─────┐   ┌─────▼───────────────────┐
                   │   PostgreSQL      │   │   code-runner (изолят)  │
                   │   (pgx pool)      │   │   без сети, лимиты       │
                   └──────────────────┘   └─────────────────────────┘
```

- **Frontend** — SPA, общается только через REST `/api/v1`. Access-токен в памяти, refresh — в httpOnly cookie.
- **Backend** — слоистая структура: `http` (роутинг/middleware/handlers) → доменные сервисы → `store` (sqlc). Домены не знают про HTTP.
- **PostgreSQL** — единое хранилище. Доступ только через sqlc-запросы.
- **code-runner** — отдельный сервис для песочницы (CH16) и судьи (CH17); изоляция обязательна.

---

## 2. Слои бэкенда и поток запроса

`Request → middleware (request-id, recover, logging, rate-limit, auth) → handler → domain service → store (sqlc/pgx) → PostgreSQL`

- **handler**: парсинг/валидация DTO, вызов сервиса, сериализация ответа через `httpx`.
- **domain service**: бизнес-логика, транзакции, доменные ошибки.
- **store**: типобезопасные запросы (sqlc), управление транзакциями (`pgx.Tx`).

Единый формат ответа/ошибки — см. RULES §3.

---

## 3. Аутентификация (поток)

- **Регистрация:** создать `users` (`student`, `email_verified=false`) → выпустить verify-токен (в БД хранится хэш) → `Mailer` (заглушка логирует).
- **Логин:** проверка argon2id → выпуск access (JWT, короткий TTL) + refresh (случайный, хранится хэш в `refresh_sessions`), refresh в httpOnly cookie.
- **Refresh:** проверка хэша → ротация (старый помечается revoked, выдаётся новый); **reuse-detection** — если пришёл уже отозванный токен, инвалидируется вся цепочка сессии.
- **Logout:** revoke текущей refresh-сессии.
- **Reset password:** reset-токен (хэш в БД) → смена пароля → инвалидация активных сессий.
- **Авторизация:** `RequireAuth` (валидирует access JWT), `RequireRole("admin")`.

---

## 4. Модель данных (PostgreSQL)

> Имена/типы — ориентир; уточняются в миграциях. Все таблицы: `created_at`/`updated_at` где уместно, FK с `ON DELETE` по смыслу, индексы под выборки.

### 4.1. Пользователи и аутентификация
- **users**: `id (uuid pk)`, `email (citext unique)`, `password_hash`, `email_verified (bool)`, `role (enum: student|admin)`, `is_blocked (bool)`, `display_name`, `avatar_url`, `bio`, `location`, `locale (enum: ru|en|uz|ja)`, `is_public (bool)`, `created_at`, `updated_at`.
- **email_verification_tokens**: `id`, `user_id (fk)`, `token_hash`, `expires_at`, `used_at`.
- **password_reset_tokens**: `id`, `user_id (fk)`, `token_hash`, `expires_at`, `used_at`.
- **refresh_sessions**: `id`, `user_id (fk)`, `token_hash`, `user_agent`, `expires_at`, `revoked_at`, `created_at`. Индекс по `user_id`.

### 4.2. Контент
- **videos**: `id`, `title`, `description`, `youtube_id`, `duration_seconds`, `difficulty (enum)`, `tags (text[])`, `language`, `created_at`, `updated_at`.
- **articles**: `id`, `title`, `slug (unique)`, `body_markdown`, `difficulty`, `tags (text[])`, `language`, `created_at`, `updated_at`.
- **quizzes**: `id`, `title`, `description`, `pass_threshold (int, %)`, `language`.
- **quiz_questions**: `id`, `quiz_id (fk)`, `prompt`, `type (enum: single|multiple)`, `position`.
- **quiz_options**: `id`, `question_id (fk)`, `text`, `is_correct (bool)`, `position`.
- **problems**: `id`, `title`, `slug (unique)`, `statement_markdown`, `difficulty`, `reference_solution_markdown`, `sample_io (jsonb)`, `language`, `created_at`, `updated_at`.
- **problem_test_cases**: `id`, `problem_id (fk)`, `input`, `expected_output`, `is_sample (bool)`. (Используется судьёй, CH17.)
- **mini_projects**: `id`, `title`, `description_markdown`, `difficulty`, `language`.
- **mini_project_steps**: `id`, `project_id (fk)`, `text`, `position`.
- **cheatsheets**: `id`, `title`, `category`, `body_markdown`, `language`.
- **glossary_terms**: `id`, `term (unique)`, `definition_markdown`, `language`.
- **tracks**: `id`, `title`, `description`, `level`, `position`, `language`.
- **track_items**: `id`, `track_id (fk)`, `content_type (enum: video|article|quiz|problem|project)`, `content_id`, `position`. Полиморфизм по паре `(content_type, content_id)`; уникальность `(track_id, content_type, content_id)`.

### 4.3. Прогресс
- **video_progress**: `user_id (fk)`, `video_id (fk)`, `watched_percent`, `last_position_seconds`, `completed (bool)`, `updated_at`. **PK (user_id, video_id)**.
- **article_reads**: `user_id (fk)`, `article_id (fk)`, `completed_at`. **PK (user_id, article_id)**.
- **quiz_attempts**: `id`, `user_id (fk)`, `quiz_id (fk)`, `score`, `passed (bool)`, `answers (jsonb)`, `created_at`.
- **problem_submissions**: `id`, `user_id (fk)`, `problem_id (fk)`, `status (enum: attempted|solved)`, `code`, `language`, `verdict (jsonb, null до CH17)`, `created_at`.
- **project_progress**: `user_id (fk)`, `project_id (fk)`, `completed_steps (jsonb: int[])`, `updated_at`. **PK (user_id, project_id)**.

### 4.4. Геймификация и активность
- **user_stats**: `user_id (pk, fk)`, `total_xp`, `level`, `current_streak`, `longest_streak`, `last_active_date`.
- **badges**: `id`, `code (unique)`, `title`, `description`, `icon`, `criteria_type`, `criteria_params (jsonb)`.
- **user_badges**: `user_id (fk)`, `badge_id (fk)`, `awarded_at`. **PK (user_id, badge_id)**.
- **activity_log**: `id`, `user_id (fk)`, `activity_type`, `ref_type`, `ref_id`, `xp_earned`, `occurred_at`. Индексы: `(user_id, occurred_at)`. Источник для heatmap, рейтинга-за-период, XP.
- **daily_challenges**: `id`, `challenge_date (unique)`, `content_type`, `content_id`, `bonus_xp`.
- **user_daily_challenges**: `user_id (fk)`, `challenge_id (fk)`, `completed_at`. **PK (user_id, challenge_id)**.

### 4.5. Социальное и сертификаты
- **notes**: `id`, `user_id (fk)`, `content_type`, `content_id`, `body`, `created_at`, `updated_at`.
- **bookmarks**: `id`, `user_id (fk)`, `content_type`, `content_id`, `created_at`. Уникальность `(user_id, content_type, content_id)`.
- **certificates**: `id`, `user_id (fk)`, `track_id (fk)`, `certificate_code (unique)`, `issued_at`. Уникальность `(user_id, track_id)`.

### 4.6. Индексы (минимум)
- По `user_id` во всех прогресс/социальных таблицах.
- По `slug` (articles, problems), `term` (glossary), `code` (badges, certificates).
- По `(user_id, occurred_at)` в `activity_log`.
- По `challenge_date` в `daily_challenges`.

---

## 5. Полиморфные связи

`track_items`, `notes`, `bookmarks` ссылаются на контент парой `(content_type, content_id)`. Целостность обеспечивается на уровне приложения (валидация существования) + enum на `content_type`. FK на конкретные таблицы не ставим (полиморфизм); при удалении контента — сервисная чистка ссылок.

---

## 6. Геймификация: правила (детализируются в CH10–11)

- XP начисляется **в той же транзакции**, что и запись в `activity_log` — единый `gamification.Recorder` (реализует `activity.Recorder`), см. **D-017**.
- XP **идемпотентен** по `(user_id, activity_type, ref_id)`: повтор действия логируется (для heatmap), но XP даётся только в первый раз.
- XP-награды по типам: `video_completed` 10, `article_read` 5, `quiz_passed` 20, `quiz_attempt` 2, `problem_solved` 30, `project_completed` 50, `daily_challenge_completed` 15.
- **Уровень** = `1 + floor(sqrt(total_xp / 100))` (уровень L достигается при `100·(L-1)²` XP). Зафиксировано на 11.1.
- Streak (`current_streak`/`longest_streak`) считается чистой `computeStreak` по последовательным UTC-дням (D-010) в той же транзакции под `FOR UPDATE` — см. **D-018**.
- Бейджи выдаются данными-управляемым движком критериев (`criteria_type` + `criteria_params jsonb`), идемпотентно, в той же транзакции — см. **D-019**.
- Ежедневный вызов (один на UTC-день) при выполнении начисляет `bonus_xp` через тот же `Recorder` (`Event.XP`-override), идемпотентно по `user_daily_challenges` — см. **D-020**.

---

## 7. Интернационализация

- **UI:** i18next, языки RU/EN/UZ/JA; локаль хранится в `users.locale`.
- **Контент:** см. решение **D-008** ниже.

---

## 8. Журнал ключевых решений (decisions)

> Зеркалируется в `STATUS.yaml → decisions`. ID не переиспользуются.

- **D-001 — Монорепо.** Backend, frontend, code-runner, deploy в одном репозитории. *Причина:* единый цикл изменений и согласованность SPMS. *Альтернатива:* polyrepo (отклонено: оверхед на раннем этапе).
- **D-002 — argon2id для паролей.** *Причина:* современный memory-hard алгоритм, рекомендация OWASP. *Альтернатива:* bcrypt (допустимо, но argon2id предпочтительнее).
- **D-003 — Refresh-токены: хранить хэш, ротация + reuse-detection.** *Причина:* минимизация ущерба при утечке БД; обнаружение кражи токена. *Альтернатива:* stateless refresh без хранения (отклонено: нельзя отозвать).
- **D-004 — sqlc + pgx, без ORM.** *Причина:* типобезопасность, контроль SQL, производительность. *Альтернатива:* GORM (отклонено по требованию стека).
- **D-005 — golang-migrate для миграций.** *Причина:* стандарт, up/down пары. *Альтернатива:* goose (равнозначно; выбран golang-migrate).
- **D-006 — Полиморфные ссылки на контент через (content_type, content_id) без FK.** *Причина:* разнородный контент в track_items/notes/bookmarks. *Компромисс:* целостность на уровне приложения.
- **D-007 — Mailer как интерфейс с log-заглушкой.** *Причина:* SMTP подключается позже без изменения вызывающего кода. *Альтернатива:* сразу реальный SMTP (отклонено для MVP).
- **D-008 — Стратегия мультиязычного контента: `language` на контентных таблицах (фаза 1).** Каждая единица контента имеет поле `language`; выдача фильтруется по предпочитаемой локали с fallback на EN. Группировка переводов одной сущности (общий `translation_group_id`) — **опционально, поздняя фаза**, если потребуется связывать переводы. *Статус:* **принято 2026-06-25.**
- **D-009 — Хранение аватаров: локальный volume + статика на старте, абстракция `Storage`.** *Причина:* простота для Docker Compose; интерфейс позволит уйти в S3 позже без изменения вызывающего кода. `users.avatar_url` указывает на `/static/avatars/<id>`. *Статус:* **принято 2026-06-25.**
- **D-010 — Таймзона для streak/heatmap: фиксированная серверная (UTC) на старте, с заделом на per-user TZ.** *Причина:* детерминированность расчётов. Поле `users.timezone` (IANA) добавляется позже, когда понадобится. *Статус:* **принято 2026-06-25.**
- **D-011 — Миграции: golang-migrate с встроенными файлами (`go:embed` + source `iofs`), драйвер БД `lib/pq`; рантайм — `pgx`.** *Причина:* миграции встроены в бинарь и применяются из любого каталога/контейнера; `lib/pq` — стандартный драйвер golang-migrate, понимает тот же `postgres://` DSN, что и pgx. Раннер — `cmd/migrate` (up/down/version), версия в `schema_migrations`. Пул приложения подключается с retry; `/readyz` пингует БД. *Статус:* **принято 2026-06-25.** *Альтернатива:* CLI golang-migrate с файлами на ФС; драйвер `pgx/v5` в migrate.
- **D-014 — Контент-чтение: общий enum `difficulty` (beginner|intermediate|advanced), `language` = enum `locale`; паттерн списков — sqlc `List*`+`Count*` с опциональными фильтрами через `sqlc.narg(...)` (`narg IS NULL OR col = narg`), пагинация limit/offset (default 20, max 100, клампится в сервисе).** *Причина:* единообразные list-эндпоинты с фильтрами/пагинацией и точным total; индексы под фильтры (btree на difficulty/language/created_at, GIN на tags). Переиспользуется для статей/квизов/задач/треков. *Статус:* **принято 2026-06-25.**
- **D-013 — Сессии: access = JWT HS256 (короткий TTL), refresh = opaque-токен (хранится SHA-256-хэш) с ротацией и reuse-detection через `family_id`.** *Механика:* при логине создаётся новая `family_id`; каждый refresh ротирует токен (старый revoked, новый в той же семье); предъявление уже отозванного токена ⇒ утечка ⇒ `RevokeRefreshFamily` (отзыв всей линии) — выполняется на пуле (коммитится), не внутри откатываемой tx ротации. Refresh — httpOnly+SameSite cookie на пути `/api/v1/auth`. Access валидируется middleware `RequireAuth` (Bearer). `cmd/migrate` читает только `DATABASE_URL` (расцеплён от app-конфига, чтобы JWT-требования не мешали миграциям). *Статус:* **принято 2026-06-25.** *Альтернатива:* stateless refresh (нельзя отозвать); хранение самих токенов (отклонено — утечка БД).
- **D-012 — sqlc (pin v1.31.1) генерирует пакет `store` (sql_package `pgx/v5`) из `db/queries`, схема из `db/migrations` (down-файлы игнорируются).** *Причина:* типобезопасный доступ к БД без ORM; `Queries.WithTx` поддерживает транзакции (нужно для XP/streak). Единый формат ошибок API — пакеты `internal/platform/apierr` (APIError + коды) и `internal/httpapi/respond` (JSON/Error, тело `{error:{code,message,details}}`, 5xx логируются, причина не утекает). *Статус:* **принято 2026-06-25.** *Альтернатива:* sqlc с отдельным файлом схемы; ручной маппинг ошибок в каждом хендлере.
- **D-015 — `activity_log` как единый append-only журнал активности; `DBRecorder` над `store.DBTX`; `user_stats` отложена на CHAPTER 11.** Поля `activity_type`/`ref_type` — `text` (не enum) ради расширяемости без миграции на каждый новый тип события; `ref_id uuid NULL` — полиморфная ссылка без FK (D-006), допускает ref-less события (напр. daily login); `xp_earned int DEFAULT 0 CHECK >= 0` (наполняется в CH11); индекс `(user_id, occurred_at DESC)`. `internal/activity.DBRecorder` реализует `Recorder` поверх `store.DBTX` (пул **или** `pgx.Tx`), выполняя один атомарный `INSERT`; приём `Tx` — задел под CH11, где XP начисляется в одной транзакции с активностью. Невалидный `UserID`/`RefID` отклоняется до обращения к БД. `user_stats` (total_xp/level/current_streak/longest_streak/last_active_date) создаётся целиком в CH11 (11.1 XP, 11.2 streaks), где живёт его логика — `activity_log` остаётся источником истины, поэтому отсрочка ничего не теряет. *Статус:* **принято 2026-06-25** (объём подтверждён пользователем). *Альтернатива:* enum для `activity_type`/`ref_type`; создать `user_stats` уже в 10.1 (отклонено).
- **D-016 — Дашборд (CH10.2): `GET /me/progress` одним запросом; `GET /me/activity` — суточные UTC-бакеты с дефолтным окном год и капом 366 дней.** `ProgressSummary` собирает 5 счётчиков (завершённые видео / прочитанные статьи / пройденные квизы DISTINCT / решённые задачи DISTINCT / завершённые мини-проекты через `jsonb_array_length(completed_steps) >= step_count`) одним `SELECT` со скалярными подзапросами; все колонки alias-qualified, т.к. sqlc сливает scope подзапросов и иначе `user_id` неоднозначен. `ActivityHeatmap` группирует по UTC-дате (`(occurred_at AT TIME ZONE 'UTC')::date`, D-010) на полуоткрытом диапазоне `[from, to)`, отдаёт `day/count/sum(xp)`. `progress.ParseHeatmapRange(from, to, now)` — чистая функция (now инъектируется в HTTP-хендлер для детерминированных тестов): пустые значения → окно в год до сегодня; `from <= to`; кап 366 дней; формат `YYYY-MM-DD`. Дни без активности на бэкенде отсутствуют — разрежённый ряд заполняет фронтенд (меньше payload). *Статус:* **принято 2026-06-25.** *Альтернатива:* запрос на каждый счётчик; плотный ряд с нулями на бэкенде; локальная TZ пользователя сразу (отложено до `users.timezone`).
- **D-017 — Геймификация (CH11): пакет `internal/gamification`; `Recorder` пишет активность + XP + уровень в одной транзакции; XP идемпотентен по `(user, activity_type, ref)`; уровень = `1+floor(sqrt(xp/100))`.** `gamification.Recorder` реализует `activity.Recorder` и заменяет `activity.DBRecorder` в проде (для `progress` и `quiz`). В одной tx: проверка `ActivityExists` (XP начисляется только при первом вхождении `(user, type, ref)` — повтор логируется для heatmap, но не даёт XP → нет фарминга, критично для квизов, шлющих событие на каждую попытку), `InsertActivity(xp_earned)`, `AddUserXP` (апсерт `user_stats`: `total_xp += xp`, `last_active_date = GREATEST(...)`, row-lock сериализует конкурентные начисления), `SetUserLevel`. XP-политика (`XPFor`) и формула уровня (`LevelForXP`) централизованы и покрыты unit-тестами; конкурентность покрыта integration-тестом (20 горутин → без потерянных апдейтов). `GET /me/stats` отдаёт `total_xp/level/current_streak/longest_streak/last_active_date`. `activity.DBRecorder` (CH10.1) сохранён как activity-only recorder. *Статус:* **принято 2026-06-25.** *Альтернатива:* XP в каждом доменном сервисе; начислять на каждое событие (фармится); линейная формула; отдельный проход подсчёта XP. *Известный edge (low):* одновременный первый дубль одинакового события может начислить XP дважды (см. open_questions; митигация на 18.x).
- **D-018 — Streak (CH11.2): чистая `computeStreak` по последовательным UTC-дням, считается в Go под тем же `FOR UPDATE`, что и XP/level.** Запись `user_stats` переведена на `EnsureUserStats` (INSERT ON CONFLICT DO NOTHING) → `LockUserStats` (`SELECT … FOR UPDATE`) → `UpdateUserStats` (заменили `AddUserXP`/`SetUserLevel`): XP, level и streak вычисляются в Go и пишутся одной атомарной транзакцией под row-lock (конкурентные начисления сериализуются — нет потерянных апдейтов). `computeStreak(prevActive, hasPrev, prevCurrent, prevLongest, newDate)`: тот же или более старый день — без изменений; ровно следующий день — инкремент; разрыв более суток — сброс в 1; `longest = max(longest, current)`. Чистая функция → табличные тесты границ (требование DoD). День активности берётся как UTC-дата `occurred_at` (D-010); per-user TZ отложена до `users.timezone`. *Статус:* **принято 2026-06-25.** *Альтернатива:* streak в SQL `CASE` (хуже тестируется); отдельный джоб по `activity_log`; локальная TZ сразу.
- **D-019 — Бейджи (CH11.3): данные-управляемые критерии (`criteria_type` + `criteria_params jsonb`), плагинный evaluator, идемпотентная выдача в транзакции recorder'а.** `badges` хранит определения; `gamification.criterionMet` — `switch` по `criteria_type` (`xp_at_least {xp}`, `streak_at_least {days}`, `activity_count_at_least {activity_type, count}` — последний через `count(DISTINCT ref_id)` в `activity_log`); неизвестный тип возвращает «не выполнен», поэтому новый критерий не ломает существующие (DoD). `awardBadges` оценивает только ещё не полученные бейджи (`ListUnearnedBadges`) и пишет `user_badges` `ON CONFLICT DO NOTHING` — без дублей даже при конкуренции; вызывается в той же транзакции под `FOR UPDATE`, что XP/streak (консистентность с породившими статами). Начальный набор бейджей — сид в миграции 0019; полноценный CRUD — админка (CH13). `GET /me/badges` отдаёт полученные бейджи. *Статус:* **принято 2026-06-25.** *Альтернатива:* захардкоженные критерии; выдача отдельным джобом; полноценный rule-engine (избыточно для MVP).
- **D-020 — Ежедневный вызов (CH11.4): выполнение через `gamification.Recorder` с `Event.XP`-override = `bonus_xp`; идемпотентность на `user_daily_challenges`.** `Recorder` теперь чтит явный `Event.XP` (если >0) поверх `XPFor(type)` — переменный per-day `bonus_xp` проходит тем же атомарным путём (XP + streak + бейджи) без отдельной ветки начисления. `daily_challenges.challenge_date` UNIQUE — один вызов на UTC-день (D-010), контент полиморфно (`content_type` + `content_id`, D-006). `DailyService.Complete` делает `INSERT user_daily_challenges ON CONFLICT DO NOTHING RETURNING`: первое выполнение → `Record(daily_challenge_completed, XP=bonus_xp)`, повтор → без награды. Нет вызова на день → 404. Создание вызовов — админка (CH13). *Статус:* **принято 2026-06-25.** *Альтернатива:* фиксированный XP в `XPFor` (не учитывает per-day bonus); отдельная транзакция начисления; вызов без uniqueness по дате.
- **D-021 — Лидерборд (CH12.1): пакет `internal/social`; all-time по `user_stats.total_xp`, period по `SUM(activity_log.xp_earned)`; фильтр `is_public AND NOT is_blocked`; публичный эндпоинт.** Кросс-пользовательские фичи (лидерборд, далее заметки/закладки/сертификаты) живут в `social`. `LeaderboardAllTime` читает денормализованный `total_xp` (индекс `total_xp DESC`); `LeaderboardPeriod` суммирует `xp_earned` из `activity_log` за UTC-окно `[from, to)` (источник истины для рейтинга-за-период, `HAVING sum>0`). Видны только согласившиеся (`is_public`) и не заблокированные. `GET /leaderboard?period=all|week|month&limit=&offset=` — **публичный** (без auth, т.к. показывает только opt-in пользователей); `social.periodWindow(period, now)` — чистая функция (week=7д, month=30д) с инъекцией `now()` для тестов; ранг = `offset + i + 1`. *Статус:* **принято 2026-06-25.** *Альтернатива:* лидерборд внутри `gamification`; только all-time; требовать auth; отдельные периодные агрегаты (преждевременно).
- **D-022 — Сертификаты (CH12.4): выдача через явный `POST /tracks/{id}/certificate` с проверкой 100% (`social` зависит от `progress`); код `GOAC-<base32>`; публичная верификация по коду.** Завершение трека вычисляемо, не событие, поэтому `CertificatesService` переиспользует `progress.TrackProgress` для проверки `TrackComplete` перед выдачей (`social → progress`, без цикла). Выдача идемпотентна: `UNIQUE (user_id, track_id)` + `ON CONFLICT DO NOTHING RETURNING` (повтор → существующий сертификат). `certificate_code` = `GOAC-` + base32(10 случайных байт, `crypto/rand`) — публичный проверяемый идентификатор; `GET /certificates/{code}` без auth отдаёт `display_name + track_title + issued_at`; `POST /tracks/{id}/certificate` и `GET /me/certificates` — под auth. Явный claim-эндпоинт декуплирует прогресс от сертификатов и оставляет место под будущий автотриггер. *Статус:* **принято 2026-06-25.** *Альтернатива:* автовыдача внутри `GET /tracks/{id}/progress` (связывает прогресс с сертификатами); хранить завершение как флаг/событие; последовательный код.
- **D-023 — Фронтенд-стек (CH14): Vite 8 + React 19 + TS; React Router 7; Tailwind v4 (`@tailwindcss/vite`) + shadcn-совместимые токены; алиас `@`→`src`; тема через класс `.dark` + CSS-переменные.** `frontend/` создан через `create-vite` (react-ts). Tailwind v4 подключён Vite-плагином; `index.css` объявляет `@import "tailwindcss"`, `@custom-variant dark`, oklch CSS-переменные (`:root`/`.dark`) и `@theme inline` для семантических утилит (`bg-background`, `text-foreground`, `bg-primary`, `border-border`, …) — это даёт shadcn-стиль без интерактивного shadcn-CLI; реальные shadcn-компоненты копируются по мере надобности (CH15). Алиас `@`→`src` в `vite.config` (`resolve.alias`) и `tsconfig` (`paths`, без устаревшего в TS7 `baseUrl`). `lib/utils.cn` (clsx+tailwind-merge), `lib/sections` (разделы+lucide), `lib/theme.useTheme` (класс на `<html>` + localStorage), `components/Layout` (topbar+sidebar NavLink+Outlet, адаптив). Маршрутизация — `BrowserRouter`+`Routes`. **React Router** выбран из двух разрешённых промтом вариантов (проще TanStack Router для SPA). **TanStack Query** — серверное состояние, добавляется в 14.2. **i18next** — 14.3. *Статус:* **принято 2026-06-25.** *Альтернатива:* TanStack Router; полный shadcn-CLI; Tailwind v3.
- **D-024 — Auth на фронте (CH14.2): access-токен только в памяти, refresh в httpOnly-cookie; авто-refresh один раз при 401; bootstrap сессии через `refresh`→`/me`.** `lib/api` — тонкий fetch-клиент: `credentials:"include"` (refresh-cookie), `Authorization: Bearer` из in-memory токена, разбор `{error:{code,message,details}}` в `ApiError`; при 401 один раз дёргает `POST /auth/refresh` и повторяет запрос, при провале вызывает `onAuthFailure` (чистит сессию). `AuthProvider` на старте делает `refresh`→`/me` (восстановление сессии после перезагрузки), даёт `login/register/logout/user`. `ProtectedRoute` ждёт bootstrap и редиректит неаутентифицированных на `/login`. Access-токен НЕ в localStorage (XSS-устойчивость). Серверное состояние — TanStack Query. *Статус:* **принято 2026-06-25.** *Альтернатива:* токен в localStorage; без refresh-bootstrap; axios-интерсепторы.
- **D-027 — Code-runner (CH16.1): недоверенный Go компилируется на ХОСТЕ (CGO off, GOPROXY off → stdlib-only) и запускается статическим бинарником в одноразовом Docker-контейнере; раннер — in-process пакет `backend/internal/runner`, не отдельный микросервис.** Источник кросс-компилируется (`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOPROXY=off`) в статический бинарник — компилятор не исполняет исходник, поэтому сборка недоверенного кода безопасна, а `GOPROXY=off` делает песочницу **stdlib-only**. Бинарник запускается в контейнере с `--network=none`, `--cap-drop=ALL`, `--security-opt=no-new-privileges`, `--user=65534:65534` (nobody), `--memory`/`--memory-swap` (равны → swap off), `--cpus`, `--pids-limit`, `--tmpfs=/tmp`, wall-clock timeout (контейнер `docker kill`ается при превышении) и output cap (`cappedWriter`, не блокирует программу). Доставка бинарника: `docker create -i` → `docker cp` → `docker start -a -i` (rootfs НЕ `--read-only`, т.к. `docker cp` в read-only запрещён; компенсируется network none + cap-drop + non-root + эфемерный контейнер). OOM детектируется через `docker inspect .State.OOMKilled`. Образ запуска — `busybox` (статический бинарник, без Go-тулчейна в sandbox). Компиляция на хосте (а не in-container) даёт sub-second прогоны (~0.4s vs ~10s холодных) и не даёт лимиту памяти убить сам компилятор. `Runner.Run(ctx, Request) → Result` (stdout/stderr/exit/compileError/timedOut/oomKilled/truncated/duration). Покрыто интеграционными тестами (реальные контейнеры): timeout, OOM, large-output, no-network, compile-error, stdin, non-stdlib-import. *Deploy (CH18):* backend нужен Go-тулчейн (кросс-компиляция) + доступ к Docker (DooD / socket); хардening (seccomp, выделенный low-trust хост, pin образов) — на security-pass. *Статус:* **принято 2026-06-26.** *Альтернатива:* компиляция в контейнере (golang:alpine — медленно, риск OOM компилятора); отдельный микросервис; `--read-only` rootfs (несовместим с docker cp); volume-mount (path-mangling на Windows); gVisor/Firecracker (избыточно для MVP).
- **D-026 — Контент адресуется по id ИЛИ slug (CH15.5): `GET /articles/{ref}` и `/problems/{ref}` (включая под-роуты read/complete и submissions/solution) принимают как UUID, так и slug.** `track_items` ссылаются на контент полиморфно по `content_id` (UUID, D-006), тогда как detail-страницы статей/задач на фронте маршрутизируются по slug (человекочитаемые URL). Чтобы переходы из программы трека работали единообразно для всех типов, `content.GetArticleBySlug`/`GetProblemBySlug` и progress-резолверы `resolveArticle`/`resolveProblem` определяют UUID через `pgxutil.ParseUUID` → lookup по id (`store.GetArticleByID`/`GetProblemByID`), иначе по slug. Видео/квизы уже маршрутизируются по UUID. Ресурс становится адресуемым и стабильным id, и slug (RESTful). By-id lookup реализован **hand-written** методами store (`internal/store/content_byid.go` — тот же пакет, читают `q.db`, мирроринг стиля sqlc) — чтобы не дёргать перегенерацию sqlc (требует переключения go-тулчейна на 1.26.4); отдельный файл sqlc-генерацию не трогает. Фронт-страницы статей/задач остаются slug-маршрутизируемыми, но раскрывают навигацию из трека по `content_id`. *Статус:* **принято 2026-06-25.** *Альтернатива:* обогащать track detail title+slug на каждый элемент (joins по 5 типам); резолвить id→slug на фронте через list-эндпоинты (ненадёжно при пагинации); перегенерировать sqlc; полностью id-маршрутизируемые detail (теряются slug-URL).
- **D-025 — Статьи на фронте (CH15.2): `react-markdown` без `rehype-raw` как слой санитизации (сырой HTML не рендерится); `rehype-highlight` с узким набором грамматик; Markdown лениво грузится отдельным чанком.** DoD требует санитизацию Markdown. `react-markdown` по умолчанию НЕ парсит встроенный HTML (плагин `rehype-raw` не подключён) — любой `<script>`/`<img onerror>` отдаётся как инертный текст, а дефолтный URL-transform отбрасывает опасные схемы (`javascript:`); это и есть защита от XSS, без отдельного `rehype-sanitize`, который иначе пришлось бы настраивать так, чтобы не срезать `language-*` классы, нужные `rehype-highlight`. Подсветка — `rehype-highlight` только с грамматиками `go/bash/json/yaml/sql/dockerfile` (вместо ~37 языков common-набора highlight.js). Тяжёлые зависимости (`react-markdown` + `highlight.js`) изолированы в ленивый чанк через `React.lazy`/`Suspense` в `ArticleDetail` → основной бандл 391 kB, Markdown-чанк 323 kB (вместо общего 714 kB, без chunk-warning). `components/Markdown` задаёт кастомные рендереры (`h1-3/p/a` с `target=_blank rel=noopener`/списки/таблицы/blockquote/inline-code) и кнопку «Открыть в песочнице» на код-блоках — заглушка-навигация на `/sandbox` до code-runner (CH16). Бэкенд: добавлен `GET /articles/{slug}/read` (статус прочтения, `found=false` если статья есть, но не прочитана) рядом с `POST /articles/{slug}/complete`. *Статус:* **принято 2026-06-25.** *Альтернатива:* `rehype-sanitize` с расширенной схемой (сложнее; риск срезать highlight-классы); `shiki` (тяжелее билд/рантайм); рендер всех языков hljs; Markdown в основном бандле без code-split.

---

## 9. Открытые архитектурные вопросы

Активные вопросы — в `STATUS.yaml → open_questions`. На 2026-06-25 открытых архитектурных вопросов нет: D-008 (мультиязычный контент), D-009 (аватары), D-010 (таймзона streak/heatmap) приняты пользователем по рекомендованным вариантам.
