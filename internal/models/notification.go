package models

import "time"

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusSent      Status = "sent"
	StatusCanceled  Status = "canceled"
	StatusFailed    Status = "failed"
)

type Notification struct {
	ID         string
	Message    string    `json:"message"`
	SendAt     time.Time `json:"send_at"`
	Status     Status    `json:"status"`
	RetryCount int       `json:"retry_count"`
}
