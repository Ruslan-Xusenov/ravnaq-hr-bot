package resume

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Resume, error)
	GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (*Resume, error)
	Create(ctx context.Context, r *Resume) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Resume, error) {
	query := `
		SELECT id, user_id, version, first_name, last_name, photo_file_id, 
		       COALESCE(address_region, ''), COALESCE(address_city, ''), COALESCE(address_detail, ''), COALESCE(expected_salary, 0), COALESCE(salary_currency, ''),
		       COALESCE(education_text, ''), COALESCE(skills_text, ''), COALESCE(languages_text, ''), COALESCE(portfolio_url, ''), pdf_file_id,
		       extra_phone_1, extra_phone_2, consent_at, is_current, created_at, updated_at
		FROM resumes
		WHERE id = $1
	`
	var res Resume
	err := r.db.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.UserID, &res.Version, &res.FirstName, &res.LastName, &res.PhotoFileID,
		&res.AddressRegion, &res.AddressCity, &res.AddressDetail, &res.ExpectedSalary, &res.SalaryCurrency,
		&res.EducationText, &res.SkillsText, &res.LanguagesText, &res.PortfolioURL, &res.PDFFileID,
		&res.ExtraPhone1, &res.ExtraPhone2, &res.ConsentAt, &res.IsCurrent, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get resume by id: %w", err)
	}
	return &res, nil
}

func (r *repository) GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (*Resume, error) {
	query := `
		SELECT id, user_id, version, first_name, last_name, photo_file_id, 
		       COALESCE(address_region, ''), COALESCE(address_city, ''), COALESCE(address_detail, ''), COALESCE(expected_salary, 0), COALESCE(salary_currency, ''),
		       COALESCE(education_text, ''), COALESCE(skills_text, ''), COALESCE(languages_text, ''), COALESCE(portfolio_url, ''), pdf_file_id,
		       extra_phone_1, extra_phone_2, consent_at, is_current, created_at, updated_at
		FROM resumes
		WHERE user_id = $1 AND is_current = true
		LIMIT 1
	`
	var res Resume
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&res.ID, &res.UserID, &res.Version, &res.FirstName, &res.LastName, &res.PhotoFileID,
		&res.AddressRegion, &res.AddressCity, &res.AddressDetail, &res.ExpectedSalary, &res.SalaryCurrency,
		&res.EducationText, &res.SkillsText, &res.LanguagesText, &res.PortfolioURL, &res.PDFFileID,
		&res.ExtraPhone1, &res.ExtraPhone2, &res.ConsentAt, &res.IsCurrent, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get current resume: %w", err)
	}
	return &res, nil
}

func (r *repository) Create(ctx context.Context, res *Resume) error {
	// First, set previous current to false
	updateQuery := `UPDATE resumes SET is_current = false WHERE user_id = $1`
	_, _ = r.db.Exec(ctx, updateQuery, res.UserID)

	// Get latest version
	var maxVersion int
	r.db.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM resumes WHERE user_id = $1`, res.UserID).Scan(&maxVersion)
	res.Version = maxVersion + 1

	query := `
		INSERT INTO resumes (
			user_id, version, first_name, last_name, photo_file_id, address_region, expected_salary, salary_currency, skills_text, extra_phone_1, extra_phone_2, is_current, consent_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW()
		) RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		res.UserID, res.Version, res.FirstName, res.LastName, res.PhotoFileID, res.AddressRegion, res.ExpectedSalary, res.SalaryCurrency, res.SkillsText, res.ExtraPhone1, res.ExtraPhone2, true,
	).Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create resume: %w", err)
	}
	return nil
}
