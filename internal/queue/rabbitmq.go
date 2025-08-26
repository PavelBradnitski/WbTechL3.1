package queue

import (
	"encoding/json"
	"fmt"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationQueue interface {
	Publish(n *models.Notification) error
	Consume(handler func(n *models.Notification) error) error
}

type rabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewRabbitMQ(url, queueName string) (NotificationQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &rabbitMQ{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (r *rabbitMQ) Publish(n *models.Notification) error {
	body, err := json.Marshal(n)
	if err != nil {
		return err
	}
	return r.channel.Publish(
		"",      // default exchange
		r.queue, // routing key = queue
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *rabbitMQ) Consume(handler func(n *models.Notification) error) error {
	msgs, err := r.channel.Consume(
		r.queue,
		"",
		false, // auto-ack disabled
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	for d := range msgs {
		var n models.Notification
		if err := json.Unmarshal(d.Body, &n); err != nil {
			_ = d.Nack(false, false) // drop bad message
			continue
		}
		if err := handler(&n); err != nil {
			_ = d.Nack(false, true) // retry later
			continue
		}
		_ = d.Ack(false)
	}
	return nil
}
