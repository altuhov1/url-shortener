# API

HTTP API сервиса URL Shortener.

- Базовый URL: `http://<host>:${APP_PORT}` (по умолчанию `http://localhost:8080`).
- Все тела запросов и ответов — `application/json; charset=utf-8`.
- Короткий код: ровно 10 символов из алфавита `[a-zA-Z0-9_]`.

## Эндпоинты

| Метод | Путь | Назначение |
|---|---|---|
| POST  | `/shorten`         | Создать сокращённую ссылку (или вернуть существующую) |
| GET   | `/shorten/{short}` | Получить оригинальный URL по короткому коду |

---

### POST `/shorten`

Сокращает оригинальный URL. Идемпотентен: для одного и того же `url` всегда
возвращается один и тот же `short_url`.

**Запрос**

```json
{ "url": "https://example.com/some/long/path?x=1" }
```

**Ответ `200 OK`**

```json
{ "short_url": "aB3_xY9zQp" }
```

**Ошибки**

| Статус | Когда |
|---|---|
| `400 Bad Request` | пустой `url`, отсутствует поле, битый JSON |
| `500 Internal Server Error` | сбой хранилища / непредвиденная ошибка |

Формат ошибки:

```json
{ "error": "url is required" }
```

**Пример**

```bash
curl -s -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/foo"}'
# {"short_url":"aB3_xY9zQp"}
```

---

### GET `/shorten/{short}`

Возвращает оригинальный URL по короткому коду.

**Ответ `200 OK`**

```json
{ "original_url": "https://example.com/some/long/path?x=1" }
```

**Ошибки**

| Статус | Когда |
|---|---|
| `404 Not Found` | код не найден в хранилище |
| `400 Bad Request` | пустой код |
| `500 Internal Server Error` | сбой хранилища |

**Пример**

```bash
curl -s http://localhost:8080/shorten/aB3_xY9zQp
# {"original_url":"https://example.com/foo"}
```

---

## Коды ошибок одной таблицей

| Статус | Сценарий |
|---|---|
| 200 | успех |
| 400 | невалидный ввод (пустой URL, битый JSON) |
| 404 | короткий код не существует |
| 500 | внутренняя ошибка сервиса |
