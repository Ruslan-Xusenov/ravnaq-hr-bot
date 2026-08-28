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
	StateResumeFirstName  = "resume_first_name"
	StateResumeLastName   = "resume_last_name"
	StateResumeRegion     = "resume_region"
	StateResumeSalary     = "resume_salary"
	StateResumeConfirm    = "resume_confirm"
)
