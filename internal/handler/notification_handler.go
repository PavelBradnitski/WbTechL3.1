package handler

import (
	"net/http"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/wb-go/wbf/ginext"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(r *ginext.Engine, svc service.NotificationService) {
	h := &NotificationHandler{svc: svc}

	r.POST("/notify", h.create)
	r.GET("/notify/:id", h.get)
	r.DELETE("/notify/:id", h.cancel)
}

func (h *NotificationHandler) create(c *ginext.Context) {
	var req models.CreateNotificationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request"})
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
		Message:     n.Message,
		ScheduledAt: n.ScheduledAt,
		Status:      n.Status,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) cancel(c *ginext.Context) {
	id := c.Param("id")
	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"status": "canceled"})
}
