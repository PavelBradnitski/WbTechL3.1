package sender

import (
	"fmt"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

// MultiSender реализует отправку уведомлений через несколько каналов (email и telegram).
type MultiSender struct {
	emailSender    Sender
	telegramSender Sender
}

// NewMultiSender создает новый экземпляр MultiSender.
func NewMultiSender(emailSender, telegramSender Sender) *MultiSender {
	return &MultiSender{emailSender: emailSender, telegramSender: telegramSender}
}

// Send отправляет уведомление через соответствующий канал в зависимости от типа уведомления.
func (m *MultiSender) Send(n *models.Notification) error {
	switch n.Type {
	case "email":
		return m.emailSender.Send(n)
	case "telegram":
		return m.telegramSender.Send(n)
	default:
		return fmt.Errorf("unsupported notification type: %s", n.Type)
	}
}
