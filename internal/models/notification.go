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
	ID                    string           `db:"id"`
	Type                  NotificationType `db:"type"`
	Status                Status           `db:"status"`
	ScheduledAt           time.Time        `db:"scheduled_at"`
	Retries               int              `db:"retries"`
	CreatedAt             time.Time        `db:"created_at"`
	UpdatedAt             time.Time        `db:"updated_at"`
	*EmailNotification    `db:"email_notifications"`
	*TelegramNotification `db:"telegram_notifications"`
}

type EmailNotification struct {
	ID             string `db:"id"`
	NotificationID string `db:"notification_id"`
	Email          string `db:"email"`
	Subject        string `db:"subject"`
	Message        string `db:"message"`
}

type TelegramNotification struct {
	ID             string `db:"id"`
	NotificationID string `db:"notification_id"`
	ChatID         string `db:"chat_id"`
	Message        string `db:"message"`
}

// CreateNotificationRequest DTO для входящих запросов (например, POST /notifications)
type CreateNotificationRequest struct {
	ChatID      string           `json:"chat_id,omitempty"`
	Email       string           `json:"email,omitempty"`
	Type        NotificationType `json:"type"` // email | telegram
	Message     string           `json:"message"`
	Subject     string           `json:"subject,omitempty"`
	ScheduledAt time.Time        `json:"scheduled_at"`
}

// NotificationResponse DTO для ответа API
type NotificationResponse struct {
	ID          string           `json:"id"`
	ChatID      string           `json:"chat_id,omitempty"`
	Email       string           `json:"email,omitempty"`
	Type        NotificationType `json:"type"`
	Message     string           `json:"message"`
	Subject     string           `json:"subject,omitempty"`
	ScheduledAt time.Time        `json:"scheduled_at"`
	Status      Status           `json:"status"`
}

// CreateNotificationResponse DTO для ответа на создание
type CreateNotificationResponse struct {
	ID string `json:"id"`
}
