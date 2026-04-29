# URL Shortener

Сервис сокращения ссылок на Go.

Принимает оригинальный URL по HTTP, выдаёт уникальный 10-символьный код (`[a-zA-Z0-9_]`) и резолвит его обратно в оригинальный URL. Хранилище выбирается при запуске через env: PostgreSQL или in-memory.


## Алгоритм генерации

`internal/services/generator.go` — 10 случайных символов из алфавита
`[a-zA-Z0-9_]` (63 символа, `math/rand/v2`). Сервис делает до 5 попыток при коллизии короткого URL и корректно обрабатывает гонку, когда параллельный запрос успел вставить тот же оригинал.


## Структура проекта

Архитектура трёхуровневая, слои общаются через интерфейсы:
`handlers.URLService` → `services.URLService` → `services.Storage`. Это позволяет подменять реализации в юнит-тестах через моки.

## Конфигурация

Все параметры читаются из переменных окружения (через `caarlos0/env`),
поддерживается `.env` в корне проекта (через `joho/godotenv`).

| Переменная | Назначение | Пример |
|---|---|---|
| `IS_BD_IN_MEMORY` | `1` — in-memory, `0` — PostgreSQL | `0` |
| `APP_PORT` | порт HTTP-сервиса | `8080` |
| `LOGS_LEVEL_APP` | уровень логов сервиса | `INFO` |
| `LOGS_LEVEL_MIGRATE` | уровень логов миграций | `INFO` |
| `PG_HOST` | хост PostgreSQL | `postgres` |
| `PG_PORT` | порт PostgreSQL | `5432` |
| `PG_SSLMODE` | sslmode | `disable` |
| `STANDART_PG_USER` | суперюзер БД (для миграций) | `st_user` |
| `STANDART_PG_PASSWORD` | пароль суперюзера | `st_user` |
| `STANDART_PG_DB_NAME` | имя БД | `name_db` |
| `PG_USERNAME_FOR_APP` | пользователь, под которым работает сервис | `backendApp` |
| `PG_USERPASS_FOR_APP` | пароль этого пользователя | `secretPassword` |

Пример `.env` лежит в репозитории, в реальном проекте он должен быть в гит игноре

## API

Описание HTTP API — в [API.md](./API.md).

## Запуск

### Через Docker Compose

```bash
docker compose up --build
```

Загрузка хвоста логов
```bash
make docker-logs    # хвост логов
```

## Тесты

```bash
make test           # быстрый прогон
make test-cover     # coverage-отчёт
```

## Линтинг

```bash
make lint           # golangci-lint run ./...
```

Конфиг линтера — `.golangci.yml` 

## Makefile — все цели

| Цель | Действие |
|---|---|
| `make all` | `lint` + `test` + `build` |
| `make test` | `go test ./...` |
| `make test-cover` | coverage-отчёт |
| `make lint` | `golangci-lint run ./...` |
| `make docker-logs` | `docker compose logs -f` |
