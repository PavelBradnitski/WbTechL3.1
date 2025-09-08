package models

import "time"

type Status string

const (
	StatusScheduled  Status = "scheduled"
	StatusSent       Status = "sent"
	StatusCanceled   Status = "canceled"
	StatusFailed     Status = "failed"
	StatusProcessing Status = "processing"
)

// Тип доставки уведомления
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeTelegram NotificationType = "telegram"
)

// Модель для БД (внутренняя)
type Notification struct {
	ID          string           `db:"id"`
	UserID      string           `db:"user_id"`
	Email       string           `db:"email"`
	Type        NotificationType `db:"type"`
	Message     string           `db:"message"`
	Subject     string           `db:"subject"`
	ScheduledAt time.Time        `db:"scheduled_at"`
	Status      Status           `db:"status"`
	Retries     int              `db:"retries"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
}

// DTO для входящих запросов (например, POST /notifications)
type CreateNotificationRequest struct {
	UserID      string           `json:"user_id,omitempty"`
	Email       string           `json:"email,omitempty"`
	Type        NotificationType `json:"type"` // email | telegram
	Message     string           `json:"message"`
	Subject     string           `json:"subject,omitempty"`
	ScheduledAt time.Time        `json:"scheduled_at"`
}

// DTO для ответа API
type NotificationResponse struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id,omitempty"`
	Email       string           `json:"email,omitempty"`
	Type        NotificationType `json:"type"`
	Message     string           `json:"message"`
	Subject     string           `json:"subject,omitempty"`
	ScheduledAt time.Time        `json:"scheduled_at"`
	Status      Status           `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// DTO для ответа на создание
type CreateNotificationResponse struct {
	ID string `json:"id"`
}
