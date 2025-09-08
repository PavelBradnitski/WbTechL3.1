package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/wb-go/wbf/ginext"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(r *ginext.Engine, svc service.NotificationService, frontendURL string) {
	log.Printf("Frontend URL: %s\n", frontendURL)
	h := &NotificationHandler{svc: svc}
	// Простая настройка CORS
	// CORS middleware
	r.Use(func(c *ginext.Context) {
		log.Printf("!!!Incoming request: %s %s\n", c.Request.Method, c.Request.URL.Path)
		log.Printf("!!!Origin header: %s\n", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
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
	r.GET("/notify/:id", h.get)
	r.GET("/notify/All", h.getAll)
	r.DELETE("/notify/:id", h.cancel)
}

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

	if req.Type == models.NotificationTypeTelegram && req.UserID == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "user_id is required for telegram notifications"})
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
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to create notification"})
		return
	}

	c.JSON(http.StatusOK, models.CreateNotificationResponse{ID: id})
}

func (h *NotificationHandler) get(c *ginext.Context) {
	id := c.Param("id")
	n, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "notification not found"})
		return
	}
	resp := models.NotificationResponse{
		ID:          n.ID,
		UserID:      n.UserID,
		Email:       n.Email,
		Type:        n.Type,
		Message:     n.Message,
		Subject:     n.Subject,
		ScheduledAt: n.ScheduledAt,
		Status:      n.Status,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) getAll(c *ginext.Context) {

	n, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "notifications not found"})
		return
	}
	var resp []models.NotificationResponse
	for _, notif := range n {
		resp = append(resp, models.NotificationResponse{
			ID:          notif.ID,
			UserID:      notif.UserID,
			Email:       notif.Email,
			Type:        notif.Type,
			Message:     notif.Message,
			Subject:     notif.Subject,
			ScheduledAt: notif.ScheduledAt,
			Status:      notif.Status,
			CreatedAt:   notif.CreatedAt,
			UpdatedAt:   notif.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

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

	c.JSON(http.StatusOK, map[string]any{"status": "canceled"})
}
