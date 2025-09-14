package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/PavelBradnitski/WbTechL3.1/internal/statuscache"
	"github.com/wb-go/wbf/ginext"
)

// NotificationHandler для работы с уведомлениями
type NotificationHandler struct {
	svc         service.NotificationService
	statusCache *statuscache.Cache
}

// NewNotificationHandler создает новый обработчик уведомлений и регистрирует маршруты
func NewNotificationHandler(r *ginext.Engine, svc service.NotificationService, frontendURL string, cache *statuscache.Cache) {
	log.Printf("Frontend URL: %s\n", frontendURL)
	h := &NotificationHandler{svc: svc, statusCache: cache}
	// CORS middleware
	r.Use(func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", frontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	r.POST("/notify", h.create)
	r.GET("/notify", h.getAll)
	r.GET("/notify/:id", h.get)
	r.DELETE("/notify/:id", h.cancel)
}

// create хендлер для создания нового уведомления.
func (h *NotificationHandler) create(c *ginext.Context) {
	var req models.CreateNotificationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request"})
		return
	}
	if req.Type != models.NotificationTypeEmail && req.Type != models.NotificationTypeTelegram {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "unsupported notification type"})
		return
	}

	if req.Type == models.NotificationTypeEmail && req.Email == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "email is required for email notifications"})
		return
	}

	if req.Type == models.NotificationTypeTelegram && req.ChatID == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "chat_id is required for telegram notifications"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "message cannot be empty"})
		return
	}

	if req.ScheduledAt.IsZero() {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "scheduled_at is required"})
		return
	}

	if req.ScheduledAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "scheduled_at cannot be in the past"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		log.Printf("err: %v\n", err)
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to create notification"})
		return
	}
	// add  to redis cache
	if h.statusCache != nil {
		err = h.statusCache.SetStatus(c.Request.Context(), id, models.StatusScheduled)
		if err != nil {
			log.Printf("failed to set status in redis for id=%v: %v", id, err)
		}
	}

	c.JSON(http.StatusOK, models.CreateNotificationResponse{ID: id})
}

func (h *NotificationHandler) get(c *ginext.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	var status models.Status

	// пробуем сначала из Redis
	if h.statusCache != nil {
		s, err := h.statusCache.GetStatus(ctx, id)
		if err == nil {
			log.Printf("Cache hit for notification %s: status=%s\n", id, s)
			status = s
		}
	}

	// если нет в Redis — берём из БД
	if status == "" {
		n, err := h.svc.Get(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, map[string]any{"error": "notification not found"})
			return
		}
		status = n.Status

		// обновляем Redis для будущих запросов
		if h.statusCache != nil {
			_ = h.statusCache.SetStatus(ctx, id, status)
		}
	}

	c.JSON(http.StatusOK, map[string]string{
		"status": string(status),
	})
}

// getAll хендлер для получения всех уведомлений. (метод для фронтенда)
func (h *NotificationHandler) getAll(c *ginext.Context) {
	n, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		// если уведомлений нет, возвращаем пустой массив
		if err == repository.ErrNotFound {
			c.JSON(http.StatusOK, []models.NotificationResponse{})
			return
		}

		// любая другая ошибка — Internal Server Error
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	var resp []models.NotificationResponse
	for _, notif := range n {
		var response models.NotificationResponse
		switch notif.Type {
		case models.NotificationTypeEmail:
			response = models.NotificationResponse{
				ID:          notif.ID,
				Email:       notif.EmailNotification.Email,
				Type:        notif.Type,
				Message:     notif.EmailNotification.Message,
				Subject:     notif.EmailNotification.Subject,
				ScheduledAt: notif.ScheduledAt,
				Status:      notif.Status,
				Retries:     notif.Retries,
			}
		case models.NotificationTypeTelegram:
			response = models.NotificationResponse{
				ID:          notif.ID,
				ChatID:      notif.TelegramNotification.ChatID,
				Type:        notif.Type,
				Message:     notif.TelegramNotification.Message,
				ScheduledAt: notif.ScheduledAt,
				Status:      notif.Status,
				Retries:     notif.Retries,
			}
		}
		resp = append(resp, response)
	}

	c.JSON(http.StatusOK, resp)
}

// cancel хендлер для отмены запланированного уведомления.
func (h *NotificationHandler) cancel(c *ginext.Context) {
	id := c.Param("id")
	n, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "notification not found"})
		return
	}
	if n.Status != models.StatusScheduled {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "only scheduled notifications can be canceled"})
		return
	}

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	// update redis cache
	if h.statusCache != nil {
		err = h.statusCache.SetStatus(c.Request.Context(), id, models.StatusCanceled)
		if err != nil {
			log.Printf("failed to set status in redis for id=%v: %v", id, err)
		}
	}
	c.JSON(http.StatusOK, map[string]any{"status": "canceled"})
}
