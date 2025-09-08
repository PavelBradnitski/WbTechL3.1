package sender

import "github.com/PavelBradnitski/WbTechL3.1/internal/models"

// Sender описывает метод для отправки уведомлений.
type Sender interface {
	Send(n *models.Notification) error
}
