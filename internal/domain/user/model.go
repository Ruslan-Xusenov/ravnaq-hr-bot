package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID
	TelegramID        int64
	TelegramUsername  *string
	TelegramFirstName *string
	TelegramLastName  *string
	LanguageCode      *string
	PrimaryPhone      *string
	Status            string
	LastActivityAt    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// Dialog States
const (
	StateNewUser          = "new_user"
	StateSelectingLanguage = "selecting_language"
	StateWaitingContact   = "waiting_contact"
	StateMainMenu         = "main_menu"

	// Resume States
	StateResumeFullName   = "resume_fullname"
	StateResumePhoto      = "resume_photo"
	StateResumeExperience = "resume_experience"
	StateResumeSalary     = "resume_salary"
	StateResumeAddress    = "resume_address"
	StateResumePhones     = "resume_phones"
	StateResumeConfirm    = "resume_confirm"

	// Admin States
	AdminStateMainMenu             = "admin_main_menu"
	AdminStateAddVacancyTitle      = "admin_add_vacancy_title"
	AdminStateAddVacancyLocation   = "admin_add_vacancy_location"
	AdminStateAddVacancySalary     = "admin_add_vacancy_salary"
	AdminStateAddVacancyDesc       = "admin_add_vacancy_desc"
	AdminStateBroadcastMessage     = "admin_broadcast_message"
)
