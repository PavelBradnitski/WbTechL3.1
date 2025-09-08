package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/notifications"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
)

func main() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
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

	fmt.Println("Worker started...")
	rabbit, err := rabbitmq.Connect(url, 5, 5*time.Second)
	if err != nil {
		panic(err)
	}
	defer rabbit.Close()
	channel, err := rabbit.Channel()
	if err != nil {
		panic(err)
	}
	defer channel.Close()
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	emailSender := notifications.NewEmailSender(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
		os.Getenv("SMTP_FROM"),
	)
	telegramSender := notifications.NewTelegramSender(os.Getenv("TELEGRAM_TOKEN"))

	sender := notifications.NewMultiSender(emailSender, telegramSender)

	worker := service.NewWorker(channel, sender, svc)
	worker.Start()
}
