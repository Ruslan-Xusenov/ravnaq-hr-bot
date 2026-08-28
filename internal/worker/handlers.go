package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/company/hrbot/internal/domain/user"
	"github.com/hibiken/asynq"
	"gopkg.in/telebot.v3"
)

type BroadcastProcessor struct {
	userRepo user.Repository
	tbot     *telebot.Bot
}

func NewBroadcastProcessor(userRepo user.Repository, tbot *telebot.Bot) *BroadcastProcessor {
	return &BroadcastProcessor{
		userRepo: userRepo,
		tbot:     tbot,
	}
}

func (p *BroadcastProcessor) HandleBroadcastMessage(ctx context.Context, t *asynq.Task) error {
	var payload BroadcastMessagePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	users, err := p.userRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	successCount := 0
	failCount := 0

	for _, u := range users {
		// Skip blocked users
		if u.Status == "blocked" {
			continue
		}

		// Only send to users of the specified language, or all if empty.
		if payload.LangCode != "" && u.LanguageCode != nil && *u.LanguageCode != payload.LangCode {
			continue
		}

		chat := &telebot.Chat{ID: u.TelegramID}
		_, err := p.tbot.Send(chat, payload.Text, telebot.ModeHTML)
		if err != nil {
			slog.Error("Failed to send broadcast to user", "user_id", u.TelegramID, "error", err)
			failCount++
			
			// Mark user as blocked if they blocked the bot
			if strings.Contains(strings.ToLower(err.Error()), "blocked") {
				u.Status = "blocked"
				_ = p.userRepo.Update(ctx, &u)
			}
		} else {
			successCount++
		}
		
		// Rate limiting to avoid Telegram "Too Many Requests" (max 30 msgs/sec, safe to wait a bit)
		time.Sleep(35 * time.Millisecond)
	}

	slog.Info("Broadcast finished", "success", successCount, "failed", failCount)
	return nil
}
