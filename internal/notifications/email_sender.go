// internal/notifications/email_sender.go
package notifications

import (
	"fmt"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailSender(host string, port int, username, password, from string) *EmailSender {
	return &EmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *EmailSender) Send(n *models.Notification) error {

	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", n.Email)
	m.SetHeader("Subject", n.Subject)
	m.SetBody("text/plain", n.Message)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	d.SSL = false // MailHog не использует TLS

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
