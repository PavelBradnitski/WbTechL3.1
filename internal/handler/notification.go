package handler

import (
	"net/http"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/wb-go/wbf/ginext"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) Routes() http.Handler {
	r := ginext.New()

	r.Engine.POST("/notify", h.Create)
	r.Engine.GET("/notify/:id", h.Get)
	r.Engine.DELETE("/notify/:id", h.Cancel)
	return r
}

type createRequest struct {
	Message string    `json:"message"`
	SendAt  time.Time `json:"send_at"`
}

type H map[string]any

func (h *NotificationHandler) Create(c *ginext.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, H{"error": "bad request"})
		return
	}

	id, err := h.svc.Create(req.Message, req.SendAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"error": "failed to create"})
		return
	}

	c.JSON(http.StatusOK, H{"id": id})
}

func (h *NotificationHandler) Get(c *ginext.Context) {
	id := c.Param("id")
	n, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

func (h *NotificationHandler) Cancel(c *ginext.Context) {
	id := c.Param("id")
	if err := h.svc.Cancel(id); err != nil {
		c.JSON(http.StatusNotFound, H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
