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

	// Check if resume exists
	res, err := b.resumeRepo.GetCurrentByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("Failed to get resume", "error", err)
		return c.Send("Texnik xatolik")
	}

	if res == nil {
		// Ask for full name
		b.state.Set(ctx, telegramID, user.StateResumeFullName)
		return c.Send("Rezyume yaratish uchun quyidagi ma'lumotlarni kiriting:\n\n1. Ism, familiya va sharifingizni to'liq kiriting:", &tele.ReplyMarkup{RemoveKeyboard: true})
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

func (b *Bot) handlePhoto(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return nil
	}

	state, _ := b.state.Get(ctx, telegramID)
	if state == user.StateResumePhoto {
		if c.Message().Photo == nil {
			return c.Send("Iltimos, rasm yuboring.")
		}
		photoID := c.Message().Photo.FileID
		b.state.SetData(ctx, telegramID, "resume_photo", photoID)
		b.state.Set(ctx, telegramID, user.StateResumeExperience)
		return c.Send("3. Qayerda ishlagansiz va qancha vaqt ichida?\n(Masalan: 'Artel, 2 yil' yoki 'Tajribam yo'q')")
	}

	return nil
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
	case user.StateResumeFullName:
		b.state.SetData(ctx, telegramID, "resume_fullname", text)
		b.state.Set(ctx, telegramID, user.StateResumePhoto)
		return c.Send("2. Iltimos, o'zingizning rasmingizni yuboring (majburiy):")

	case user.StateResumePhoto:
		return c.Send("Bu bosqichda faqat rasm qabul qilinadi. Iltimos, rasm yuboring.")

	case user.StateResumeExperience:
		b.state.SetData(ctx, telegramID, "resume_experience", text)
		b.state.Set(ctx, telegramID, user.StateResumeSalary)
		return c.Send("4. Oylik maoshingiz qancha bo'lishini xohlaysiz? (Faqat raqamda kiriting, masalan: 5000000)")

	case user.StateResumeSalary:
		salary, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return c.Send("Iltimos, oylik maoshni faqat raqamlar bilan kiriting.")
		}
		b.state.SetData(ctx, telegramID, "resume_salary", fmt.Sprintf("%f", salary))
		b.state.Set(ctx, telegramID, user.StateResumeAddress)
		return c.Send("5. Yashash manzilingizni kiriting (Viloyat, shahar/tuman):")

	case user.StateResumeAddress:
		b.state.SetData(ctx, telegramID, "resume_address", text)
		b.state.Set(ctx, telegramID, user.StateResumePhones)
		return c.Send("6. Qo'shimcha 2 ta telefon raqamingizni kiriting:\n(Masalan: +998901234567, +998901234568)")

	case user.StateResumePhones:
		b.state.SetData(ctx, telegramID, "resume_phones", text)
		b.state.Set(ctx, telegramID, user.StateResumeConfirm)
		
		// Show preview and ask for confirmation
		fullName, _ := b.state.GetData(ctx, telegramID, "resume_fullname")
		photoID, _ := b.state.GetData(ctx, telegramID, "resume_photo")
		experience, _ := b.state.GetData(ctx, telegramID, "resume_experience")
		address, _ := b.state.GetData(ctx, telegramID, "resume_address")
		phones, _ := b.state.GetData(ctx, telegramID, "resume_phones")
		salaryStr, _ := b.state.GetData(ctx, telegramID, "resume_salary")
		salary, _ := strconv.ParseFloat(salaryStr, 64)

		preview := fmt.Sprintf("📋 REZYUME PREVYU:\n\n👤 F.I.Sh: %s\n💼 Tajriba: %s\n💰 Maosh: %.0f UZS\n📍 Manzil: %s\n📞 Q'oshimcha raqamlar: %s\n\nBarcha ma'lumotlar to'g'rimi?", fullName, experience, salary, address, phones)
		
		btnConfirm := tele.Btn{Text: "✅ Tasdiqlash", Data: "resume_confirm"}
		btnCancel := tele.Btn{Text: "❌ Bekor qilish", Data: "resume_cancel"}
		markup := &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{*btnConfirm.Inline(), *btnCancel.Inline()},
			},
		}

		if photoID != "" {
			photo := &tele.Photo{File: tele.File{FileID: photoID}, Caption: preview}
			return c.Send(photo, markup)
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
		b.state.Set(ctx, telegramID, user.StateResumeFullName)
		c.Edit(c.Message().Text + "\n\nRezyumeni tahrirlash boshlandi.")
		return c.Send("Rezyume yaratish uchun quyidagi ma'lumotlarni kiriting:\n\n1. Ism, familiya va sharifingizni to'liq kiriting:", &tele.ReplyMarkup{RemoveKeyboard: true})
	}

	if data == "resume_cancel" {
		b.state.ClearData(ctx, telegramID)
		c.Edit(c.Message().Caption + "\n\nBekor qilindi.") // photo captions
		if c.Message().Text != "" {
			c.Edit(c.Message().Text + "\n\nBekor qilindi.")
		}
		return b.sendMainMenu(c, u)
	}

	if data == "resume_confirm" {
		fullName, _ := b.state.GetData(ctx, telegramID, "resume_fullname")
		photoID, _ := b.state.GetData(ctx, telegramID, "resume_photo")
		experience, _ := b.state.GetData(ctx, telegramID, "resume_experience")
		address, _ := b.state.GetData(ctx, telegramID, "resume_address")
		phones, _ := b.state.GetData(ctx, telegramID, "resume_phones")
		salaryStr, _ := b.state.GetData(ctx, telegramID, "resume_salary")
		salary, _ := strconv.ParseFloat(salaryStr, 64)

		parts := strings.Split(phones, ",")
		var phone1, phone2 *string
		if len(parts) > 0 {
			p1 := strings.TrimSpace(parts[0])
			phone1 = &p1
		}
		if len(parts) > 1 {
			p2 := strings.TrimSpace(parts[1])
			phone2 = &p2
		}

		photoPtr := &photoID
		if photoID == "" {
			photoPtr = nil
		}

		res := &resume.Resume{
			UserID:         u.ID,
			FirstName:      fullName,
			LastName:       "", // not used separately anymore
			PhotoFileID:    photoPtr,
			AddressRegion:  address,
			ExpectedSalary: salary,
			SalaryCurrency: "UZS",
			SkillsText:     experience,
			ExtraPhone1:    phone1,
			ExtraPhone2:    phone2,
		}

		if err := b.resumeRepo.Create(ctx, res); err != nil {
			slog.Error("Failed to save resume", "error", err)
			return c.Send("Texnik xatolik")
		}

		b.state.ClearData(ctx, telegramID)
		if c.Message().Caption != "" {
			c.Edit(c.Message().Caption + "\n\nTasdiqlandi.")
		} else {
			c.Edit(c.Message().Text + "\n\nTasdiqlandi.")
		}
		
		c.Send("Rezyume saqlandi!", &tele.ReplyMarkup{RemoveKeyboard: true})
		return b.sendMainMenu(c, u)
	}

	return nil
}
