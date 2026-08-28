package resume

import (
	"time"

	"github.com/google/uuid"
)

type Resume struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Version        int
	FirstName      string
	LastName       string
	PhotoFileID    *string
	AddressRegion  string
	AddressCity    string
	AddressDetail  string
	ExpectedSalary float64
	SalaryCurrency string
	EducationText  string
	SkillsText     string
	LanguagesText  string
	PortfolioURL   string
	ExtraPhone1    *string
	ExtraPhone2    *string
	PDFFileID      *uuid.UUID
	ConsentAt      time.Time
	IsCurrent      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type WorkExperience struct {
	ID           uuid.UUID
	ResumeID     uuid.UUID
	CompanyName  string
	PositionName string
	StartedAt    time.Time
	EndedAt      *time.Time
	IsCurrent    bool
	Duties       string
	LeavingReason *string
	City         *string
	SortOrder    int
}
