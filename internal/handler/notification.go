package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/notify", h.Create)
	r.Get("/notify/{id}", h.Get)
	r.Delete("/notify/{id}", h.Cancel)
	return r
}

type createRequest struct {
	Message string    `json:"message"`
	SendAt  time.Time `json:"send_at"`
}

func (h *NotificationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, err := h.svc.Create(req.Message, req.SendAt)
	if err != nil {
		http.Error(w, "failed to create", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *NotificationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	n, err := h.svc.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n)
}

func (h *NotificationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Cancel(id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
