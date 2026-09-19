package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AgentHandler defines the agent interaction required by channel adapters.
type AgentHandler interface {
	Chat(ctx context.Context, userInput string, sender string) (string, error)
}

// DiscordConfig holds credentials for Discord integration.
type DiscordConfig struct {
	BotToken   string `json:"bot_token"`
	ChannelID  string `json:"channel_id"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

// DiscordAdapter connects an agent to a Discord text channel via REST API v10.
type DiscordAdapter struct {
	cfg             DiscordConfig
	handler         AgentHandler
	client          *http.Client
	lastProcessedID string
	selfID          string
}

// NewDiscordAdapter creates a new Discord channel adapter.
func NewDiscordAdapter(cfg DiscordConfig, handler AgentHandler) *DiscordAdapter {
	return &DiscordAdapter{
		cfg:     cfg,
		handler: handler,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// SendMessage posts a message directly to the configured Discord channel.
func (d *DiscordAdapter) SendMessage(ctx context.Context, content string) error {
	if d.cfg.BotToken == "" || d.cfg.ChannelID == "" {
		return fmt.Errorf("discord bot token or channel ID missing")
	}

	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", d.cfg.ChannelID)
	payload, _ := json.Marshal(map[string]string{"content": content})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+d.cfg.BotToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

// Start begins polling for incoming user messages in the configured channel.
func (d *DiscordAdapter) Start(ctx context.Context) {
	if d.cfg.BotToken == "" || d.cfg.ChannelID == "" || d.handler == nil {
		return
	}

	d.fetchSelfID()

	ticker := time.NewTicker(2500 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.poll(ctx)
			}
		}
	}()
}

func (d *DiscordAdapter) fetchSelfID() {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bot "+d.cfg.BotToken)
	resp, err := d.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var me struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&me); err == nil {
		d.selfID = me.ID
	}
}

func (d *DiscordAdapter) poll(ctx context.Context) {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?limit=5", d.cfg.ChannelID)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bot "+d.cfg.BotToken)

	resp, err := d.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var msgs []struct {
		ID      string `json:"id"`
		Content string `json:"content"`
		Author  struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Bot      bool   `json:"bot"`
		} `json:"author"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil || len(msgs) == 0 {
		return
	}

	// First run: sync to newest message to avoid backlog replay
	if d.lastProcessedID == "" {
		d.lastProcessedID = msgs[0].ID
		return
	}

	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]
		if m.ID <= d.lastProcessedID {
			continue
		}
		d.lastProcessedID = m.ID

		if m.Author.Bot || m.Author.ID == d.selfID {
			continue
		}

		cleanContent := strings.TrimSpace(m.Content)
		if cleanContent == "" {
			continue
		}

		go func(content, author string) {
			reply, err := d.handler.Chat(ctx, content, author)
			if err == nil && reply != "" {
				_ = d.SendMessage(ctx, reply)
			}
		}(cleanContent, m.Author.Username)
	}
}
