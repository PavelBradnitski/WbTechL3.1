package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/statuscache"
	"github.com/wb-go/wbf/rabbitmq"
)

// Scheduler определяет интерфейс для планировщика уведомлений.
type Scheduler interface {
	Start()
	Stop()
}

// NotificationScheduler отвечает за периодическую проверку базы данных на наличие новых уведомлений
type NotificationScheduler struct {
	svc         NotificationService
	publisher   *rabbitmq.Publisher
	statusCache *statuscache.Cache
	queueName   string
	interval    time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewNotificationScheduler создает новый экземпляр NotificationScheduler.
func NewNotificationScheduler(svc NotificationService, conn *rabbitmq.Connection, statusCache *statuscache.Cache, queueName string, interval time.Duration) (*NotificationScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	ch, err := conn.Channel()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	exchange := rabbitmq.NewExchange("jobs.exchange", "direct")
	exchange.Durable = true
	if err := exchange.BindToChannel(ch); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	queueManager := rabbitmq.NewQueueManager(ch)
	_, err = queueManager.DeclareQueue(queueName, rabbitmq.QueueConfig{Durable: true})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := ch.QueueBind(queueName, queueName, exchange.Name(), false, nil); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// publisher
	pub := rabbitmq.NewPublisher(ch, exchange.Name())

	return &NotificationScheduler{
		svc:         svc,
		publisher:   pub,
		statusCache: statusCache,
		queueName:   queueName,
		interval:    interval,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start запускает планировщик уведомлений.
func (s *NotificationScheduler) Start() {
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.processPending()
			case <-s.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop останавливает планировщик уведомлений.
func (s *NotificationScheduler) Stop() {
	s.cancel()
}

func (s *NotificationScheduler) processPending() {
	// резервируем пачку уведомлений
	notifications, err := s.svc.ReservePending(s.ctx, 50)
	if err != nil {
		log.Println("scheduler: failed to reserve pending notifications:", err)
		return
	}
	if len(notifications) == 0 {
		return
	}

	for _, n := range notifications {
		log.Printf("scheduler: sending notification %v to queue", n.ID)

		body, err := json.Marshal(n)
		if err != nil {
			log.Printf("failed to marshal notification %v: %v", n.ID, err)
			continue
		}

		// сразу публикуем
		err = s.publisher.Publish(
			body,
			s.queueName,        // routing key
			"application/json", // content type
		)
		if err != nil {
			log.Printf("scheduler: failed to publish notification %v: %v", n.ID, err)
			continue
		}

		// сохраняем статус в Redis
		if err := s.statusCache.SetStatus(s.ctx, n.ID, models.StatusProcessing); err != nil {
			log.Printf("failed to set status in redis for id=%v: %v", n.ID, err)
		}
	}
}
