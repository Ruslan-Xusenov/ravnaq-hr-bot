package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	HasApplied(ctx context.Context, userID, vacancyID uuid.UUID) (bool, error)
	Create(ctx context.Context, app *Application) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Application, error)
	GetAll(ctx context.Context, limit, offset int) ([]Application, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) HasApplied(ctx context.Context, userID, vacancyID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM applications WHERE user_id = $1 AND vacancy_id = $2 AND status != $3 LIMIT 1`
	var exists int
	err := r.db.QueryRow(ctx, query, userID, vacancyID, StatusWithdrawn).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *repository) Create(ctx context.Context, app *Application) error {
	query := `
		INSERT INTO applications (user_id, vacancy_id, resume_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, submitted_at
	`
	err := r.db.QueryRow(ctx, query, app.UserID, app.VacancyID, app.ResumeID, StatusSubmitted).
		Scan(&app.ID, &app.SubmittedAt)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	return nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Application, error) {
	query := `
		SELECT id, user_id, vacancy_id, resume_id, status, submitted_at
		FROM applications
		WHERE user_id = $1
		ORDER BY submitted_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var apps []Application
	for rows.Next() {
		var a Application
		if err := rows.Scan(&a.ID, &a.UserID, &a.VacancyID, &a.ResumeID, &a.Status, &a.SubmittedAt); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func (r *repository) GetAll(ctx context.Context, limit, offset int) ([]Application, error) {
	query := `
		SELECT id, user_id, vacancy_id, resume_id, status, submitted_at
		FROM applications
		ORDER BY submitted_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query all applications: %w", err)
	}
	defer rows.Close()

	var apps []Application
	for rows.Next() {
		var a Application
		if err := rows.Scan(&a.ID, &a.UserID, &a.VacancyID, &a.ResumeID, &a.Status, &a.SubmittedAt); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE applications SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}
