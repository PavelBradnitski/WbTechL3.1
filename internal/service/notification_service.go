package service

import (
	"context"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/google/uuid"
)

// NotificationService описывает методы для работы с уведомлениями.
type NotificationService interface {
	Create(ctx context.Context, req *models.CreateNotificationRequest) (string, error)
	Get(ctx context.Context, id string) (*models.Notification, error)
	GetAll(ctx context.Context) ([]*models.Notification, error)
	Cancel(ctx context.Context, id string) error
	ReservePending(ctx context.Context, limit int) ([]*models.Notification, error)
	UpdateStatus(ctx context.Context, id string, status models.Status) error
	IncrementRetries(ctx context.Context, id string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

// NewNotificationService создает новый экземпляр NotificationService.
func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

// Create создает новое уведомление.
func (s *notificationService) Create(ctx context.Context, req *models.CreateNotificationRequest) (string, error) {
	n := &models.Notification{
		ID:          uuid.NewString(),
		UserID:      req.UserID,
		Message:     req.Message,
		Subject:     req.Subject,
		Email:       req.Email,
		Type:        req.Type,
		ScheduledAt: req.ScheduledAt,
		Status:      "scheduled",
		Retries:     0,
	}
	return s.repo.Create(ctx, n)
}

// Get возвращает уведомление по его ID.
func (s *notificationService) Get(ctx context.Context, id string) (*models.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

// GetAll возвращает все уведомления.
func (s *notificationService) GetAll(ctx context.Context) ([]*models.Notification, error) {
	return s.repo.GetAll(ctx)
}

// Cancel отменяет запланированное уведомление.
func (s *notificationService) Cancel(ctx context.Context, id string) error {
	return s.repo.Cancel(ctx, id)
}

// ReservePending возвращает уведомления со статусом 'pending' и send_at <= now()
func (s *notificationService) ReservePending(ctx context.Context, limit int) ([]*models.Notification, error) {
	return s.repo.ReservePending(ctx, limit)
}

// UpdateStatus обновляет статус уведомления.
func (s *notificationService) UpdateStatus(ctx context.Context, id string, status models.Status) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// IncrementRetries увеличивает счетчик попыток отправки уведомления.
func (s *notificationService) IncrementRetries(ctx context.Context, id string) error {
	return s.repo.IncrementRetries(ctx, id)
}
