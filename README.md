# secunda-task-manager

REST API сервис управления задачами в командах с ролевой моделью, историей изменений и мониторингом.

## Стек

| Компонент       | Технология               |
| --------------- | ------------------------ |
| Язык            | Go 1.25                  |
| HTTP            | Fiber v2                 |
| БД              | MySQL 8 + sqlx (raw SQL) |
| Миграции        | golang-migrate           |
| Кэш             | Redis 7 + go-redis/v9    |
| Конфиг          | caarlos0/env/v11 (ENV)   |
| Логи            | zerolog                  |
| JWT             | golang-jwt/jwt/v5        |
| Circuit Breaker | gobreaker/v2             |
| Метрики         | Prometheus               |
| Тесты           | testcontainers-go        |

## Быстрый старт

```bash
cp .env.example .env          # настроить переменные
docker-compose up --build -d  # поднять MySQL + Redis + app
curl http://localhost:8081/readyz  # проверить готовность
```

## API

### Аутентификация

| Метод | URL                | Описание                              |
| ----- | ------------------ | ------------------------------------- |
| POST  | `/api/v1/register` | Регистрация пользователя              |
| POST  | `/api/v1/login`    | Вход, возвращает access + refresh JWT |
| POST  | `/api/v1/refresh`  | Обновление токенов                    |

### Команды

| Метод | URL                        | Права                                          |
| ----- | -------------------------- | ---------------------------------------------- |
| POST  | `/api/v1/teams/`           | Создать команду (становится owner)             |
| GET   | `/api/v1/teams/`           | Список команд пользователя                     |
| POST  | `/api/v1/teams/:id/invite` | Пригласить пользователя (только owner / admin) |

### Задачи

| Метод | URL                                                                  | Описание                        |
| ----- | -------------------------------------------------------------------- | ------------------------------- |
| POST  | `/api/v1/tasks/`                                                     | Создать задачу (член команды)   |
| GET   | `/api/v1/tasks/?team_id=1&status=todo&assignee_id=5&page=1&limit=20` | Список с фильтрами и пагинацией |
| PUT   | `/api/v1/tasks/:id`                                                  | Обновить задачу                 |
| GET   | `/api/v1/tasks/:id/history`                                          | История изменений               |
| POST  | `/api/v1/tasks/:id/comments`                                         | Добавить комментарий            |
| GET   | `/api/v1/tasks/:id/comments`                                         | Список комментариев             |

### Статистика

| Метод | URL                       | Описание                                                            |
| ----- | ------------------------- | ------------------------------------------------------------------- |
| GET   | `/api/v1/stats/teams`     | Команды: кол-во участников + done-задачи за 7 дней (JOIN 3+ таблиц) |
| GET   | `/api/v1/stats/top-users` | Топ-3 пользователя по задачам в команде за месяц (RANK())           |
| GET   | `/api/v1/stats/integrity` | Задачи с assignee не из команды (NOT EXISTS)                        |

### Private (мониторинг, порт 8081)

| URL        | Описание                             |
| ---------- | ------------------------------------ |
| `/livez`   | Liveness probe                       |
| `/readyz`  | Readiness probe (MySQL + Redis ping) |
| `/metrics` | Prometheus метрики                   |

## Конфигурация (ENV)

```env
SERVER_PORT=8080
SERVER_PRIVATE_PORT=8081
SERVER_LOG_LEVEL=info          # debug / info / warn / error

MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=taskuser
MYSQL_PASSWORD=taskpassword
MYSQL_DBNAME=taskmanager
MYSQL_MAX_OPEN_CONNS=25
MYSQL_MAX_IDLE_CONNS=5
MYSQL_CONN_MAX_LIFETIME=5m
MYSQL_CONN_MAX_IDLE_TIME=1m

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_POOL_SIZE=10

ACCESS_TOKEN_KEY=change_me
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_KEY=change_me
REFRESH_TOKEN_TTL=168h
TOKEN_ISSUER=secunda-task-manager

APP_ALLOWED_ORIGINS=http://localhost:3000
```

## Архитектура

```
cmd/service/main.go      — точка входа
internal/app/app.go      — инициализация зависимостей + graceful shutdown
internal/
  config/                — Load() из ENV
  cache/                 — Redis Client wrapper + версионированный кэш задач
  repository/            — raw SQL через sqlx
  services/              — бизнес-логика (зависит от интерфейсов)
  transport/http/        — Fiber handlers + middleware
pkg/
  jwt/                   — Generate / Validate
  logger/                — zerolog wrapper
migrations/              — SQL-файлы (001..005)
tests/integration/       — тесты репозитория через testcontainers
```

### Ключевые решения

- **DI вручную** в `app.go`: нет фреймворков, зависимости видны явно
- **Трёхслойная архитектура**: handler → service (интерфейс) → repository (интерфейс)
- **Транзакции** передаются явно: создание команды = команда + owner в одной транзакции; обновление задачи = update + history в одной транзакции
- **Graceful shutdown**: SIGTERM → drain HTTP (15s) → close Redis → close MySQL (LIFO defers)
- **Rate limiting**: Redis INCR + EXPIRE, 100 req/min на пользователя
- **Circuit breaker** для email-сервиса: 5 последовательных ошибок → open, сброс через 60s
- **История изменений**: дельта-подход — одна строка на изменённое поле, хранятся старое и новое значение

## Тестирование

```bash
# Unit-тесты (без Docker)
go test ./internal/...

# Интеграционные тесты (требует Docker)
go test -tags integration ./tests/integration/... -v
```

Интеграционные тесты поднимают MySQL-контейнер через testcontainers, прогоняют миграции и тестируют repository-слой напрямую. Один контейнер на весь прогон — каждая суита очищает все таблицы в `SetupSuite` и стартует с чистой базой.

---

## Вопросы и допущения

### Кэширование списка задач

В требованиях указано кэшировать «список задач команды (TTL 5 мин)». Однако единственный эндпоинт для получения задач (`GET /tasks`) принимает фильтры: статус, приоритет, assignee, страница. Кэшировать один «список команды» здесь не получится — каждая комбинация фильтров даёт отдельный результат.

Реализован **версионированный кэш**: ключ включает `team_id`, версию команды и набор фильтров. При любом изменении задачи версия инкрементируется — все кэшированные списки этой команды становятся невалидными. Стратегия зависит от типа нагрузки (чтение/запись) — возможно, нужен другой подход.

Помимо задач, имеет смысл кэшировать членство пользователей в командах и их роли, а также сами команды. Эти данные меняются редко, но участвуют почти в каждом запросе (проверка прав, проверка членства). Для них подходит стратегия **write-through**: запись идёт одновременно в БД и кэш, чтение — всегда из кэша, в базу за этими данными не ходим. Кэш всегда актуален, нет cache miss на горячем пути.

### Проверка прав на чтение

В требованиях для эндпоинтов:

- `GET /tasks?team_id=1&...` — список задач с фильтрацией
- `GET /tasks/:id/history` — история изменений

проверка прав доступа явно не была прописана. Поскольку `team_id` передаётся параметром запроса, а не берётся из токена, добавлять неявную проверку «является ли пользователь членом команды» без уточнения требований не стал. Если такая проверка нужна — добавить несложно.

### Пагинация

Требование «пагинация на уровне БД» реализовано через LIMIT/OFFSET в запросе на список задач. На остальные эндпоинты (история изменений, комментарии, статистика) пагинация не добавлялась — в требованиях она не упоминалась, а усложнять без необходимости не стал.
