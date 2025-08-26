package repository

import (
	"errors"
	"sync"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

type NotificationRepository interface {
	Save(n *models.Notification) error
	Get(id string) (*models.Notification, error)
	Cancel(id string) error
}

type memoryRepo struct {
	mu   sync.RWMutex
	data map[string]*models.Notification
}

func NewMemoryRepo() NotificationRepository {
	return &memoryRepo{
		data: make(map[string]*models.Notification),
	}
}

func (r *memoryRepo) Save(n *models.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[n.ID] = n
	return nil
}

func (r *memoryRepo) Get(id string) (*models.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if n, ok := r.data[id]; ok {
		return n, nil
	}
	return nil, errors.New("not found")
}

func (r *memoryRepo) Cancel(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n, ok := r.data[id]; ok {
		n.Status = models.StatusCanceled
		return nil
	}
	return errors.New("not found")
}
