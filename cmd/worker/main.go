package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

func main() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "notifications"
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

	conf := &rabbitmq.ConsumerConfig{
		Queue:    queueName,
		Consumer: "worker-1",
		AutoAck:  false,
		// остальные поля можно оставить по умолчанию
	}

	msgs, err := channel.Consume(
		conf.Queue,
		conf.Consumer,
		conf.AutoAck,
		conf.Exclusive,
		conf.NoLocal,
		conf.NoWait,
		conf.Args,
	)
	if err != nil {
		log.Fatalf("consume: %v", err)
	}

	ctx := context.Background()
	log.Println("worker started, waiting for messages...")
	for d := range msgs {
		log.Println("message received, processing...")

		var n models.Notification
		if err := json.Unmarshal(d.Body, &n); err != nil {
			log.Printf("invalid message: %v", err)
			// невалидный формат — удаляем сообщение из очереди (ack), чтобы не повторять
			if err := d.Ack(false); err != nil {
				log.Printf("failed to ack invalid message: %v", err)
			}
			continue
		}

		strat := retry.Strategy{
			Attempts: 3,
			Delay:    time.Second,
			Backoff:  2,
		}
		log.Printf("received: %v", n)
		err := retry.Do(func() error {
			// здесь будет реальная логика отправки уведомления
			//допустим, с вероятностью 50% имитируем ошибку
			if rand.Intn(2) == 0 {
				log.Printf("send failed, will retry... %v\n", n)
				return fmt.Errorf("send failed")
			}
			log.Printf("sent notification id=%v message=%q", n.ID, n.Message)
			return nil
		}, strat)
		if err != nil {
			// обработка провалена после всех попыток
			log.Printf("processing failed for id=%v: %v", n.ID, err)
			// обновляем retries в сервисе (чтобы scheduler мог переотправлять/рассматривать)
			if err := svc.IncrementRetries(ctx, n.ID); err != nil {
				log.Printf("failed to increment retries for id=%v: %v", n.ID, err)
			}
			// удаляем сообщение из очереди — не будем зацикливать его в Rabbit.
			if err := d.Ack(false); err != nil {
				log.Printf("failed to ack failed message id=%v: %v", n.ID, err)
			}
			continue
		}

		if err := svc.MarkAsSent(ctx, n.ID); err != nil {
			log.Printf("failed to mark notification %v as sent: %v", n.ID, err)
		}
		if err := d.Ack(false); err != nil {
			log.Printf("failed to ack message id=%v: %v", n.ID, err)
		}
	}
}
