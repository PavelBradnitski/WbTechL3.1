package service

import (
	"context"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/google/uuid"
)

type NotificationService interface {
	Create(ctx context.Context, req *models.CreateNotificationRequest) (string, error)
	Get(ctx context.Context, id string) (*models.Notification, error)
	Cancel(ctx context.Context, id string) error
	GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error)
	UpdateStatus(ctx context.Context, id, status string, retries int) error
	IncrementRetries(ctx context.Context, id string) error
	MarkAsSent(ctx context.Context, id string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) Create(ctx context.Context, req *models.CreateNotificationRequest) (string, error) {
	n := &models.Notification{
		ID:          uuid.NewString(),
		UserID:      req.UserID,
		Message:     req.Message,
		ScheduledAt: req.ScheduledAt,
		Status:      "scheduled",
		Retries:     0,
	}
	return s.repo.Create(ctx, n)
}

func (s *notificationService) Get(ctx context.Context, id string) (*models.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *notificationService) Cancel(ctx context.Context, id string) error {
	return s.repo.Cancel(ctx, id)
}

func (s *notificationService) GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error) {
	// возвращает уведомления со статусом 'pending' и send_at <= now()
	return s.repo.GetPending(ctx, limit)
}

func (s *notificationService) UpdateStatus(ctx context.Context, id, status string, retries int) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *notificationService) IncrementRetries(ctx context.Context, id string) error {
	return s.repo.IncrementRetries(ctx, id)
}

func (s *notificationService) MarkAsSent(ctx context.Context, id string) error {
	return s.repo.UpdateStatus(ctx, id, "sent")
}
