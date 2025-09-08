package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

type TelegramSender struct {
	botToken string
}

func NewTelegramSender(botToken string) *TelegramSender {
	return &TelegramSender{botToken: botToken}
}

func (s *TelegramSender) Send(n *models.Notification) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	payload := map[string]string{
		"chat_id": n.UserID,
		"text":    n.Message,
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("telegram send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram send failed: %s", string(body))
	}

	return nil
}
