package service

import (
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/queue"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/google/uuid"
)

// Интерфейс для бизнес-логики
type NotificationService interface {
	Create(message string, sendAt time.Time) (string, error)
	Get(id string) (*models.Notification, error)
	Cancel(id string) error
}

type notificationService struct {
	repo  repository.NotificationRepository
	queue queue.NotificationQueue
}

func NewNotificationService(repo repository.NotificationRepository, q queue.NotificationQueue) NotificationService {
	return &notificationService{repo: repo, queue: q}
}

func (s *notificationService) Create(message string, sendAt time.Time) (string, error) {
	id := uuid.New().String()
	n := &models.Notification{
		ID:      id,
		Message: message,
		SendAt:  sendAt,
		Status:  models.StatusScheduled,
	}
	if err := s.repo.Save(n); err != nil {
		return "", err
	}

	if err := s.queue.Publish(n); err != nil {
		return "", err
	}

	return id, nil
}

func (s *notificationService) Get(id string) (*models.Notification, error) {
	return s.repo.Get(id)
}

func (s *notificationService) Cancel(id string) error {
	return s.repo.Cancel(id)
}
