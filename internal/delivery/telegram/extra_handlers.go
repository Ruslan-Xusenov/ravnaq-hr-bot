package telegram

import (
	"context"
	"fmt"

	"github.com/company/hrbot/internal/domain/user"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) handleProfileMenu(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID

	u, err := b.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || u == nil {
		return c.Send("Texnik xatolik")
	}

	phone := "Kiritilmagan"
	if u.PrimaryPhone != nil {
		phone = *u.PrimaryPhone
	}

	lang := "O'zbek"
	if u.LanguageCode != nil {
		if *u.LanguageCode == "ru" {
			lang = "Русский"
		} else if *u.LanguageCode == "en" {
			lang = "English"
		}
	}

	text := fmt.Sprintf("👤 <b>Mening profilim</b>\n\n🆔 ID: %d\n📝 Ism: %s\n📞 Telefon: %s\n🌐 Til: %s",
		u.TelegramID, u.TelegramFirstName, phone, lang)

	return c.Send(text, tele.ModeHTML)
}

func (b *Bot) handleAbout(c tele.Context) error {
	text := `🏢 <b>Biz haqimizda</b>

Bu maxsus HR Bot bo'lib, sizga kompaniyamizdagi mavjud bo'sh ish o'rinlarini ko'rish, ularga ariza topshirish va o'z rezyumeyingizni qulay tarzda boshqarish imkonini beradi.

Biz bilan birga o'z karyerangizni quring! 🚀`
	return c.Send(text, tele.ModeHTML)
}

func (b *Bot) handleFAQ(c tele.Context) error {
	text := `❓ <b>Ko'p beriladigan savollar (FAQ)</b>

<b>1. Rezyumeni qanday yarataman?</b>
- Asosiy menyudan "📄 Rezyume" tugmasini bosing va bot so'ragan ma'lumotlarni kiriting.

<b>2. Bo'sh ish o'rinlariga qanday ariza topshiraman?</b>
- "💼 Bo'sh ish o'rinlari" tugmasini bosib, ro'yxatdan o'zingizga yoqqan ishni tanlang va "Ariza topshirish" tugmasini bosing (buning uchun avval rezyume yaratgan bo'lishingiz kerak).

<b>3. Tilni qanday o'zgartiraman?</b>
- "⚙️ Sozlamalar" bo'limiga kiring va o'zingizga qulay tilni tanlang.`
	return c.Send(text, tele.ModeHTML)
}

func (b *Bot) handleContactUs(c tele.Context) error {
	text := `📞 <b>Aloqa</b>

Agar sizda savollar, takliflar yoki texnik muammolar bo'lsa, biz bilan quyidagi manzillar orqali bog'lanishingiz mumkin:

📍 Manzil: Toshkent shahar
📧 Email: hr@example.com
☎️ Telefon: +998 90 123 45 67
💬 Telegram: @hr_support`
	return c.Send(text, tele.ModeHTML)
}

func (b *Bot) handleSettings(c tele.Context) error {
	ctx := context.Background()
	telegramID := c.Sender().ID
	
	// Set state to selecting language so language callback works
	b.state.Set(ctx, telegramID, user.StateSelectingLanguage)

	btnUz := tele.Btn{Text: "🇺🇿 O'zbekcha", Data: "lang_uz"}
	btnRu := tele.Btn{Text: "🇷🇺 Русский", Data: "lang_ru"}
	btnEn := tele.Btn{Text: "🇬🇧 English", Data: "lang_en"}

	markup := &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*btnUz.Inline(), *btnRu.Inline(), *btnEn.Inline()},
		},
	}

	return c.Send("⚙️ <b>Sozlamalar</b>\n\nIltimos, o'zingizga qulay tilni tanlang:", markup, tele.ModeHTML)
}
