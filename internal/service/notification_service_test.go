package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationRepository is a mock implementation of the NotificationRepository interface.
type MockNotificationRepository struct {
	mock.Mock
}

// Create mocks the Create method.
func (m *MockNotificationRepository) Create(ctx context.Context, n *models.Notification) (string, error) {
	args := m.Called(ctx, n)
	return args.String(0), args.Error(1)
}

// GetByID mocks the GetByID method.
func (m *MockNotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	args := m.Called(ctx, id)
	// Handle nil return from the mock
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

// GetAll mocks the GetAll method.
func (m *MockNotificationRepository) GetAll(ctx context.Context) ([]*models.Notification, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Notification), args.Error(1)
}

// Cancel mocks the Cancel method.
func (m *MockNotificationRepository) Cancel(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ReservePending mocks the ReservePending method.
func (m *MockNotificationRepository) ReservePending(ctx context.Context, limit int) ([]*models.Notification, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*models.Notification), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method.
func (m *MockNotificationRepository) UpdateStatus(ctx context.Context, id string, status models.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// IncrementRetries mocks the IncrementRetries method.
func (m *MockNotificationRepository) IncrementRetries(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNotificationServiceCreate(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	req := &models.CreateNotificationRequest{
		ChatID:      "user123",
		Message:     "Test message",
		Subject:     "Test subject",
		Email:       "test@example.com",
		Type:        models.NotificationTypeEmail,
		ScheduledAt: time.Now().Add(time.Hour),
	}

	expectedID := uuid.NewString()
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Notification")).
		Return(expectedID, nil)

	id, err := service.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceGet(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	notificationID := "123"
	expectedNotification := &models.Notification{
		ID:          notificationID,
		Type:        models.NotificationTypeEmail,
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      models.StatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		EmailNotification: &models.EmailNotification{
			Email:   "test@example.com",
			Message: "Test message",
			Subject: "Test subject",
		},
	}

	mockRepo.On("GetByID", mock.Anything, notificationID).Return(expectedNotification, nil)

	notification, err := service.Get(context.Background(), notificationID)

	assert.NoError(t, err)
	assert.Equal(t, expectedNotification, notification)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceGetAll(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	expectedNotifications := []*models.Notification{
		{
			ID:          "1",
			Type:        models.NotificationTypeEmail,
			ScheduledAt: time.Now().Add(time.Hour),
			Status:      models.StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			EmailNotification: &models.EmailNotification{
				Email:   "test1@example.com",
				Message: "Test message 1",
				Subject: "Test subject 1",
			},
		},
		{
			ID:          "2",
			Type:        models.NotificationTypeTelegram,
			ScheduledAt: time.Now().Add(2 * time.Hour),
			Status:      models.StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			TelegramNotification: &models.TelegramNotification{
				ChatID:  "user2",
				Message: "Test message 2",
			},
		},
	}

	mockRepo.On("GetAll", mock.Anything).Return(expectedNotifications, nil)

	notifications, err := service.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedNotifications, notifications)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceCancel(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	notificationID := "123"

	mockRepo.On("Cancel", mock.Anything, notificationID).Return(nil)

	err := service.Cancel(context.Background(), notificationID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceReservePending(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	limit := 10
	expectedNotifications := []*models.Notification{
		{
			ID:          "1",
			Type:        models.NotificationTypeEmail,
			ScheduledAt: time.Now().Add(-time.Minute), // Set ScheduledAt in the past
			Status:      models.StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			EmailNotification: &models.EmailNotification{
				Email:   "test1@example.com",
				Message: "Test message 1",
				Subject: "Test subject 1",
			},
		},
	}

	mockRepo.On("ReservePending", mock.Anything, limit).Return(expectedNotifications, nil)

	notifications, err := service.ReservePending(context.Background(), limit)

	assert.NoError(t, err)
	assert.Equal(t, expectedNotifications, notifications)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceUpdateStatus(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	notificationID := "123"
	newStatus := models.StatusSent

	mockRepo.On("UpdateStatus", mock.Anything, notificationID, newStatus).Return(nil)

	err := service.UpdateStatus(context.Background(), notificationID, newStatus)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceIncrementRetries(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	notificationID := "123"

	mockRepo.On("IncrementRetries", mock.Anything, notificationID).Return(nil)

	err := service.IncrementRetries(context.Background(), notificationID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Add tests for error cases for each method.  For example:

func TestNotificationServiceGetError(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	notificationID := "123"
	expectedError := errors.New("not found")

	mockRepo.On("GetByID", mock.Anything, notificationID).Return(nil, expectedError)

	notification, err := service.Get(context.Background(), notificationID)

	assert.Error(t, err)
	assert.Nil(t, notification)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationServiceCreateError(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo)

	req := &models.CreateNotificationRequest{
		ChatID:      "user123",
		Message:     "Test message",
		Subject:     "Test subject",
		Email:       "test@example.com",
		Type:        models.NotificationTypeEmail,
		ScheduledAt: time.Now().Add(time.Hour),
	}

	expectedError := errors.New("creation error")
	mockRepo.On("Create", mock.Anything, mock.Anything).Return("", expectedError)

	id, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
