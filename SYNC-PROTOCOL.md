# GoAcademy — SYNC-PROTOCOL

> Протокол синхронизации состояния проекта между чатами (сессиями). Обеспечивает непрерывность работы через SPMS.
> `STATUS.yaml` — единственный источник истины о текущем состоянии. Этот документ описывает, как его читать, валидировать и обновлять.
> Связанные файлы: [STATUS.yaml](./STATUS.yaml), [ROADMAP.md](./ROADMAP.md), [RULES.md](./RULES.md), [EMERGENCY_SYNC.yaml](./EMERGENCY_SYNC.yaml).

---

## 0. Принципы

1. **Единый источник истины** — текущая позиция и прогресс всегда берутся из `STATUS.yaml`, не из памяти чата.
2. **Согласованность** — `STATUS.yaml`, `ROADMAP.md`, `ARCHITECTURE.md`, `DESCRIPTION.md` не должны противоречить друг другу.
3. **Не предполагать молча** — при расхождениях или неопределённости задаётся вопрос пользователю.
4. **DoD прежде перехода** — нельзя двигаться к следующему Stage, не закрыв DoD текущего и не обновив SPMS.

---

## 1. Начало чата (Session Start)

При старте каждой новой сессии ассистент выполняет:

1. **Загрузка SPMS.** Прочитать `STATUS.yaml`, затем `ROADMAP.md`; при необходимости `ARCHITECTURE.md`, `RULES.md`, `DESCRIPTION.md`.
2. **Валидация `STATUS.yaml`** по чек-листу §5. При провале — перейти к §6 (восстановление).
3. **Сверка позиции.** Убедиться, что `current.chapter`/`current.stage` существуют в `ROADMAP.md` и согласованы с `completed`/`in_progress`.
4. **Подтверждение позиции пользователю.** Кратко сообщить:
   - текущая глава/этап и статус;
   - что уже завершено (последние пункты `completed`);
   - `next_actions`;
   - открытые вопросы/блокеры, если есть.
5. **Ожидание подтверждения** (особенно при наличии открытых вопросов или статусе `blocked`).
6. **Продолжение** с `current.stage` после подтверждения.

**Стартовое сообщение (шаблон):**
```
📍 GoAcademy — позиция восстановлена из STATUS.yaml (status_id: <id>)
Текущее: <CHAPTER X> / <STAGE X.Y> — <статус>
Завершено: <N>/<total> этапов (<%>)
Следующие действия: <next_actions[0..]>
Открытые вопросы/блокеры: <... или "нет">
Подтвердите продолжение или дайте уточнения.
```

---

## 2. Работа в течение чата

- **По завершении Stage:**
  1. Проверить DoD этапа (ROADMAP + общий DoD).
  2. Создать новую версию `STATUS.yaml`:
     - новый `status_id` (формат `YYYYMMDD-HHMMSS`);
     - добавить запись в `completed` (`stage_id`, `title`, `completed_date`, `evidence` — ссылки на код/тесты/файлы);
     - сдвинуть `current` на следующий Stage, `status: in_progress`;
     - обновить `in_progress`, `next_actions`, `metrics`;
     - при необходимости — `decisions`.
  3. Валидировать `STATUS.yaml` по §5.
- **При архитектурном решении:** записать в `decisions` (`STATUS.yaml`) и в `ARCHITECTURE.md` (присвоить `D-NNN`).
- **При блокере:** см. §3.
- **При изменении плана:** обновить `ROADMAP.md` (ID этапов не переиспользуются), отразить в `metrics`.

---

## 3. Обработка блокеров

1. Установить `current.status: blocked`.
2. Добавить запись в `open_questions` (`question`, `priority`, `context`).
3. Обновить `EMERGENCY_SYNC.yaml → open_blockers`.
4. **Не продолжать** заблокированный Stage до разрешения. Сформулировать пользователю конкретный вопрос/варианты.
5. После разрешения: при необходимости — `decision`, снять `blocked` → `in_progress`, убрать вопрос из `open_questions`.

---

## 4. Завершение чата (Session End)

Перед завершением сессии ассистент:
1. Приводит `STATUS.yaml` к финальному актуальному состоянию (новый `status_id`).
2. Обновляет `EMERGENCY_SYNC.yaml` (`last_known_stage`, `critical_info`, `open_blockers`, `last_update`).
3. Даёт **summary**: что сделано в сессии, текущая позиция, `next_actions`, открытые вопросы.

**Summary (шаблон):**
```
✅ Сессия завершена.
Сделано: <...>
Позиция: <CHAPTER X / STAGE X.Y — статус>
Прогресс: <N>/<total> (<%>)
Дальше: <next_actions>
Открытые вопросы: <...>
```

---

## 5. Валидация STATUS.yaml (чек-лист)

`STATUS.yaml` считается валидным, если:
- [ ] `status_id` уникален и в формате `YYYYMMDD-HHMMSS`.
- [ ] `project: "GoAcademy"`.
- [ ] `current.chapter` и `current.stage` существуют в `ROADMAP.md`.
- [ ] `current.status` ∈ {`in_progress`, `completed`, `blocked`}.
- [ ] Каждый элемент `completed` имеет `stage_id`, `title`, `completed_date` (YYYY-MM-DD), `evidence`.
- [ ] `stage_id` в `completed` не дублируются и существуют в ROADMAP.
- [ ] `in_progress.stage_id` совпадает с `current.stage` (если статус не `completed`-всего-проекта).
- [ ] `metrics.completed_stages == len(completed)`.
- [ ] `metrics.total_stages == 57` (или согласовано с ROADMAP при изменении плана).
- [ ] `metrics.completion_percentage == round(completed/total*100)`.
- [ ] `metrics.estimated_remaining_stages == total_stages - completed_stages`.
- [ ] Каждый `open_questions[].priority` ∈ {`critical`, `high`, `medium`, `low`}.
- [ ] Если `current.status == blocked`, то `open_questions` непуст.
- [ ] `tech_stack` имеет ключи `backend`, `frontend`, `database`, `infra`.
- [ ] YAML синтаксически корректен.

При невыполнении пункта — исправить перед продолжением; при невозможности восстановить согласованность — §6.

---

## 6. Аварийное восстановление (Emergency Recovery)

Если `STATUS.yaml` потерян/повреждён/противоречив:
1. Прочитать `EMERGENCY_SYNC.yaml` → взять `last_known_stage`, `critical_info`, `open_blockers`.
2. Сверить с `ROADMAP.md` и фактическим состоянием кода (что реально реализовано) — определить ближайший достоверный Stage.
3. Пересобрать `STATUS.yaml` из этих данных, выставить консервативную позицию (не завышать прогресс).
4. Подтвердить восстановленную позицию у пользователя перед продолжением.
5. Зафиксировать факт восстановления в `notes`.

---

## 7. Согласованность файлов SPMS

- `ROADMAP.md` ↔ `STATUS.yaml`: позиции и количество этапов согласованы.
- `ARCHITECTURE.md` ↔ `STATUS.yaml.decisions`: каждое `D-NNN` присутствует в обоих.
- `DESCRIPTION.md` ↔ промт: продуктовые требования не расходятся.
- `RULES.md`: правила обновления SPMS соблюдаются (RULES §11).
- При любом расхождении — синхронизировать и отметить в `notes`.

---

## 8. Формат идентификаторов

- `status_id`: `YYYYMMDD-HHMMSS` (напр. `20260625-143000`).
- Stage ID: `STAGE X.Y` (как в ROADMAP, уникальны, не переиспользуются).
- Chapter ID: `CHAPTER X`.
- Decision ID: `D-NNN` (трёхзначный, монотонный).
