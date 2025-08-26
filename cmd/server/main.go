package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/PavelBradnitski/WbTechL3.1/internal/handler"
	"github.com/PavelBradnitski/WbTechL3.1/internal/queue"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
)

func main() {
	repo := repository.NewMemoryRepo()

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

	svc := service.NewNotificationService(repo, q)
	h := handler.NewNotificationHandler(svc)

	fmt.Println("API server started on :8080")
	if err := http.ListenAndServe(":8080", h.Routes()); err != nil {
		panic(err)
	}
}
