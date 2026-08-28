package application

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusSubmitted = "submitted"
	StatusInReview  = "in_review"
	StatusInterview = "interview"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusWithdrawn = "withdrawn"
	StatusCancelled = "cancelled"
)

type Application struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	VacancyID       uuid.UUID
	ResumeID        uuid.UUID
	Status          string
	AdminNote       *string
	RejectionReason *string
	AssignedAdminID *uuid.UUID
	InterviewAt     *time.Time
	AcceptedAt      *time.Time
	LinkSentAt      *time.Time
	SubmittedAt     time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
