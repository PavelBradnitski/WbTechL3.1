# ---- builder ----
FROM golang:1.24 AS builder
WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем бинарники
RUN CGO_ENABLED=0 GOOS=linux go build -o notify-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o notify-worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o notify-scheduler ./cmd/scheduler


# ---- runtime ----
FROM alpine:3.20
WORKDIR /app

# Копируем бинарники
COPY --from=builder /app/notify-api .
COPY --from=builder /app/notify-worker .
COPY --from=builder /app/notify-scheduler .


# Копируем миграции для runtime
COPY --from=builder /app/internal/migrations ./migrations

EXPOSE 8081

CMD ["./notify-api"]
