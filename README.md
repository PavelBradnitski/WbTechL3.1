# WbTechL3.1
DelayedNotifier — отложенные уведомления через очереди
## Описание

**Delayed Notifier** — сервис для отложенной отправки Email и Telegram уведомлений. Позволяет создавать уведомления, которые будут отправлены в заданное время. Использует PostgreSQL, RabbitMQ и Email(SMTP).

---

## Архитектура

- **API (cmd/delayed-notifier):** HTTP-сервер, принимает запросы на создание, получение и удаление уведомлений.
- **Worker (cmd/worker):** Фоновый воркер, который принимает уведомления из RabbitMQ, готовые к отправке, и отправляет email через SMTP или в Telegram.
- **Scheduler(cmd/scheduler):** Планировщик, который ищет в БД уведомления, готовые для отправки, и посылает их в RabbitMQ.
- **PostgreSQL:** Хранит уведомления.
- **Mailhog (SMTP):** Тестирование отправки email-сообщений.

## Архитектура проекта
```
├── cmd/           # Основные исполняемые приложения
│  ├── scheduler/    # Планировщик задач
│  │  └── main.go      # Точка входа приложения планировщика
│  ├── server/       # HTTP API сервер для управления уведомлениями
│  │  └── main.go      # Точка входа API сервера
│  └── worker/       # Воркер для обработки и отправки уведомлений
│    └── main.go       # Точка входа воркера
├── frontend/      # Фронтенд (веб-интерфейс)
│  └── index.html    # Основная HTML страница фронтенда
├── internal/      # Внутренние пакеты (не предназначены для внешнего использования)
│  ├── db/           # Работа с базой данных
│  │  └── init/        # Инициализация базы данных
│  │    └── init.sql     # Настройка пользователя и схемы БД
│  ├── handler/      # HTTP обработчики (handlers) для API endpoints
│  │  └── notification_handler.go # Обработчики для работы с уведомлениями
│  ├── migrations/   # Миграции базы данных
│  │  ├── 0001_init.down.sql   # Миграция для отката изменений (удаление таблиц, данных)
│  │  └── 0001_init.up.sql     # Миграция для создания таблиц и начальной инициализации
│  ├── models/       # Модели данных (структуры Go)
│  │  └── notification.go      # Структура, представляющая уведомление (Notification)
│  ├── repository/   # Уровень доступа к данным (Data Access Layer - DAL)
│  │  └── notification_repo.go # Методы для работы с уведомлениями в базе данных (CRUD операции)
│  ├── sender/       # Пакет для отправки уведомлений различными способами
│  │  ├── email_sender.go      # Реализация отправки уведомлений по электронной почте
│  │  ├── multisender.go       # Комбинирует несколько отправителей (например, email и telegram)
│  │  ├── sender.go            # Интерфейс для отправителей уведомлений
│  │  └── telegram_sender.go   # Реализация отправки уведомлений через Telegram
│  └── service/      # Бизнес-логика приложения (Services)
│    ├── notification_service.go # Логика управления уведомлениями (создание, отправка, получение)
│    ├── scheduler.go          # Логика планирования отправки уведомлений
│    └── worker.go             # Логика обработки уведомлений (получение, форматирование, отправка)
├── docker-compose.yml  # Конфигурация Docker Compose для локального развертывания
├── dockerfile          # Dockerfile для сборки основного приложения
├── frontend.Dockerfile # Dockerfile для сборки фронтенда
├── go.mod              # Файл управления зависимостями Go (модуль)
├── go.sum              # Файл с контрольными суммами зависимостей Go
└── README.md           # описание проекта
```

---

## Быстрый старт

### Запуск через Docker Compose

```bash
git clone <repo-url>
cd WbTechL3.1
# Запустите инфраструктуру и сборку контейнеров
docker compose up --build
```

- API будет доступен на `http://localhost:8081`
- фронтенд будет доступен на `http://localhost:8080`
- Mailhog будет доступен на `http://localhost:8025` 
- PostgreSQL: порт 5672

### Переменные окружения

Скопируйте файл `env.example` в `.env` и настройте переменные под ваше окружение:


## Примеры HTTP-запросов

### Создать уведомление

**Пример с email**
```bash
curl -X POST http://localhost:8081/notify \
  -H 'Content-Type: application/json' \
  -d '{
  "user_id": "",
  "email": "user123@example.com",
  "type": "email",
  "message": "Привет! Это ваше запланированное уведомление через почту.",
  "subject": "Запланированное уведомление",
  "scheduled_at": "2024-11-10T10:00:00Z"
}'
```
**Пример с telegram**
```bash
curl -X POST http://localhost:8081/notify \
  -H 'Content-Type: application/json' \
  -d '{
  "user_id": "user123",
  "email": "",
  "type": "telegram",
  "message": "Привет! Это ваше запланированное уведомление через телеграм.",
  "subject": "",
  "scheduled_at": "2024-11-10T10:00:00Z"
}'
```
**Ответ:**
```json
{
  "id": "<uuid>",
}
```

### Получить уведомление

```bash
curl http://localhost:8080/notify/<id>
```
**Ответ:**
```json
{
  "id": "notification456",
  "user_id": "user456",
  "type": "telegram",
  "message": "Привет из Telegram! Уведомление успешно отправлено.",
  "scheduled_at": "2024-11-08T08:00:00Z",
  "status": "sent",
  "created_at": "2024-11-08T07:50:00Z",
  "updated_at": "2024-11-08T08:01:00Z"
}

```

### Отменить уведомление

```bash
curl -X DELETE http://localhost:8080/notify/<id>
```
**Ответ:** HTTP 204 No Content

---

## Формат уведомления

```json
{
  "id": "string (uuid)",
  "send_at": "RFC3339 datetime",
  "message": "string",
  "status": "scheduled|processing|sent|failed|canceled",
  "email": "string",
  "user_id": "string"
}
```


