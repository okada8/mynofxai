package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"nofx/config"
	"nofx/logger"
	"time"
)

// SendTelegramMessage sends a message to the configured Telegram chat
func SendTelegramMessage(message string) error {
	cfg := config.Get()
	token := cfg.TelegramBotToken
	chatID := cfg.TelegramChatID

	if token == "" || chatID == "" {
		return fmt.Errorf("Telegram configuration missing (token or chat_id)")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
		"parse_mode": "Markdown", // Optional: support Markdown
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %s", resp.Status)
	}

	logger.Infof("✓ Telegram notification sent: %s", message)
	return nil
}
