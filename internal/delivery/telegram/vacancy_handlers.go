package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/company/hrbot/internal/domain/application"
	"github.com/company/hrbot/internal/domain/user"
	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/google/uuid"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) handleVacanciesMenu(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Send("Bot ma'lumotlar bazasi yangilangani sababli profilingiz topilmadi. Iltimos, /start buyrug'ini yuboring.")
	}
	lang := *u.LanguageCode

	vacancies, err := b.vacancyRepo.GetActive(ctx, 5, 0) // Limit to 5 for now
	if err != nil {
		slog.Error("Failed to get vacancies", "error", err)
		return c.Send("Texnik xatolik")
	}

	if len(vacancies) == 0 {
		return c.Send(b.i18n.Get(lang, "no_vacancies"))
	}

	markup := b.buildVacanciesListMarkup(vacancies)
	return c.Send("Faol vakansiyalar (Batafsil ma'lumot uchun tanlang):", markup)
}

func (b *Bot) buildVacanciesListMarkup(vacancies []vacancy.Vacancy) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, v := range vacancies {
		btn := tele.Btn{Text: "💼 " + v.Title, Data: "view_vac_" + v.ID.String()}
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)
	return markup
}

func (b *Bot) handleBackToVacanciesCallback(c tele.Context) error {
	ctx := context.Background()
	vacancies, err := b.vacancyRepo.GetActive(ctx, 10, 0)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Xatolik yuz berdi", ShowAlert: true})
	}

	if len(vacancies) == 0 {
		return c.Edit("Hozirda faol vakansiyalar yo'q.")
	}

	markup := b.buildVacanciesListMarkup(vacancies)
	return c.Edit("Faol vakansiyalar (Batafsil ma'lumot uchun tanlang):", markup)
}

func (b *Bot) handleApplyCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	data := c.Callback().Data

	// Extract vacancy ID from "apply_uuid"
	vacancyIDStr := data[len("apply_"):]
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		return c.Send("Xatolik")
	}

	u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
	lang := *u.LanguageCode

	// Check if user has a resume
	res, err := b.resumeRepo.GetCurrentByUserID(ctx, u.ID)
	if err != nil {
		return c.Send("Texnik xatolik")
	}
	if res == nil {
		c.Respond(&tele.CallbackResponse{Text: b.i18n.Get(lang, "err_no_resume"), ShowAlert: true})
		return nil
	}

	// Check if already applied
	hasApplied, err := b.appRepo.HasApplied(ctx, u.ID, vacancyID)
	if err != nil {
		return c.Send("Texnik xatolik")
	}
	if hasApplied {
		c.Respond(&tele.CallbackResponse{Text: b.i18n.Get(lang, "err_already_applied"), ShowAlert: true})
		return nil
	}

	phones := ""
	if res.ExtraPhone1 != nil {
		phones = *res.ExtraPhone1
	}
	if res.ExtraPhone2 != nil {
		if phones != "" {
			phones += ", "
		}
		phones += *res.ExtraPhone2
	}
	if phones == "" {
		phones = "Yo'q"
	}
	preview := fmt.Sprintf("📋 SIZNING REZYUMEYINGIZ:\n\n👤 F.I.Sh: %s\n💼 Tajriba: %s\n💰 Maosh: %.0f %s\n📍 Manzil: %s\n📞 Qo'shimcha raqamlar: %s\n\nShu arizani yuborasizmi yoki qayta rezyume to'ldirasizmi?", 
		res.FirstName, res.SkillsText, res.ExpectedSalary, res.SalaryCurrency, res.AddressRegion, phones)
	
	btnConfirm := tele.Btn{Text: "✅ Shu rezyumeni yuborish", Data: "confirm_apply_" + vacancyID.String()}
	btnNew := tele.Btn{Text: "📝 Yangi rezyume to'ldirish", Data: "new_resume_apply_" + vacancyID.String()}
	markup := &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*btnConfirm.Inline()},
			{*btnNew.Inline()},
		},
	}

	c.Respond()
	if res.PhotoFileID != nil && *res.PhotoFileID != "" {
		photo := &tele.Photo{File: tele.File{FileID: *res.PhotoFileID}, Caption: preview}
		return c.Send(photo, markup)
	}
	return c.Send(preview, markup)
}

