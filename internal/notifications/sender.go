// internal/notifications/sender.go
package notifications

import "github.com/PavelBradnitski/WbTechL3.1/internal/models"

type Sender interface {
	Send(n *models.Notification) error
}
