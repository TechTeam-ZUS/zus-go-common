package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Card templates (Lark/Feishu header colors).
const (
	TemplateBlue   = "blue"
	TemplateOrange = "orange"
	TemplateRed    = "red"
)

// timeLayout matches logger's text handler output format.
const timeLayout = "2006-01-02T15:04:05.000Z07:00"

var (
	notifyBot     *Bot
	notifyEnabled = true
)

// Bot posts interactive cards to a single webhook URL (Lark/Feishu bot). Each bot has its own URL, so create one Bot per notification target.
type Bot struct {
	url         string
	serviceName string
	client      *http.Client
}

func NewBot(url, serviceName string) {
	notifyBot = &Bot{
		url:         url,
		serviceName: serviceName,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

func GetBot() *Bot {
	return notifyBot
}

// Notify sends msg to the registered bot, if any and not disabled.
func (bot *Bot) Notify(level slog.Level, title, msg string) {
	if notifyBot == nil || !notifyEnabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notificationTitle := fmt.Sprintf("[%s] %s | %s", bot.serviceName, level, title)
	if err := notifyBot.sendNotification(ctx, level, notificationTitle, msg); err != nil {
		fmt.Fprintf(os.Stderr, "logger: failed to send notification: %v\n", err)
	}
}

func (b *Bot) sendNotification(ctx context.Context, level slog.Level, title, content string) error {
	if level < slog.LevelWarn {
		return b.sendText(ctx, title, content)
	}

	return b.sendAlert(ctx, level, title, content)
}

func (b *Bot) sendText(ctx context.Context, title, content string) error {
	return b.send(ctx, NewCard(title, TemplateBlue, content))
}

func (b *Bot) sendAlert(ctx context.Context, level slog.Level, title, content string) error {
	var template string
	if level >= slog.LevelError {
		template = TemplateRed
	} else {
		template = TemplateOrange
	}

	return b.send(ctx, NewCard(title, template, content))
}

// Send posts an arbitrary JSON payload to the bot's webhook URL.
func (b *Bot) send(ctx context.Context, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("notify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: unexpected status %d", resp.StatusCode)
	}
	return nil
}
