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
	ID             uuid.UUID
	Title          string
	Slug           string
	Department     *string
	Location       *string
	EmploymentType *string
	Schedule       *string
	SalaryFrom     *float64
	SalaryTo       *float64
	SalaryCurrency *string
	Description    *string
	Requirements   *string
	Benefits       *string
	ExternalURL    *string
	Status         string
	PublishedAt    *time.Time
	ExpiresAt      *time.Time
	SortOrder      int
	CreatedBy      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
