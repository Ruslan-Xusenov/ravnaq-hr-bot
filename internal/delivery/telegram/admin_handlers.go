package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/company/hrbot/internal/domain/user"
	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/company/hrbot/internal/domain/application"
	"github.com/company/hrbot/internal/domain/channel"
	"github.com/google/uuid"
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

	return c.Send("🛠 Admin Paneliga xush kelibsiz. Nima qilamiz?", b.adminMenuMarkup())
}

func (b *Bot) adminMenuMarkup() *tele.ReplyMarkup {
	btnNewVacancy := tele.ReplyButton{Text: "➕ Yangi ish o'rni"}
	btnViewApps := tele.ReplyButton{Text: "📋 Arizalarni ko'rish"}
	btnBroadcast := tele.ReplyButton{Text: "📢 Xabar yuborish"}
	btnEditTexts := tele.ReplyButton{Text: "📝 Matnlarni tahrirlash"}
	btnChannels := tele.ReplyButton{Text: "📢 Kanallarni sozlash"}
	btnUsersCount := tele.ReplyButton{Text: "👥 Foydalanuvchilar soni"}
	btnActiveVacancies := tele.ReplyButton{Text: "💼 Aktiv ish o'rinlari"}
	btnBack := tele.ReplyButton{Text: "🔙 Orqaga (Asosiy menyu)"}

	return &tele.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{
			{btnNewVacancy, btnActiveVacancies},
			{btnViewApps, btnUsersCount},
			{btnBroadcast, btnEditTexts},
			{btnChannels},
			{btnBack},
		},
		ResizeKeyboard: true,
	}
}

func (b *Bot) handleAdminText(c tele.Context, state string) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	text := c.Message().Text

	if text == "🔙 Orqaga (Asosiy menyu)" {
		u, _ := b.userRepo.GetByTelegramID(ctx, telegramID)
		return b.sendMainMenu(c, u)
	}

	if text == "📢 Kanallarni sozlash" {
		return b.handleAdminChannels(c)
	}

	if text == "➕ Kanal qo'shish" {
		b.state.Set(ctx, telegramID, user.AdminStateAddChannel)
		return c.Send("Kanal ID yoki username ni kiriting (Masalan: -1001234567890 yoki @kanalnomi).\nEslatma: Botni avval ushbu kanalga admin qilishingiz shart!", &tele.ReplyMarkup{RemoveKeyboard: true})
	}

	if state == user.AdminStateMainMenu {
		switch text {
		case "➕ Yangi ish o'rni":
			b.state.Set(ctx, telegramID, user.AdminStateAddVacancyTitle)
			return c.Send("Yangi ish o'rni nomini kiriting (Masalan: Sotuvchi):", &tele.ReplyMarkup{RemoveKeyboard: true})
		case "💼 Aktiv ish o'rinlari":
			return b.handleVacanciesMenu(c)
		case "📋 Arizalarni ko'rish":
			return b.handleAdminViewApplications(c)
		case "👥 Foydalanuvchilar soni":
			count, err := b.userRepo.Count(ctx)
			if err != nil {
				return c.Send("Xatolik yuz berdi. Iltimos keyinroq urinib ko'ring.")
			}
			return c.Send(fmt.Sprintf("📊 Jami foydalanuvchilar soni: %d ta", count))
		case "📢 Xabar yuborish":
			b.state.Set(ctx, telegramID, user.AdminStateBroadcastMessage)
			return c.Send("Barcha foydalanuvchilarga yuboriladigan xabarni kiriting (rasm/video/matn):", &tele.ReplyMarkup{RemoveKeyboard: true})
		
		case "📝 Matnlarni tahrirlash":
			btnAbout := tele.Btn{Text: "🏢 Biz haqimizda", Data: "admin_edit_text_about"}
			btnFaq := tele.Btn{Text: "❓ Ko'p beriladigan savollar", Data: "admin_edit_text_faq"}
			btnContact := tele.Btn{Text: "📞 Aloqa", Data: "admin_edit_text_contact"}
			markup := &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{*btnAbout.Inline()},
					{*btnFaq.Inline()},
					{*btnContact.Inline()},
				},
			}
			return c.Send("Qaysi bo'lim matnini tahrirlamoqchisiz?", markup)

		default:
			return c.Send("Noma'lum buyruq.")
		}
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

		baseSlug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
		uniqueSlug := fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])

		v := &vacancy.Vacancy{
			Title:          title,
			Slug:           uniqueSlug,
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
		b.state.Set(ctx, telegramID, user.AdminStateMainMenu)
		return c.Send(fmt.Sprintf("Xabar %d ta foydalanuvchiga yuborildi.", count), b.adminMenuMarkup())
	
	case user.AdminStateEditTextInput:
		sectionID, _ := b.state.GetData(ctx, telegramID, "admin_edit_text_id")
		if sectionID != "" {
			err := b.textRepo.Set(ctx, sectionID, text)
			if err != nil {
				return c.Send("Xatolik yuz berdi: " + err.Error())
			}
			b.state.Set(ctx, telegramID, user.AdminStateMainMenu)
			return c.Send("✅ Matn muvaffaqiyatli saqlandi!", b.adminMenuMarkup())
		}
		return c.Send("Xatolik: Tahrirlash bo'limi aniqlanmadi.")
		
	case user.AdminStateAddChannel:
		// Attempt to parse chat to verify bot is admin
		var chat *tele.Chat
		var err error
		chatIDInt, parseErr := strconv.ParseInt(text, 10, 64)
		if parseErr == nil {
			chat, err = b.Client.ChatByID(chatIDInt)
		} else {
			chat, err = b.Client.ChatByUsername(text)
		}

		if err != nil {
			return c.Send("Kanal topilmadi! Kanal ID si yoki username'ini to'g'ri kiritganingizga va bot u yerda admin ekanligiga ishonch hosil qiling. Boshqatdan kiriting:")
		}
		
		b.state.SetData(ctx, telegramID, "admin_add_channel_id", fmt.Sprintf("%d", chat.ID))
		b.state.SetData(ctx, telegramID, "admin_add_channel_title", chat.Title)
		b.state.Set(ctx, telegramID, user.AdminStateAddChannelLink)
		return c.Send(fmt.Sprintf("✅ Kanal topildi: <b>%s</b>\n\nEndi foydalanuvchilar obuna bo'lishi uchun kanal havolasini (URL) kiriting (Masalan: https://t.me/mychannel):", chat.Title), tele.ModeHTML)

	case user.AdminStateAddChannelLink:
		chatIDStr, _ := b.state.GetData(ctx, telegramID, "admin_add_channel_id")
		title, _ := b.state.GetData(ctx, telegramID, "admin_add_channel_title")
		
		var chatID int64
		fmt.Sscanf(chatIDStr, "%d", &chatID)

		ch := &channel.Channel{
			ChatID: chatID,
			Title:  title,
			URL:    text,
		}

		err := b.channelRepo.Create(ctx, ch)
		if err != nil {
			return c.Send("Kanalni saqlashda xatolik yuz berdi: " + err.Error())
		}

		b.state.Set(ctx, telegramID, user.AdminStateMainMenu)
		c.Send("✅ Kanal muvaffaqiyatli qo'shildi!")
		return b.handleAdminChannels(c)
	}

	return nil
}

