package telegram

import (
	"context"
	"log/slog"

	"github.com/company/hrbot/internal/domain/user"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) handleStart(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	// Check if user exists
	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		return c.Send("Texnik xatolik / Техническая ошибка / Technical error")
	}

	if u == nil {
		// New user, ask for language
		btnUz := tele.Btn{Text: b.i18n.Get("uz", "btn_uz"), Data: "lang_uz"}
		btnRu := tele.Btn{Text: b.i18n.Get("ru", "btn_ru"), Data: "lang_ru"}
		btnEn := tele.Btn{Text: b.i18n.Get("en", "btn_en"), Data: "lang_en"}

		markup := &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{*btnUz.Inline(), *btnRu.Inline(), *btnEn.Inline()},
			},
		}

		b.state.Set(ctx, telegramID, user.StateSelectingLanguage)
		return c.Send(b.i18n.Get("uz", "greeting"), markup)
	}

	// User exists, check if phone exists
	if u.PrimaryPhone == nil || *u.PrimaryPhone == "" {
		return b.askContact(c, u)
	}

	// Send main menu
	return b.sendMainMenu(c, u)
}

func (b *Bot) handleLanguageCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	state, _ := b.state.Get(ctx, telegramID)
	if state != user.StateSelectingLanguage {
		// Ignore if not in language selection state, or maybe they just wanted to change language
		// We can allow changing language anytime, but let's stick to flow
	}

	data := c.Callback().Data
	var langCode string
	switch data {
	case "lang_uz":
		langCode = "uz"
	case "lang_ru":
		langCode = "ru"
	case "lang_en":
		langCode = "en"
	default:
		langCode = "uz"
	}

	// Create user with just language first
	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	if u == nil {
		u = &user.User{
			TelegramID:        telegramID,
			TelegramUsername:  &c.Sender().Username,
			TelegramFirstName: &c.Sender().FirstName,
			TelegramLastName:  &c.Sender().LastName,
			LanguageCode:      &langCode,
			Status:            "active",
		}
		if err := b.userRepo.Create(ctx, u); err != nil {
			slog.Error("Failed to create user", "error", err)
			return c.Send("Texnik xatolik")
		}
	} else {
		u.LanguageCode = &langCode
		b.userRepo.Update(ctx, u)
	}

	// Remove inline keyboard
	c.Edit(c.Message().Text)

	// Acknowledge callback
	c.Respond()

	// Ask for contact
	return b.askContact(c, u)
}

func (b *Bot) askContact(c tele.Context, u *user.User) error {
	ctx := context.Background()
	lang := "uz"
	if u.LanguageCode != nil {
		lang = *u.LanguageCode
	}

	b.state.Set(ctx, u.TelegramID, user.StateWaitingContact)

	btnContact := tele.ReplyButton{Text: b.i18n.Get(lang, "btn_send_contact"), Contact: true}
	markup := &tele.ReplyMarkup{
		ReplyKeyboard:       [][]tele.ReplyButton{{btnContact}},
		ResizeKeyboard:      true,
		OneTimeKeyboard:     true,
	}

	return c.Send(b.i18n.Get(lang, "ask_contact"), markup)
}

func (b *Bot) handleContact(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	state, _ := b.state.Get(ctx, telegramID)
	if state != user.StateWaitingContact {
		return nil // Ignore if not waiting for contact
	}

	contact := c.Message().Contact
	if contact == nil || contact.UserID != telegramID {
		return c.Send("Iltimos, faqat o'z raqamingizni pastdagi tugma orqali yuboring.")
	}

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Send("Profil topilmadi. Iltimos, /start buyrug'ini yuboring.")
	}

	// Save phone
	phone := contact.PhoneNumber
	// Basic normalization, ensure it has '+' if missing
	if len(phone) > 0 && phone[0] != '+' {
		phone = "+" + phone
	}
	
	u.PrimaryPhone = &phone
	if err := b.userRepo.Update(ctx, u); err != nil {
		slog.Error("Failed to update user phone", "error", err)
		return c.Send("Texnik xatolik")
	}

	lang := *u.LanguageCode
	c.Send(b.i18n.Get(lang, "contact_received"), &tele.ReplyMarkup{RemoveKeyboard: true})

	return b.sendMainMenu(c, u)
}

func (b *Bot) sendMainMenu(c tele.Context, u *user.User) error {
	ctx := context.Background()
	lang := *u.LanguageCode

	b.state.Set(ctx, u.TelegramID, user.StateMainMenu)

	btnVacancies := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_vacancies")}
	btnResume := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_resume")}
	btnProfile := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_profile")}
	btnApplications := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_applications")}
	btnAbout := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_about")}
	btnFAQ := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_faq")}
	btnContact := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_contact")}
	btnSettings := tele.ReplyButton{Text: b.i18n.Get(lang, "menu_settings")}

	markup := &tele.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{
			{btnVacancies, btnResume},
			{btnProfile, btnApplications},
			{btnAbout, btnFAQ},
			{btnContact, btnSettings},
		},
		ResizeKeyboard: true,
	}

	return c.Send(b.i18n.Get(lang, "main_menu"), markup)
}
