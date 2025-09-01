package main

import (
	"log"
	"os"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
)

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
	// создаём планировщик
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	rabbit, err := rabbitmq.Connect(url, 5, 5*time.Second)
	if err != nil {
		panic(err)
	}
	scheduler := service.NewScheduler(svc, rabbit, "notifications", 5*time.Second)
	scheduler.Start()
	defer scheduler.Stop()

	// блокируем main
	select {}
}
