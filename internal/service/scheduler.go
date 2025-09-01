// internal/service/scheduler.go
package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	//amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
)

type Scheduler interface {
	Start()
	Stop()
}

type NotificationScheduler struct {
	svc        NotificationService
	rabbitConn *rabbitmq.Connection
	queueName  string
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewScheduler(svc NotificationService, conn *rabbitmq.Connection, queueName string, interval time.Duration) *NotificationScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationScheduler{
		svc:        svc,
		rabbitConn: conn,
		queueName:  queueName,
		interval:   interval,
		ctx:        ctx,
		cancel:     cancel,
	}
}

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

func (s *NotificationScheduler) Stop() {
	s.cancel()
}

func (s *NotificationScheduler) processPending() {
	// получаем список уведомлений
	notifications, err := s.svc.GetPendingNotifications(s.ctx, 50)
	if err != nil {
		log.Println("scheduler: failed to fetch pending notifications:", err)
		return
	}
	if len(notifications) == 0 {
		return
	}

	// открываем канал
	ch, err := s.rabbitConn.Channel()
	if err != nil {
		log.Println("scheduler: failed to open rabbit channel:", err)
		return
	}
	defer ch.Close()

	// 1. Объявляем exchange
	ex := rabbitmq.NewExchange("jobs.exchange", "direct")
	ex.Durable = true
	if err := ex.BindToChannel(ch); err != nil {
		log.Println("scheduler: failed to declare exchange:", err)
		return
	}

	// 2. Объявляем очередь
	qm := rabbitmq.NewQueueManager(ch)
	_, err = qm.DeclareQueue(s.queueName, rabbitmq.QueueConfig{Durable: true})
	if err != nil {
		log.Println("scheduler: failed to declare queue:", err)
		return
	}

	// 3. Биндим очередь к exchange
	if err := ch.QueueBind(s.queueName, s.queueName, ex.Name(), false, nil); err != nil {
		log.Println("scheduler: failed to bind queue:", err)
		return
	}

	// 4. Паблишер
	pub := rabbitmq.NewPublisher(ch, ex.Name())

	for _, n := range notifications {
		log.Printf("scheduler: processing notification %v", n.ID)
		body, err := json.Marshal(n) // сериализация всей структуры
		if err != nil {
			log.Printf("failed to marshal notification %v: %v", n.ID, err)
			continue
		}

		// msgBody := []byte(n.Message) // или json.Marshal(n)

		err = pub.Publish(
			body,
			s.queueName,        // routing key = имя очереди
			"application/json", // content type
		)
		if err != nil {
			log.Printf("scheduler: failed to publish notification %v: %v", n.ID, err)
			//_ = s.svc.IncrementRetries(s.ctx, n.ID)
			continue
		}

		// if err := s.svc.MarkAsSent(s.ctx, n.ID); err != nil {
		// 	log.Printf("scheduler: failed to mark notification %v as sent: %v", n.ID, err)
		// }
	}
}
