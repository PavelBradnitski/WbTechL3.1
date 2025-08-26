package main

import (
	"fmt"
	"os"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/queue"
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

	q, err := queue.NewRabbitMQ(url, queueName)
	if err != nil {
		panic(err)
	}

	fmt.Println("Worker started...")

	err = q.Consume(func(n *models.Notification) error {
		now := time.Now().UTC()
		fmt.Println("Received notification:", n.ID, "Scheduled at:", n.SendAt, "Now:", now)
		if n.SendAt.After(now) {
			// пока рано, вернём в очередь
			return fmt.Errorf("too early, retry later")
		}

		// имитация отправки
		fmt.Printf("Sending notification %s: %s\n", n.ID, n.Message)
		// здесь можно вызвать e-mail/SMS/http клиент

		return nil
	})

	if err != nil {
		panic(err)
	}
}