func (b *Bot) handleConfirmApplyCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	data := c.Callback().Data
	vacancyIDStr := data[len("confirm_apply_"):]
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		return c.Send("Xatolik")
	}

	u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
	res, _ := b.resumeRepo.GetCurrentByUserID(ctx, u.ID)

	app := &application.Application{
		UserID:    u.ID,
		VacancyID: vacancyID,
		ResumeID:  res.ID,
	}

	if err := b.appRepo.Create(ctx, app); err != nil {
		slog.Error("Failed to apply", "error", err)
		return c.Send("Texnik xatolik")
	}

	c.Edit(c.Message().Caption + "\n\n✅ Arizangiz muvaffaqiyatli yuborildi!")
	if c.Message().Text != "" {
		c.Edit(c.Message().Text + "\n\n✅ Arizangiz muvaffaqiyatli yuborildi!")
	}
	c.Respond()

	// Notify admins
	for _, adminID := range b.adminIDs {
		adminChat := &tele.Chat{ID: adminID}
		name := ""
		if u.TelegramFirstName != nil {
			name = *u.TelegramFirstName
		}
		if u.TelegramLastName != nil {
			name += " " + *u.TelegramLastName
		}
		if name == "" {
			name = "Noma'lum"
		}
		btnApprove := tele.Btn{Text: "✅ Tasdiqlash", Data: "admin_apprv_app_" + app.ID.String()}
		btnReject := tele.Btn{Text: "❌ Rad etish", Data: "admin_rejct_app_" + app.ID.String()}
		markup := &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{*btnApprove.Inline(), *btnReject.Inline()},
			},
		}
		v, _ := b.vacancyRepo.GetByID(ctx, app.VacancyID)
		res, _ := b.resumeRepo.GetByID(ctx, app.ResumeID)
		
		title := "Noma'lum vakansiya"
		if v != nil {
			title = v.Title
		}
		
		experience := "Yo'q"
		address := "Yo'q"
		var p1, p2 string
		if res != nil {
			if res.SkillsText != "" {
				experience = res.SkillsText
			}
			if res.AddressRegion != "" {
				address = res.AddressRegion
			}
			if res.ExtraPhone1 != nil {
				p1 = *res.ExtraPhone1
			}
			if res.ExtraPhone2 != nil {
				p2 = *res.ExtraPhone2
			}
		}

		phones := p1
		if p2 != "" {
			phones += ", " + p2
		}
		if phones == "" {
			phones = "Yo'q"
		}

		text := fmt.Sprintf("🔔 Yangi ariza keldi!\n\n💼 Vakansiya: %s\n👤 Nomzod: %s\n💼 Tajriba: %s\n📍 Manzil: %s\n📞 Qo'shimcha telefon: %s\n📅 Vaqt: %s\nHolati: %s", 
			title, name, experience, address, phones, app.SubmittedAt.Format("02-Jan-2006 15:04"), app.Status)

		if res != nil && res.PhotoFileID != nil && *res.PhotoFileID != "" {
			photo := &tele.Photo{File: tele.File{FileID: *res.PhotoFileID}, Caption: text}
			b.Client.Send(adminChat, photo, markup)
		} else {
			b.Client.Send(adminChat, text, markup)
		}
	}

	return nil
}

func (b *Bot) handleNewResumeApplyCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	data := c.Callback().Data
	vacancyIDStr := data[len("new_resume_apply_"):]

	// Set pending vacancy apply in state
	b.state.SetData(ctx, telegramID, "pending_apply_vacancy", vacancyIDStr)

	// Start resume flow
	b.state.Set(ctx, telegramID, user.StateResumeFullName)
	c.Edit(c.Message().Caption + "\n\nYangi rezyume to'ldirish boshlandi.")
	if c.Message().Text != "" {
		c.Edit(c.Message().Text + "\n\nYangi rezyume to'ldirish boshlandi.")
	}
	c.Respond()

	return c.Send("Rezyume yaratish uchun quyidagi ma'lumotlarni kiriting:\n\n1. Ism, familiya va sharifingizni to'liq kiriting:", &tele.ReplyMarkup{RemoveKeyboard: true})
}

func (b *Bot) handleMyApplications(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Send("Bot ma'lumotlar bazasi yangilangani sababli profilingiz topilmadi. Iltimos, /start buyrug'ini yuboring.")
	}
	lang := *u.LanguageCode

	apps, err := b.appRepo.GetByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("Failed to get applications", "error", err)
		return c.Send("Texnik xatolik")
	}

	if len(apps) == 0 {
		return c.Send(b.i18n.Get(lang, "no_applications"))
	}

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, app := range apps {
		vac, _ := b.vacancyRepo.GetByID(ctx, app.VacancyID)
		title := "Noma'lum"
		if vac != nil {
			title = vac.Title
		}

		statusIcon := "⏳"
		if app.Status == "rejected" {
			statusIcon = "❌"
		} else if app.Status == "approved" {
			statusIcon = "✅"
		}

		btnText := fmt.Sprintf("💼 %s (%s)", title, statusIcon)
		btn := tele.Btn{Text: btnText, Data: "view_app_" + app.ID.String()}
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)

	return c.Send("Sizning arizalaringiz quyidagilar. Batafsil ma'lumot uchun ustiga bosing:", markup)
}

