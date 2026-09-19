package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TelegramConfig holds credentials for Telegram integration.
type TelegramConfig struct {
	Token  string `json:"token"`
	ChatID string `json:"chat_id"`
}

// TelegramAdapter connects an agent to Telegram via Bot API polling.
type TelegramAdapter struct {
	cfg        TelegramConfig
	handler    AgentHandler
	client     *http.Client
	lastUpdate int
}

// NewTelegramAdapter creates a new Telegram adapter.
func NewTelegramAdapter(cfg TelegramConfig, handler AgentHandler) *TelegramAdapter {
	return &TelegramAdapter{
		cfg:     cfg,
		handler: handler,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// SendMessage sends text to the target Telegram chat.
func (t *TelegramAdapter) SendMessage(ctx context.Context, text string) error {
	if t.cfg.Token == "" || t.cfg.ChatID == "" {
		return fmt.Errorf("telegram token or chat ID missing")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.cfg.Token)
	payload, _ := json.Marshal(map[string]string{
		"chat_id": t.cfg.ChatID,
		"text":    text,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

// Start launches polling for incoming updates.
func (t *TelegramAdapter) Start(ctx context.Context) {
	if t.cfg.Token == "" || t.handler == nil {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.poll(ctx)
			}
		}
	}()
}

func (t *TelegramAdapter) poll(ctx context.Context) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=5", t.cfg.Token, t.lastUpdate+1)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data struct {
		OK     bool `json:"ok"`
		Result []struct {
			UpdateID int `json:"update_id"`
			Message  *struct {
				From struct {
					FirstName string `json:"first_name"`
					Username  string `json:"username"`
				} `json:"from"`
				Chat struct {
					ID int64 `json:"id"`
				} `json:"chat"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || !data.OK {
		return
	}

	for _, item := range data.Result {
		if item.UpdateID > t.lastUpdate {
			t.lastUpdate = item.UpdateID
		}

		if item.Message == nil || item.Message.Text == "" {
			continue
		}

		chatIDStr := strconv.FormatInt(item.Message.Chat.ID, 10)
		if t.cfg.ChatID != "" && t.cfg.ChatID != chatIDStr {
			continue // Filter specific authorized chat
		}

		sender := item.Message.From.FirstName
		if sender == "" {
			sender = item.Message.From.Username
		}
		if sender == "" {
			sender = "TelegramUser"
		}

		text := strings.TrimSpace(item.Message.Text)
		go func(msgText, senderName, replyChatID string) {
			reply, err := t.handler.Chat(ctx, msgText, senderName)
			if err == nil && reply != "" {
				apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.cfg.Token)
				vals := url.Values{
					"chat_id": {replyChatID},
					"text":    {reply},
				}
				_, _ = t.client.PostForm(apiURL, vals)
			}
		}(text, sender, chatIDStr)
	}
}
