package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Telegram sends messages via the Bot API.
type Telegram struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegram(botToken, chatID string) *Telegram {
	return &Telegram{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Telegram) Send(text string) error {
	if t.botToken == "" || t.chatID == "" {
		log.Printf("[telegram] not configured, skip send")
		return nil
	}
	body, _ := json.Marshal(map[string]string{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	log.Printf("[telegram] sending to chat_id=%s: %.80s...", t.chatID, text)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[telegram] send error: %v", err)
		return fmt.Errorf("telegram post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		log.Printf("[telegram] send failed: %d %s", resp.StatusCode, b)
		return fmt.Errorf("telegram %d: %s", resp.StatusCode, b)
	}
	log.Printf("[telegram] sent ok")
	return nil
}