func (b *Bot) handleViewApplicationCallback(c tele.Context) error {
	ctx := context.Background()
	data := c.Callback().Data
	appIDStr := data[len("view_app_"):]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Xato ID", ShowAlert: true})
	}

	app, err := b.appRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ariza topilmadi", ShowAlert: true})
	}

	vac, _ := b.vacancyRepo.GetByID(ctx, app.VacancyID)
	title := "Noma'lum"
	if vac != nil {
		title = vac.Title
	}

	statusText := app.Status
	if app.Status == "submitted" {
		statusText = "⏳ Ko'rib chiqilmoqda"
	} else if app.Status == "rejected" {
		statusText = "Ko'rib chiqildi" // As user requested in screenshot
	} else if app.Status == "approved" {
		statusText = "✅ Qabul qilingan"
	}

	text := fmt.Sprintf("💼 <b>Vakansiya: %s</b>\n\n📊 Holati: %s\n🕒 Topshirilgan vaqt: %s",
		title, statusText, app.SubmittedAt.Format("02.01.2006 15:04"))

	markup := &tele.ReplyMarkup{}
	backBtn := markup.Data("⬅️ Orqaga", "back_to_apps")
	markup.Inline(markup.Row(backBtn))

	c.Edit(text, tele.ModeHTML, markup)
	return c.Respond()
}

func (b *Bot) handleBackToApplicationsCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Profil topilmadi", ShowAlert: true})
	}

	apps, err := b.appRepo.GetByUserID(ctx, u.ID)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Xatolik", ShowAlert: true})
	}

	if len(apps) == 0 {
		c.Edit("Sizda arizalar yo'q.")
		return c.Respond()
	}

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, app := range apps {
		vac, _ := b.vacancyRepo.GetByID(ctx, app.VacancyID)
		title := "Noma'lum"
		if vac != nil {
			title = vac.Title
		}

		statusIcon := "⏳"
		if app.Status == "rejected" {
			statusIcon = "❌"
		} else if app.Status == "approved" {
			statusIcon = "✅"
		}

		btnText := fmt.Sprintf("💼 %s (%s)", title, statusIcon)
		btn := tele.Btn{Text: btnText, Data: "view_app_" + app.ID.String()}
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)

	c.Edit("Sizning arizalaringiz quyidagilar. Batafsil ma'lumot uchun ustiga bosing:", markup)
	return c.Respond()
}

func (b *Bot) handleViewVacancyCallback(c tele.Context) error {
	ctx := context.Background()
	data := c.Callback().Data
	vIDStr := data[len("view_vac_"):]
	vID, err := uuid.Parse(vIDStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Noto'g'ri vakansiya ID"})
	}

	v, err := b.vacancyRepo.GetByID(ctx, vID)
	if err != nil || v == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Vakansiya topilmadi"})
	}

	salaryDisplay := "Kelishiladi"
	if v.SalaryText != nil && *v.SalaryText != "" {
		salaryDisplay = *v.SalaryText
	} else if v.SalaryFrom != nil && *v.SalaryFrom > 0 {
		if v.SalaryTo != nil && *v.SalaryTo > 0 {
			salaryDisplay = fmt.Sprintf("%.0f - %.0f %s", *v.SalaryFrom, *v.SalaryTo, *v.SalaryCurrency)
		} else {
			salaryDisplay = fmt.Sprintf("%.0f %s dan", *v.SalaryFrom, *v.SalaryCurrency)
		}
	}

	dept := "Noma'lum"
	if v.Department != nil && *v.Department != "" {
		dept = *v.Department
	}

	empType := "To'liq"
	if v.EmploymentType != nil && *v.EmploymentType != "" {
		empType = *v.EmploymentType
	}

	schedule := "Belgilanmagan"
	if v.Schedule != nil && *v.Schedule != "" {
		schedule = *v.Schedule
	}

	text := fmt.Sprintf("🏢 <b>Bo'lim:</b> %s\n💼 <b>Vakansiya:</b> %s\n\n📍 <b>Manzil:</b> %s\n💰 <b>Maosh:</b> %s\n⏰ <b>Ish vaqti:</b> %s\n📊 <b>Bandlik:</b> %s\n\n📝 <b>Batafsil ma'lumot:</b>\n%s",
		dept, v.Title, *v.Location, salaryDisplay, schedule, empType, *v.Description)

	// Get language
	telegramID := c.Sender().ID
	u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
	lang := "uz"
	if u != nil && u.LanguageCode != nil {
		lang = *u.LanguageCode
	}

	btnApply := tele.Btn{Text: b.i18n.Get(lang, "btn_apply"), Data: "apply_" + v.ID.String()}
	btnBack := tele.Btn{Text: "🔙 Orqaga", Data: "back_to_vacancies"}
	
	markup := &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*btnApply.Inline()},
			{*btnBack.Inline()},
		},
	}

	return c.Edit(text, markup, tele.ModeHTML)
}
