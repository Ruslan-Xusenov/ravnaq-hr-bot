package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/company/hrbot/internal/domain/resume"
	"github.com/company/hrbot/internal/domain/user"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) handleResumeMenu(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Send("Texnik xatolik")
	}

	lang := *u.LanguageCode

	// Check if resume exists
	res, err := b.resumeRepo.GetCurrentByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("Failed to get resume", "error", err)
		return c.Send("Texnik xatolik")
	}

	if res == nil {
		// Ask for first name
		b.state.Set(ctx, telegramID, user.StateResumeFirstName)
		return c.Send(b.i18n.Get(lang, "ask_first_name"), &tele.ReplyMarkup{RemoveKeyboard: true})
	}

	// Show existing resume options
	btnEdit := tele.Btn{Text: "📝 Tahrirlash", Data: "resume_edit"}
	markup := &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*btnEdit.Inline()},
		},
	}
	return c.Send("Sizning rezyumeyingiz mavjud. Pastdagi tugma orqali tahrirlashingiz mumkin.", markup)
}

func (b *Bot) handleText(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	text := c.Message().Text

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return nil
	}

	lang := *u.LanguageCode
	state, _ := b.state.Get(ctx, telegramID)

	// Handle Admin states and "🛠 Admin Panel" button
	if text == "🛠 Admin Panel" {
		return b.handleAdminMenu(c)
	}

	if strings.HasPrefix(state, "admin_") {
		return b.handleAdminText(c, state)
	}

	// Handle Main Menu buttons
	if state == user.StateMainMenu {
		switch text {
		case b.i18n.Get(lang, "menu_resume"):
			return b.handleResumeMenu(c)
		case b.i18n.Get(lang, "menu_vacancies"):
			return b.handleVacanciesMenu(c)
		case b.i18n.Get(lang, "menu_applications"):
			return b.handleMyApplications(c)
		case b.i18n.Get(lang, "menu_profile"):
			return b.handleProfileMenu(c)
		case b.i18n.Get(lang, "menu_about"):
			return b.handleAbout(c)
		case b.i18n.Get(lang, "menu_faq"):
			return b.handleFAQ(c)
		case b.i18n.Get(lang, "menu_contact"):
			return b.handleContactUs(c)
		case b.i18n.Get(lang, "menu_settings"):
			return b.handleSettings(c)
		}
		return nil
	}

	// Handle Resume Flow
	switch state {
	case user.StateResumeFirstName:
		b.state.SetData(ctx, telegramID, "resume_first_name", text)
		b.state.Set(ctx, telegramID, user.StateResumeLastName)
		return c.Send(b.i18n.Get(lang, "ask_last_name"))

	case user.StateResumeLastName:
		b.state.SetData(ctx, telegramID, "resume_last_name", text)
		b.state.Set(ctx, telegramID, user.StateResumeRegion)
		return c.Send(b.i18n.Get(lang, "ask_region"))

	case user.StateResumeRegion:
		b.state.SetData(ctx, telegramID, "resume_region", text)
		b.state.Set(ctx, telegramID, user.StateResumeSalary)
		return c.Send(b.i18n.Get(lang, "ask_salary"))

	case user.StateResumeSalary:
		salary, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return c.Send(b.i18n.Get(lang, "err_invalid_number"))
		}
		b.state.SetData(ctx, telegramID, "resume_salary", fmt.Sprintf("%f", salary))
		b.state.Set(ctx, telegramID, user.StateResumeConfirm)
		
		// Show preview and ask for confirmation
		firstName, _ := b.state.GetData(ctx, telegramID, "resume_first_name")
		lastName, _ := b.state.GetData(ctx, telegramID, "resume_last_name")
		region, _ := b.state.GetData(ctx, telegramID, "resume_region")

		preview := fmt.Sprintf("Ism: %s\nFamiliya: %s\nViloyat: %s\nKutilayotgan maosh: %f", firstName, lastName, region, salary)
		
		btnConfirm := tele.Btn{Text: b.i18n.Get(lang, "btn_confirm"), Data: "resume_confirm"}
		btnCancel := tele.Btn{Text: b.i18n.Get(lang, "btn_cancel"), Data: "resume_cancel"}
		markup := &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{*btnConfirm.Inline(), *btnCancel.Inline()},
			},
		}

		return c.Send(preview, markup)
	}

	return nil
}

func (b *Bot) handleResumeCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	data := c.Callback().Data

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Send("User not found")
	}

	if data == "resume_edit" {
		b.state.Set(ctx, telegramID, user.StateResumeFirstName)
		c.Edit(c.Message().Text + "\n\nRezyumeni tahrirlash boshlandi.")
		return c.Send(b.i18n.Get(*u.LanguageCode, "ask_first_name"), &tele.ReplyMarkup{RemoveKeyboard: true})
	}

	if data == "resume_cancel" {
		b.state.ClearData(ctx, telegramID)
		c.Edit(c.Message().Text + "\n\nBekor qilindi.")
		return b.sendMainMenu(c, u)
	}

	if data == "resume_confirm" {
		firstName, _ := b.state.GetData(ctx, telegramID, "resume_first_name")
		lastName, _ := b.state.GetData(ctx, telegramID, "resume_last_name")
		region, _ := b.state.GetData(ctx, telegramID, "resume_region")
		salaryStr, _ := b.state.GetData(ctx, telegramID, "resume_salary")
		salary, _ := strconv.ParseFloat(salaryStr, 64)

		res := &resume.Resume{
			UserID:         u.ID,
			FirstName:      firstName,
			LastName:       lastName,
			AddressRegion:  region,
			ExpectedSalary: salary,
			SalaryCurrency: "UZS",
		}

		if err := b.resumeRepo.Create(ctx, res); err != nil {
			slog.Error("Failed to save resume", "error", err)
			return c.Send("Texnik xatolik")
		}

		b.state.ClearData(ctx, telegramID)
		c.Edit(c.Message().Text + "\n\nTasdiqlandi.")
		
		c.Send("Rezyume saqlandi!", &tele.ReplyMarkup{RemoveKeyboard: true})
		return b.sendMainMenu(c, u)
	}

	return nil
}