func (b *Bot) handleAdminEditTextCallback(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	data := c.Callback().Data
	
	// Ensure user is admin
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

	sectionID := data[len("admin_edit_text_"):]
	b.state.SetData(ctx, telegramID, "admin_edit_text_id", sectionID)
	b.state.Set(ctx, telegramID, user.AdminStateEditTextInput)
	
	c.Edit(c.Message().Text + "\n\nIltimos, ushbu bo'lim uchun yangi matnni kiriting (bu matn bot orqali barchaga ko'rinadi):")
	return c.Respond()
}

func (b *Bot) handleAdminViewApplications(c tele.Context) error {
	ctx := context.Background()
	
	apps, err := b.appRepo.GetAll(ctx, 20, 0)
	if err != nil || len(apps) == 0 {
		return c.Send("Hozircha arizalar yo'q.")
	}

	var rows [][]tele.InlineButton
	for _, app := range apps {
		v, _ := b.vacancyRepo.GetByID(ctx, app.VacancyID)
		title := "Noma'lum"
		if v != nil {
			title = v.Title
		}
		
		res, _ := b.resumeRepo.GetByID(ctx, app.ResumeID)
		candidate := "Noma'lum"
		if res != nil {
			candidate = res.FirstName
		}

		statusIcon := "⏳"
		if app.Status == application.StatusRejected {
			statusIcon = "❌"
		} else if app.Status == application.StatusAccepted {
			statusIcon = "✅"
		}

		btnText := fmt.Sprintf("💼 %s | %s (%s)", title, candidate, statusIcon)
		btn := tele.Btn{Text: btnText, Data: "admin_vw_app_" + app.ID.String()}
		rows = append(rows, []tele.InlineButton{*btn.Inline()})
	}
	
	markup := &tele.ReplyMarkup{}
	markup.InlineKeyboard = rows
	return c.Send("Barcha arizalar quyidagilar. Batafsil ma'lumot uchun ustiga bosing:", markup)
}

func (b *Bot) handleAdminViewAppCallback(c tele.Context) error {
	ctx := context.Background()
	data := c.Callback().Data
	appIDStr := data[13:]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Xatolik: Noto'g'ri ID", ShowAlert: true})
	}

	app, err := b.appRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ariza topilmadi", ShowAlert: true})
	}

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
	
	markup := &tele.ReplyMarkup{}
	var rows [][]tele.InlineButton
	if app.Status == application.StatusSubmitted {
		btnApprove := tele.Btn{Text: "✅ Tasdiqlash", Data: "admin_apprv_app_" + app.ID.String()}
		btnReject := tele.Btn{Text: "❌ Rad etish", Data: "admin_rejct_app_" + app.ID.String()}
		rows = append(rows, []tele.InlineButton{*btnApprove.Inline(), *btnReject.Inline()})
	}
	
	btnBack := tele.Btn{Text: "🔙 Orqaga", Data: "admin_bck_apps"}
	rows = append(rows, []tele.InlineButton{*btnBack.Inline()})
	markup.InlineKeyboard = rows

	c.Bot().Delete(c.Message())

	if res != nil && res.PhotoFileID != nil && *res.PhotoFileID != "" {
		photo := &tele.Photo{File: tele.File{FileID: *res.PhotoFileID}, Caption: text}
		c.Send(photo, markup)
	} else {
		c.Send(text, markup)
	}
	return c.Respond()
}

