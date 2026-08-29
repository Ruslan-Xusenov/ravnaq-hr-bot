package telegram

import (
	"log/slog"
	"strings"
	"time"

	"github.com/company/hrbot/internal/config"
	"github.com/company/hrbot/internal/domain/application"
	"github.com/company/hrbot/internal/domain/bottext"
	"github.com/company/hrbot/internal/domain/resume"
	"github.com/company/hrbot/internal/domain/user"
	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/company/hrbot/internal/domain/channel"
	"github.com/company/hrbot/pkg/i18n"
	tele "gopkg.in/telebot.v3"
)

type Bot struct {
	Client      *tele.Bot
	userRepo    user.Repository
	resumeRepo  resume.Repository
	vacancyRepo vacancy.Repository
	appRepo     application.Repository
	textRepo    bottext.Repository
	channelRepo channel.Repository
	state       *StateManager
	i18n        *i18n.Translator
	adminIDs    []int64
}

func NewBot(
	cfg *config.Config,
	ur user.Repository,
	rr resume.Repository,
	vr vacancy.Repository,
	ar application.Repository,
	tr bottext.Repository,
	cr channel.Repository,
	state *StateManager,
	trans *i18n.Translator,
) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Client:      b,
		userRepo:    ur,
		resumeRepo:  rr,
		vacancyRepo: vr,
		appRepo:     ar,
		textRepo:    tr,
		channelRepo: cr,
		state:       state,
		i18n:        trans,
		adminIDs:    cfg.TelegramAdminIDs,
	}, nil
}

func (b *Bot) Start() {
	slog.Info("Starting Telegram Bot (Polling mode)")
	b.Client.Start()
}

func (b *Bot) Stop() {
	slog.Info("Stopping Telegram Bot")
	b.Client.Stop()
}

func (b *Bot) SetupHandlers() {
	b.Client.Use(b.subscriptionMiddleware())

	b.Client.Handle("/start", b.handleStart)
	
	b.Client.Handle(tele.OnCallback, func(c tele.Context) error {
		// Route callbacks
		data := c.Callback().Data
		if data == "check_sub" {
			return b.handleCheckSubscriptionCallback(c)
		}
		if strings.HasPrefix(data, "admin_del_chan_") {
			return b.handleAdminDeleteChannelCallback(c)
		}
		if data == "lang_uz" || data == "lang_ru" || data == "lang_en" {
			return b.handleLanguageCallback(c)
		}
		if data == "resume_confirm" || data == "resume_cancel" || data == "resume_edit" {
			return b.handleResumeCallback(c)
		}
		if len(data) > 6 && data[:6] == "apply_" {
			return b.handleApplyCallback(c)
		}
		if len(data) > 14 && data[:14] == "confirm_apply_" {
			return b.handleConfirmApplyCallback(c)
		}
		if len(data) > 17 && data[:17] == "new_resume_apply_" {
			return b.handleNewResumeApplyCallback(c)
		}
		if len(data) > 9 && data[:9] == "view_vac_" {
			return b.handleViewVacancyCallback(c)
		}
		if data == "back_to_vacancies" {
			return b.handleBackToVacanciesCallback(c)
		}
		if len(data) > 16 && data[:16] == "admin_edit_text_" {
			return b.handleAdminEditTextCallback(c)
		}
		if len(data) > 16 && data[:16] == "admin_apprv_app_" {
			return b.handleAdminApproveAppCallback(c)
		}
		if len(data) > 16 && data[:16] == "admin_rejct_app_" {
			return b.handleAdminRejectAppCallback(c)
		}
		return nil
	})
	b.Client.Handle(tele.OnContact, b.handleContact)
	b.Client.Handle(tele.OnText, b.handleText)
	b.Client.Handle(tele.OnPhoto, b.handlePhoto)
	
	// Setup more handlers here
}
