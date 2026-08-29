package vacancy

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusPaused    = "paused"
	StatusExpired   = "expired"
	StatusArchived  = "archived"
)

type Vacancy struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Department     *string    `json:"department,omitempty"`
	Location       *string    `json:"location,omitempty"`
	EmploymentType *string    `json:"employment_type,omitempty"`
	Schedule       *string    `json:"schedule,omitempty"`
	SalaryFrom     *float64   `json:"salary_from,omitempty"`
	SalaryTo       *float64   `json:"salary_to,omitempty"`
	SalaryCurrency *string    `json:"salary_currency,omitempty"`
	SalaryText     *string    `json:"salary_text,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Requirements   *string    `json:"requirements,omitempty"`
	Benefits       *string    `json:"benefits,omitempty"`
	ExternalURL    *string    `json:"external_url,omitempty"`
	Status         string     `json:"status"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	SortOrder      int        `json:"sort_order"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