func (b *Bot) handleAdminBackToAppsCallback(c tele.Context) error {
	c.Bot().Delete(c.Message())
	b.handleAdminViewApplications(c)
	return c.Respond()
}

func (b *Bot) handleAdminApproveAppCallback(c tele.Context) error {
	return b.processAppStatusChange(c, "admin_apprv_app_", application.StatusAccepted, "✅ Tasdiqlandi", "🎉 Tabriklaymiz! Arizangiz ma'qullandi. Tez orada siz bilan bog'lanamiz.")
}

func (b *Bot) handleAdminRejectAppCallback(c tele.Context) error {
	return b.processAppStatusChange(c, "admin_rejct_app_", application.StatusRejected, "❌ Rad etildi", "")
}

func (b *Bot) processAppStatusChange(c tele.Context, prefix, newStatus, adminStatusText, userMsg string) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	data := c.Callback().Data
	
	// Ensure user is admin
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

	appIDStr := data[len(prefix):]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Xatolik: Noto'g'ri ID", ShowAlert: true})
	}

	app, err := b.appRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ariza topilmadi", ShowAlert: true})
	}

	if app.Status != application.StatusSubmitted {
		return c.Respond(&tele.CallbackResponse{Text: "Bu ariza holati avval o'zgartirilgan", ShowAlert: true})
	}

	err = b.appRepo.UpdateStatus(ctx, appID, newStatus)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Xatolik yuz berdi", ShowAlert: true})
	}

	// Update Admin message
	newMarkup := &tele.ReplyMarkup{}
	btnBack := tele.Btn{Text: "🔙 Orqaga", Data: "admin_bck_apps"}
	newMarkup.InlineKeyboard = [][]tele.InlineButton{{*btnBack.Inline()}}

	if c.Message().Caption != "" {
		newCaption := strings.Replace(c.Message().Caption, "Holati: submitted", "Holati: " + adminStatusText, 1)
		c.Edit(newCaption, newMarkup)
	} else {
		newText := strings.Replace(c.Message().Text, "Holati: submitted", "Holati: " + adminStatusText, 1)
		c.Edit(newText, newMarkup)
	}
	c.Respond(&tele.CallbackResponse{Text: "Holat o'zgartirildi!"})

	// Notify User
	if userMsg != "" {
		u, err := b.userRepo.GetByID(ctx, app.UserID)
		if err == nil && u != nil {
			b.Client.Send(&tele.User{ID: u.TelegramID}, userMsg)
		}
	}

	return nil
}

func (b *Bot) handleAdminChannels(c tele.Context) error {
	ctx := context.Background()
	channels, err := b.channelRepo.GetAll(ctx)
	if err != nil {
		return c.Send("Kanallarni olishda xatolik yuz berdi.")
	}

	if len(channels) == 0 {
		return c.Send("Hozircha majburiy kanallar qo'shilmagan.", b.adminChannelsMarkup())
	}

	text := "📋 Majburiy kanallar ro'yxati:\n\n"
	markup := &tele.ReplyMarkup{}
	var rows [][]tele.InlineButton
	for i, ch := range channels {
		text += fmt.Sprintf("%d. %s\n", i+1, ch.Title)
		btnDelete := tele.Btn{Text: fmt.Sprintf("🗑 %d-ni o'chirish", i+1), Data: "admin_del_chan_" + ch.ID.String()}
		rows = append(rows, []tele.InlineButton{*btnDelete.Inline()})
	}
	markup.InlineKeyboard = rows

	c.Send(text, markup)
	return c.Send("Kanal sozlamalari:", b.adminChannelsMarkup())
}

func (b *Bot) adminChannelsMarkup() *tele.ReplyMarkup {
	btnAdd := tele.ReplyButton{Text: "➕ Kanal qo'shish"}
	btnBack := tele.ReplyButton{Text: "🔙 Orqaga (Asosiy menyu)"}

	return &tele.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{
			{btnAdd},
			{btnBack},
		},
		ResizeKeyboard: true,
	}
}

func (b *Bot) handleAdminDeleteChannelCallback(c tele.Context) error {
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
		return c.Respond(&tele.CallbackResponse{Text: "Siz admin emassiz", ShowAlert: true})
	}

	data := c.Callback().Data
	chIDStr := data[len("admin_del_chan_"):]
	chID, err := uuid.Parse(chIDStr)
	if err == nil {
		b.channelRepo.DeleteByID(ctx, chID)
	}

	c.Respond(&tele.CallbackResponse{Text: "Kanal o'chirildi!"})
	c.Delete()
	return b.handleAdminChannels(c)
}
