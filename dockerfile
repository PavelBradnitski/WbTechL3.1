# Сборка бинарника
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o notify-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o notify-worker ./cmd/worker

# Финальный образ
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/notify-api .
COPY --from=builder /app/notify-worker .
CMD ["./notify-api"]
