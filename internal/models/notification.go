package models

import "time"

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusSent      Status = "sent"
	StatusCanceled  Status = "canceled"
	StatusFailed    Status = "failed"
)

// Модель для БД (внутренняя)
type Notification struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Message     string    `db:"message"`
	ScheduledAt time.Time `db:"scheduled_at"`
	Status      Status    `db:"status"`
	Retries     int       `db:"retries"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// DTO для входящих запросов (например, POST /notifications)
type CreateNotificationRequest struct {
	UserID      string    `json:"user_id"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// DTO для ответа API (можно не отдавать retries, если он "внутренний")
type NotificationResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DTO для ответа на создание
type CreateNotificationResponse struct {
	ID string `json:"id"`
}