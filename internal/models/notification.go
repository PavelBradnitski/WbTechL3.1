package models

import "time"

// Status Статус уведомления
type Status string

const (
	// StatusScheduled статус при создании
	StatusScheduled Status = "scheduled"
	// StatusSent статус после попытки отправки
	StatusSent Status = "sent"
	// StatusCanceled статус после отмены
	StatusCanceled Status = "canceled"
	// StatusFailed статус после неудачной попытки отправки
	StatusFailed Status = "failed"
	// StatusProcessing статус при начале обработки планировщиком
	StatusProcessing Status = "processing"
)

// NotificationType Тип доставки уведомления
type NotificationType string

const (
	// NotificationTypeEmail константа для email уведомлений
	NotificationTypeEmail NotificationType = "email"
	// NotificationTypeTelegram константа для telegram уведомлений
	NotificationTypeTelegram NotificationType = "telegram"
)

// Notification Модель для БД (внутренняя)
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

// CreateNotificationRequest DTO для входящих запросов (например, POST /notifications)
type CreateNotificationRequest struct {
	UserID      string           `json:"user_id,omitempty"`
	Email       string           `json:"email,omitempty"`
	Type        NotificationType `json:"type"` // email | telegram
	Message     string           `json:"message"`
	Subject     string           `json:"subject,omitempty"`
	ScheduledAt time.Time        `json:"scheduled_at"`
}

// NotificationResponse DTO для ответа API
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

// CreateNotificationResponse DTO для ответа на создание
type CreateNotificationResponse struct {
	ID string `json:"id"`
}
