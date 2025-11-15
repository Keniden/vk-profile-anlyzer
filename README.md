# INTEAM — анализатор профиля VK (Go)

Сервис для анализа профиля пользователя ВКонтакте. Бэкенд на Go подтягивает данные из VK API (стена, друзья, подарки), строит вектор активности, отправляет их в GigaChat для генерации текстового резюме и сохраняет результат в базе и объектном хранилище.

---

## Функциональность :rocket:

- Авторизация через email/пароль и через VK OAuth.
- Анализ профиля по VK ID: сбор данных из VK API и генерация краткого описания с помощью GigaChat.
- Сохранение профиля и результата анализа в БД, возможность повторного чтения без повторных запросов к VK.
- Кэширование ответов VK через Redis и сохранение «снапшотов» профиля в Minio.
- Метрики Prometheus (`/metrics`) и трассировки OpenTelemetry.
- gRPC‑сервер со стандартным health‑чеком (порт `9090`) для интеграции с оркестраторами.

## Архитектура и стек :computer:

**Backend**

- Go
- `github.com/gin-gonic/gin` — HTTP API.
- GORM (Postgres / SQLite) — хранение пользователей и профилей.
- VK API — данные пользователя: стена, друзья, подарки.
- GigaChat — генерация текстового резюме профиля на основе собранных данных.
- Redis — кэширование запросов к VK.
- Minio (S3‑совместимое хранилище) — сохранение сырых JSON‑снапшотов профиля.
- JWT‑аутентификация, VK OAuth, Prometheus‑метрики, OpenTelemetry‑трейсинг.
- Отдельный gRPC‑health‑сервер для проверки живости сервиса.

**Frontend**

- Простой SPA на «голом» HTML/CSS/JS (`internal/frontend`), отдаётся статикой по пути `/static`.
- Форма, куда можно ввести VK ID и отправить запрос на анализ профиля.

## Основные эндпоинты HTTP API :clipboard:

- `GET /healthz` — health‑check сервиса.

**Авторизация**

- `POST /auth/login` — логин по email/паролю, возвращает `access_token` и `refresh_token` (и, опционально, устанавливает их в куки).
- `GET /auth/vk/login` — редирект на VK OAuth.
- `GET /auth/vk/callback` — callback от VK; создаёт пользователя (если его ещё нет) и возвращает пару JWT‑токенов.

**Защищённые эндпоинты** (требуется JWT, мидлварь `auth.JWTMiddleware`):

- `GET /me` — информация о текущем пользователе.
- `GET /profiles/{vk_id}` — получить сохранённый профиль.
- `POST /profiles/{vk_id}/analyze` — инициировать анализ профиля VK и сохранить/обновить результат.

**Метрики**

- `GET /metrics` — Prometheus‑метрики (если включены в конфиге).

## Конфигурация

Конфиг загружается через Viper из файла `config.yaml` (если указан флаг `-config`) и/или из переменных окружения с префиксом `INTEAM_`. Основные параметры:

- `INTEAM_DB_DRIVER` — драйвер БД: `postgres` или `sqlite`.
- `INTEAM_DB_DSN` — DSN для подключения (например, `host=db user=postgres password=postgres dbname=inteam sslmode=disable` или путь к файлу SQLite).
- `INTEAM_VK_BASE_URL` — базовый URL VK API (например, `https://api.vk.com/method`).
- `INTEAM_VK_API_VERSION` — версия API VK (например, `5.199`).
- `INTEAM_VK_ACCESS_TOKEN` — сервисный access‑token VK.
- `INTEAM_GIGACHAT_BASE_URL` — URL GigaChat.
- `INTEAM_GIGACHAT_TOKEN` — токен доступа к GigaChat.
- `INTEAM_REDIS_ADDR` — адрес Redis (опционально, если не нужен кэш — можно не задавать).
- `INTEAM_MINIO_ENDPOINT`, `INTEAM_MINIO_ACCESS_KEY_ID`, `INTEAM_MINIO_SECRET_ACCESS_KEY`, `INTEAM_MINIO_BUCKET` — настройки Minio (если не заданы — объектное хранилище отключено).
- `INTEAM_AUTH_JWT_SECRET` — секрет для подписи JWT.
- `INTEAM_AUTH_VK_CLIENT_ID`, `INTEAM_AUTH_VK_CLIENT_SECRET`, `INTEAM_AUTH_VK_REDIRECT_URL` — параметры VK OAuth.
- `INTEAM_HTTP_HOST`, `INTEAM_HTTP_PORT` — адрес и порт HTTP‑сервера (по умолчанию `0.0.0.0:8080`).
- `INTEAM_TELEMETRY_ENABLED`, `INTEAM_TELEMETRY_OTLP_ENDPOINT` — включение и настройки OTEL‑экспортера.

Часть параметров имеет значения по умолчанию (таймауты, включение метрик и т.п.), но драйвер БД, DSN и секркты нужно задать явно.

## Запуск локально

### Вариант 1 — Docker Compose (полный стек)

```bash
docker-compose up --build
```

Это поднимет:

- `api` — Go‑бэкенд на `8080`,
- Postgres,
- Redis,
- Minio,
- Prometheus и Grafana.

После запуска API будет доступен на `http://localhost:8080`, метрики — на `/metrics`. Minio доступен на `http://localhost:9000`, консоль — на `http://localhost:9001` (логин/пароль по умолчанию `minioadmin/minioadmin`).

### Вариант 2 — напрямую через Go

```bash
export INTEAM_DB_DRIVER=sqlite
export INTEAM_DB_DSN=./inteam.db
export INTEAM_VK_BASE_URL=https://api.vk.com/method
export INTEAM_VK_API_VERSION=5.199
export INTEAM_VK_ACCESS_TOKEN=<vk_token>
export INTEAM_GIGACHAT_BASE_URL=<gigachat_url>
export INTEAM_GIGACHAT_TOKEN=<gigachat_token>
export INTEAM_AUTH_JWT_SECRET=<jwt_secret>

go run ./cmd/server
```

При первом старте GORM выполнит `AutoMigrate` для таблиц пользователей и профилей. Миграции в каталоге `migrations` можно использовать отдельно, если нужен контроль над схемой БД.

## Запуск gRPC‑health‑сервера

Для интеграции с Kubernetes/Consul и т.п. есть отдельный бинарь, который поднимает gRPC‑сервер со стандартным health‑сервисом:

```bash
go run ./cmd/grpcserver
```

Сервер слушает порт `9090` и всегда возвращает статус `SERVING` (можно расширить при необходимости).
