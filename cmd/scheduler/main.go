package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/PavelBradnitski/WbTechL3.1/internal/statuscache"
	"github.com/joho/godotenv"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/redis"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}
}

func main() {
	// читаем параметры подключения
	masterDSN := os.Getenv("POSTGRES_MASTER_DSN")
	if masterDSN == "" {
		masterDSN = "postgres://notify_user:notify_pass@postgres:5432/notifications?sslmode=disable"
	}
	slaveDSNs := []string{}
	if s := os.Getenv("POSTGRES_SLAVE_DSN"); s != "" {
		slaveDSNs = append(slaveDSNs, s)
	}
	// создаём подключение через нашу обёртку
	db, err := dbpg.New(masterDSN, slaveDSNs, &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	// репозиторий
	repo := repository.NewNotificationRepo(db.Master)

	// сервис
	svc := service.NewNotificationService(repo)

	// подключение к Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.New(redisAddr, "", 0)
	// main scheduler
	statusCache := statuscache.New(redisClient)
	ctx := context.Background()
	// проверим соединение
	if err := redisClient.Set(ctx, "scheduler:ping", "ok"); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	// создаём планировщик
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	rabbit, err := rabbitmq.Connect(url, 5, 5*time.Second)
	if err != nil {
		panic(err)
	}
	scheduler, err := service.NewNotificationScheduler(svc, rabbit, statusCache, "notifications", 5*time.Second)
	if err != nil {
		log.Fatalf("failed to create scheduler: %v", err)
	}
	scheduler.Start()
	defer scheduler.Stop()

	// блокируем main
	select {}
}
