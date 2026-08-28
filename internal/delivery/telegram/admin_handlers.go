package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/company/hrbot/internal/domain/user"
	"github.com/company/hrbot/internal/domain/vacancy"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) handleAdminMenu(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	isAdmin := false
	for _, id := range b.adminIDs {
		if id == telegramID {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return c.Send("Siz admin emassiz.")
	}

	b.state.Set(ctx, telegramID, user.AdminStateMainMenu)

	btnNewVacancy := tele.ReplyButton{Text: "➕ Yangi ish o'rni"}
	btnViewApps := tele.ReplyButton{Text: "📋 Arizalarni ko'rish"}
	btnBroadcast := tele.ReplyButton{Text: "📢 Xabar yuborish"}
	btnBack := tele.ReplyButton{Text: "🔙 Orqaga (Asosiy menyu)"}

	markup := &tele.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{
			{btnNewVacancy, btnViewApps},
			{btnBroadcast},
			{btnBack},
		},
		ResizeKeyboard: true,
	}

	return c.Send("🛠 Admin Paneliga xush kelibsiz. Nima qilamiz?", markup)
}

func (b *Bot) handleAdminText(c tele.Context, state string) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	text := c.Message().Text

	if text == "🔙 Orqaga (Asosiy menyu)" {
		u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
		return b.sendMainMenu(c, u)
	}

	if state == user.AdminStateMainMenu {
		switch text {
		case "➕ Yangi ish o'rni":
			b.state.Set(ctx, telegramID, user.AdminStateAddVacancyTitle)
			return c.Send("Yangi ish o'rni nomini kiriting (Masalan: Sotuvchi):", &tele.ReplyMarkup{RemoveKeyboard: true})
		case "📋 Arizalarni ko'rish":
			return b.handleAdminViewApplications(c)
		case "📢 Xabar yuborish":
			b.state.Set(ctx, telegramID, user.AdminStateBroadcastMessage)
			return c.Send("Barcha foydalanuvchilarga yuboriladigan xabar matnini kiriting:", &tele.ReplyMarkup{RemoveKeyboard: true})
		}
		return nil
	}

	switch state {
	case user.AdminStateAddVacancyTitle:
		b.state.SetData(ctx, telegramID, "vac_title", text)
		b.state.Set(ctx, telegramID, user.AdminStateAddVacancyLocation)
		return c.Send("Ish joyi manzilini kiriting (Masalan: Samarqand):")
	
	case user.AdminStateAddVacancyLocation:
		b.state.SetData(ctx, telegramID, "vac_loc", text)
		b.state.Set(ctx, telegramID, user.AdminStateAddVacancySalary)
		return c.Send("Taklif qilinayotgan maoshni raqamda kiriting (Masalan: 5000000 yoki kelishilgan):")
	
	case user.AdminStateAddVacancySalary:
		b.state.SetData(ctx, telegramID, "vac_salary", text)
		b.state.Set(ctx, telegramID, user.AdminStateAddVacancyDesc)
		return c.Send("Ish haqida to'liq ma'lumotni kiriting (Talablar, vazifalar va h.k):")
	
	case user.AdminStateAddVacancyDesc:
		title, _ := b.state.GetData(ctx, telegramID, "vac_title")
		loc, _ := b.state.GetData(ctx, telegramID, "vac_loc")
		salaryText, _ := b.state.GetData(ctx, telegramID, "vac_salary")

		currency := "UZS"
		dept := "General"
		empType := "To'liq"
		schedule := "Dushanba-Juma"
		var salaryFrom *float64

		v := &vacancy.Vacancy{
			Title:          title,
			Slug:           strings.ToLower(strings.ReplaceAll(title, " ", "-")),
			Department:     &dept,
			Location:       &loc,
			EmploymentType: &empType,
			Schedule:       &schedule,
			SalaryFrom:     salaryFrom,
			SalaryTo:       nil,
			SalaryCurrency: &currency,
			SalaryText:     &salaryText,
			Description:    &text,
			Status:         vacancy.StatusPublished,
		}

		if err := b.vacancyRepo.Create(ctx, v); err != nil {
			return c.Send("Xatolik yuz berdi: " + err.Error())
		}
		
		b.state.ClearData(ctx, telegramID)
		u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
		c.Send("Yangi ish o'rni muvaffaqiyatli saqlandi va e'lon qilindi! ✅")
		return b.sendMainMenu(c, u)

	case user.AdminStateBroadcastMessage:
		users, _ := b.userRepo.GetAll(ctx)
		count := 0
		for _, u := range users {
			if u.TelegramID != telegramID {
				_, _ = b.Client.Send(&tele.User{ID: u.TelegramID}, text)
				count++
			}
		}
		b.state.ClearData(ctx, telegramID)
		u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
		c.Send(fmt.Sprintf("Xabar %d ta foydalanuvchiga yuborildi! ✅", count))
		return b.sendMainMenu(c, u)
	}

	return nil
}

func (b *Bot) handleAdminViewApplications(c tele.Context) error {
	ctx := context.Background()
	
	apps, err := b.appRepo.GetAll(ctx, 20, 0)
	if err != nil || len(apps) == 0 {
		return c.Send("Hozircha arizalar yo'q.")
	}

	for _, app := range apps {
		v, _ := b.vacancyRepo.GetByID(ctx, app.VacancyID)
		res, _ := b.resumeRepo.GetByID(ctx, app.ResumeID)
		
		title := "Noma'lum vakansiya"
		if v != nil {
			title = v.Title
		}

		candidate := "Noma'lum nomzod"
		experience := "Yo'q"
		address := "Yo'q"
		var p1, p2 string
		if res != nil {
			candidate = res.FirstName
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

		text := fmt.Sprintf("💼 Vakansiya: %s\n👤 Nomzod: %s\n💼 Tajriba: %s\n📍 Manzil: %s\n📞 Qo'shimcha telefon: %s\n📅 Vaqt: %s\nHolati: %s", 
			title, candidate, experience, address, phones, app.SubmittedAt.Format("02-Jan-2006 15:04"), app.Status)
		
		if res != nil && res.PhotoFileID != nil && *res.PhotoFileID != "" {
			photo := &tele.Photo{File: tele.File{FileID: *res.PhotoFileID}, Caption: text}
			c.Send(photo)
		} else {
			c.Send(text)
		}
	}
	return nil
}
