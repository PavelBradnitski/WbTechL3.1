package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/sender"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

// Worker отвечает за получение сообщений из очереди и обработку уведомлений.
type Worker struct {
	channel *rabbitmq.Channel
	sender  sender.Sender
	service NotificationService
}

// NewWorker создает новый экземпляр Worker.
func NewWorker(channel *rabbitmq.Channel, sender sender.Sender, svc NotificationService) *Worker {
	return &Worker{channel: channel, sender: sender, service: svc}
}

// Start запускает обработку сообщений из очереди.
func (w *Worker) Start() {
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "notifications"
	}
	conf := &rabbitmq.ConsumerConfig{
		Queue:    queueName,
		Consumer: "worker-1",
		AutoAck:  false,
	}
	msgs, err := w.channel.Consume(
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
			if n.Type != "email" && n.Type != "telegram" {
				log.Printf("unsupported notification type: %s", n.Type)
				// помечаем как failed в БД, чтобы не брать в работу
				if err := w.service.UpdateStatus(ctx, n.ID, models.StatusFailed); err != nil {
					log.Printf("failed to update status for id=%v: %v", n.ID, err)
				}
				// удаляем из очереди, чтобы не зацикливать
				if err := d.Ack(false); err != nil {
					log.Printf("failed to ack message id=%v: %v", n.ID, err)
				}
				return nil
			}
			switch n.Type {
			case "email":
				// Отправляем email
				if err := w.sender.Send(&n); err != nil {
					log.Printf("failed to send email: %v", err)
					if err := w.service.IncrementRetries(ctx, n.ID); err != nil {
						log.Printf("failed to increment retries for id=%v: %v", n.ID, err)
					}
					return err
				}

			case "telegram":
				// Отправляем telegram
				if err := w.sender.Send(&n); err != nil {
					log.Printf("failed to send telegram: %v", err)
					if err := w.service.IncrementRetries(ctx, n.ID); err != nil {
						log.Printf("failed to increment retries for id=%v: %v", n.ID, err)
					}
					return err
				}
			}
			log.Printf("sent notification id=%v message=%q", n.ID, n.TelegramNotification.Message)
			return nil
		}, strat)
		if err != nil {
			// все попытки исчерпаны
			log.Printf("processing failed for id=%v after retries: %v", n.ID, err)
			// помечаем как failed в БД, чтобы не брать в работу
			w.service.UpdateStatus(ctx, n.ID, models.StatusFailed)
			// удаляем из очереди, чтобы не зацикливать
			if err := d.Nack(false, false); err != nil {
				log.Printf("failed to nack message id=%v: %v", n.ID, err)
			}
			continue
		}

		if err := w.service.UpdateStatus(ctx, n.ID, models.StatusSent); err != nil {
			log.Printf("failed to mark notification %v as sent: %v", n.ID, err)
		}
		// подтверждаем успешную обработку
		if err := d.Ack(false); err != nil {
			log.Printf("failed to ack message id=%v: %v", n.ID, err)
		}
	}
}
