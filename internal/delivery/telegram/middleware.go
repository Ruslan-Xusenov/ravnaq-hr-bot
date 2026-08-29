package telegram

import (
	"context"
	"fmt"

	"github.com/company/hrbot/internal/domain/channel"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) subscriptionMiddleware() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// Do not block admin commands or callbacks
			telegramID := c.Sender().ID
			isAdmin := false
			for _, id := range b.adminIDs {
				if id == telegramID {
					isAdmin = true
					break
				}
			}
			if isAdmin {
				return next(c)
			}

			// Do not block "check_sub" callback
			if c.Callback() != nil && c.Callback().Data == "check_sub" {
				return next(c)
			}

			ctx := context.Background()
			channels, err := b.channelRepo.GetAll(ctx)
			if err != nil || len(channels) == 0 {
				return next(c) // No mandatory channels, proceed
			}

			var unjoinedChannels []channel.Channel
			for _, ch := range channels {
				chat := &tele.Chat{ID: ch.ChatID}
				member, err := b.Client.ChatMemberOf(chat, c.Sender())
				if err != nil {
					// If error checking, maybe bot is not admin anymore. We skip blocking to be safe.
					continue
				}
				if member.Role == tele.Left || member.Role == tele.Kicked {
					unjoinedChannels = append(unjoinedChannels, ch)
				}
			}

			if len(unjoinedChannels) > 0 {
				markup := &tele.ReplyMarkup{}
				var rows []tele.Row
				for i, ch := range unjoinedChannels {
					btn := tele.Btn{Text: fmt.Sprintf("📢 %d-kanalga obuna bo'lish", i+1), URL: ch.URL}
					rows = append(rows, markup.Row(btn))
				}
				checkBtn := tele.Btn{Text: "✅ Tasdiqlash", Data: "check_sub"}
				rows = append(rows, markup.Row(checkBtn))
				markup.Inline(rows...)

				text := "Botdan foydalanish uchun quyidagi kanallarga obuna bo'lishingiz kerak:"
				if c.Callback() != nil {
					// If it's a callback, try to edit the message or just send a new one
					return c.EditOrSend(text, markup)
				}
				return c.Send(text, markup)
			}

			return next(c)
		}
	}
}

func (b *Bot) handleCheckSubscriptionCallback(c tele.Context) error {
	// Re-run the same logic. If they joined, they pass the middleware and we can execute the original intended action.
	// However, we don't know the original action. Easiest is to send the /start menu if they successfully subscribed.
	ctx := context.Background()
	channels, err := b.channelRepo.GetAll(ctx)
	if err == nil && len(channels) > 0 {
		for _, ch := range channels {
			chat := &tele.Chat{ID: ch.ChatID}
			member, err := b.Client.ChatMemberOf(chat, c.Sender())
			if err == nil && (member.Role == tele.Left || member.Role == tele.Kicked) {
				return c.Respond(&tele.CallbackResponse{Text: "Hali barcha kanallarga obuna bo'lmadingiz!", ShowAlert: true})
			}
		}
	}

	c.Respond(&tele.CallbackResponse{Text: "Tasdiqlandi!"})
	c.Delete() // Delete the forced subscription message
	
	// Simply call handleStart to show main menu
	return b.handleStart(c)
}
