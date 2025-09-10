package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wb-go/wbf/ginext"
)

// MockNotificationService - мок для сервиса уведомлений
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) Create(ctx context.Context, req *models.CreateNotificationRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockNotificationService) Get(ctx context.Context, id string) (*models.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationService) GetAll(ctx context.Context) ([]*models.Notification, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Notification), args.Error(1)
}

func (m *MockNotificationService) Cancel(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationService) ReservePending(ctx context.Context, limit int) ([]*models.Notification, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*models.Notification), args.Error(1)
}

func (m *MockNotificationService) UpdateStatus(ctx context.Context, id string, status models.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockNotificationService) IncrementRetries(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestCreateNotificationHandlerSuccess - Тест успешного создания уведомления
func TestCreateNotificationHandlerSuccess(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	scheduledAt := time.Now().Add(time.Hour)
	requestBody := models.CreateNotificationRequest{
		Type:        models.NotificationTypeEmail,
		Email:       "test@example.com",
		Message:     "Test message",
		ScheduledAt: scheduledAt,
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Expectations
	expectedID := "test-notification-id"
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateNotificationRequest")).Return(expectedID, nil)
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.CreateNotificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, response.ID)

	mockService.AssertExpectations(t)
}

// TestCreateNotificationHandlerInvalidRequest - Тест с невалидным запросом (ошибка парсинга JSON)
func TestCreateNotificationHandlerInvalidRequest(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	// Test - Невалидный JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestCreateNotificationHandlerUnsupportedType - Тест с неподдерживаемым типом уведомления
func TestCreateNotificationHandlerUnsupportedType(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	scheduledAt := time.Now().Add(time.Hour)
	requestBody := models.CreateNotificationRequest{
		Type:        "unknown",
		Email:       "test@example.com",
		Message:     "Test message",
		ScheduledAt: scheduledAt,
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unsupported notification type", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestCreateNotificationHandlerMissingEmail - Тест с отсутствующим email для email-уведомления
func TestCreateNotificationHandlerMissingEmail(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	scheduledAt := time.Now().Add(time.Hour)
	requestBody := models.CreateNotificationRequest{
		Type:        models.NotificationTypeEmail,
		Message:     "Test message",
		ScheduledAt: scheduledAt,
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "email is required for email notifications", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestCreateNotificationHandlerMissingUserID - Тест с отсутствующим user_id для telegram-уведомления
func TestCreateNotificationHandlerMissingUserID(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	scheduledAt := time.Now().Add(time.Hour)
	requestBody := models.CreateNotificationRequest{
		Type:        models.NotificationTypeTelegram,
		Message:     "Test message",
		ScheduledAt: scheduledAt,
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user_id is required for telegram notifications", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestCreateNotificationHandlerMissingMessage - Тест с отсутствующим сообщением
func TestCreateNotificationHandlerMissingMessage(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	scheduledAt := time.Now().Add(time.Hour)
	requestBody := models.CreateNotificationRequest{
		Type:        models.NotificationTypeEmail,
		Email:       "test@example.com",
		ScheduledAt: scheduledAt,
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "message cannot be empty", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestCreateNotificationHandlerMissingScheduledAt - Тест с отсутствующим ScheduledAt
func TestCreateNotificationHandlerMissingScheduledAt(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	requestBody := models.CreateNotificationRequest{
		Type:    models.NotificationTypeEmail,
		Email:   "test@example.com",
		Message: "Test message",
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "scheduled_at is required", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestCreateNotificationHandlerScheduledAtInPast - Тест с ScheduledAt в прошлом
func TestCreateNotificationHandlerScheduledAtInPast(t *testing.T) {
	// Setup
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.POST("/notify", handler.create)

	scheduledAt := time.Now().Add(-time.Hour)
	requestBody := models.CreateNotificationRequest{
		Type:        models.NotificationTypeEmail,
		Email:       "test@example.com",
		Message:     "Test message",
		ScheduledAt: scheduledAt,
	}

	jsonValue, _ := json.Marshal(requestBody)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "scheduled_at cannot be in the past", response["error"])

	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestGetNotificationHandlerSuccess tests the get handler when the notification is found.
func TestGetNotificationHandlerSuccess(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.GET("/notify/:id", handler.get)

	notificationID := "123"
	expectedNotification := &models.Notification{
		ID:          notificationID,
		UserID:      "user123",
		Email:       "test@example.com",
		Type:        models.NotificationTypeEmail,
		Message:     "Test message",
		Subject:     "Test subject",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      models.StatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("Get", mock.Anything, notificationID).Return(expectedNotification, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notify/"+notificationID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.NotificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedNotification.ID, response.ID)
	assert.Equal(t, expectedNotification.UserID, response.UserID)
	assert.Equal(t, expectedNotification.Email, response.Email)
	assert.Equal(t, expectedNotification.Type, response.Type)
	assert.Equal(t, expectedNotification.Message, response.Message)
	assert.Equal(t, expectedNotification.Subject, response.Subject)
	assert.True(t, expectedNotification.CreatedAt.Truncate(time.Second).Equal(response.CreatedAt.Truncate(time.Second)), "CreatedAt times are not equal")
	assert.Equal(t, expectedNotification.Status, response.Status)
	assert.WithinDuration(t, expectedNotification.CreatedAt, response.CreatedAt, time.Second)
	assert.WithinDuration(t, expectedNotification.UpdatedAt, response.UpdatedAt, time.Second)

	mockService.AssertExpectations(t)
}

// TestGetNotificationHandlerNotFound tests the get handler when the notification is not found.
func TestGetNotificationHandlerNotFound(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.GET("/notify/:id", handler.get)

	notificationID := "123"
	mockService.On("Get", mock.Anything, notificationID).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notify/"+notificationID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "notification not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestGetAllNotificationHandlerSuccess(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.GET("/notify", handler.getAll)

	expectedNotifications := []*models.Notification{
		{
			ID:          "1",
			UserID:      "",
			Email:       "test1@example.com",
			Type:        models.NotificationTypeEmail,
			Message:     "Test message 1",
			Subject:     "Test subject 1",
			ScheduledAt: time.Now().Add(time.Hour),
			Status:      models.StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "2",
			UserID:      "user2",
			Email:       "",
			Type:        models.NotificationTypeTelegram,
			Message:     "Test message 2",
			ScheduledAt: time.Now().Add(2 * time.Hour),
			Status:      models.StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockService.On("GetAll", mock.Anything).Return(expectedNotifications, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notify", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.NotificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response, len(expectedNotifications))

	for i, expected := range expectedNotifications {
		assert.Equal(t, expected.ID, response[i].ID)
		assert.Equal(t, expected.UserID, response[i].UserID)
		assert.Equal(t, expected.Email, response[i].Email)
		assert.Equal(t, expected.Type, response[i].Type)
		assert.Equal(t, expected.Message, response[i].Message)
		assert.Equal(t, expected.Subject, response[i].Subject)
		assert.Equal(t, expected.Status, response[i].Status)

		assert.WithinDuration(t, expected.CreatedAt, response[i].CreatedAt, time.Second)
		assert.WithinDuration(t, expected.UpdatedAt, response[i].UpdatedAt, time.Second)
	}

	mockService.AssertExpectations(t)
}

func TestGetAllNotificationHandlerNotFound(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.GET("/notify", handler.getAll)

	mockService.On("GetAll", mock.Anything).Return([]*models.Notification{}, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notify", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "notifications not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestCancelNotificationHandlerSuccess(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := gin.New()
	router.POST("/notify/:id/cancel", handler.cancel)

	notificationID := "123"

	scheduledNotification := &models.Notification{
		ID:          notificationID,
		UserID:      "user123",
		Email:       "test@example.com",
		Type:        models.NotificationTypeEmail,
		Message:     "Test message",
		Subject:     "Test subject",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      models.StatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("Get", mock.Anything, notificationID).Return(scheduledNotification, nil)
	mockService.On("Cancel", mock.Anything, notificationID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notify/"+notificationID+"/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "canceled", response["status"])

	mockService.AssertExpectations(t)
}

func TestCancelNotificationHandlerNotFound(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.DELETE("/notify/:id", handler.cancel)

	notificationID := "123"

	mockService.On("Get", mock.Anything, notificationID).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notify/"+notificationID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "notification not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestCancelNotificationHandlerNotScheduled(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.DELETE("/notify/:id", handler.cancel)

	notificationID := "123"

	processingNotification := &models.Notification{
		ID:          notificationID,
		UserID:      "user123",
		Email:       "test@example.com",
		Type:        models.NotificationTypeEmail,
		Message:     "Test message",
		Subject:     "Test subject",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      models.StatusProcessing,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("Get", mock.Anything, notificationID).Return(processingNotification, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notify/"+notificationID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "only scheduled notifications can be canceled", response["error"])

	mockService.AssertExpectations(t)
}

func TestCancelNotificationHandlerCancelError(t *testing.T) {
	mockService := new(MockNotificationService)
	handler := &NotificationHandler{svc: mockService}
	router := ginext.New()
	router.DELETE("/notify/:id", handler.cancel)

	notificationID := "123"

	scheduledNotification := &models.Notification{
		ID:          notificationID,
		UserID:      "user123",
		Email:       "test@example.com",
		Type:        models.NotificationTypeEmail,
		Message:     "Test message",
		Subject:     "Test subject",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      models.StatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("Get", mock.Anything, notificationID).Return(scheduledNotification, nil)
	mockService.On("Cancel", mock.Anything, notificationID).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notify/"+notificationID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, assert.AnError.Error(), response["error"]) // Verifies the error message

	mockService.AssertExpectations(t)
}
