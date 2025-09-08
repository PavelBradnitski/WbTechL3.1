package notifications

import (
	"fmt"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

type MultiSender struct {
	emailSender    Sender
	telegramSender Sender
}

func NewMultiSender(emailSender, telegramSender Sender) *MultiSender {
	return &MultiSender{emailSender: emailSender, telegramSender: telegramSender}
}

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
