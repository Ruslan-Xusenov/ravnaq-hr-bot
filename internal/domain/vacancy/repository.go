package vacancy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetActive(ctx context.Context, limit, offset int) ([]Vacancy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Vacancy, error)
	GetAll(ctx context.Context, limit, offset int) ([]Vacancy, error)
	Create(ctx context.Context, v *Vacancy) error
	Update(ctx context.Context, v *Vacancy) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetActive(ctx context.Context, limit, offset int) ([]Vacancy, error) {
	query := `
		SELECT id, title, slug, department, location, employment_type, schedule, 
		       salary_from, salary_to, salary_currency, salary_text, description, requirements, benefits,
		       status, published_at
		FROM vacancies
		WHERE status = $1
		ORDER BY published_at DESC NULLS LAST, created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, StatusPublished, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query active vacancies: %w", err)
	}
	defer rows.Close()

	var vacancies []Vacancy
	for rows.Next() {
		var v Vacancy
		err := rows.Scan(
			&v.ID, &v.Title, &v.Slug, &v.Department, &v.Location, &v.EmploymentType, &v.Schedule,
			&v.SalaryFrom, &v.SalaryTo, &v.SalaryCurrency, &v.SalaryText, &v.Description, &v.Requirements, &v.Benefits,
			&v.Status, &v.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		vacancies = append(vacancies, v)
	}

	return vacancies, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Vacancy, error) {
	query := `
		SELECT id, title, slug, department, location, employment_type, schedule, 
		       salary_from, salary_to, salary_currency, salary_text, description, requirements, benefits,
		       status, published_at
		FROM vacancies
		WHERE id = $1
	`
	var v Vacancy
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.Title, &v.Slug, &v.Department, &v.Location, &v.EmploymentType, &v.Schedule,
		&v.SalaryFrom, &v.SalaryTo, &v.SalaryCurrency, &v.SalaryText, &v.Description, &v.Requirements, &v.Benefits,
		&v.Status, &v.PublishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get vacancy: %w", err)
	}
	return &v, nil
}

func (r *repository) GetAll(ctx context.Context, limit, offset int) ([]Vacancy, error) {
	query := `
		SELECT id, title, slug, department, location, employment_type, schedule, 
		       salary_from, salary_to, salary_currency, salary_text, description, requirements, benefits,
		       status, published_at
		FROM vacancies
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all vacancies: %w", err)
	}
	defer rows.Close()

	var vacancies []Vacancy
	for rows.Next() {
		var v Vacancy
		err := rows.Scan(
			&v.ID, &v.Title, &v.Slug, &v.Department, &v.Location, &v.EmploymentType, &v.Schedule,
			&v.SalaryFrom, &v.SalaryTo, &v.SalaryCurrency, &v.SalaryText, &v.Description, &v.Requirements, &v.Benefits,
			&v.Status, &v.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		vacancies = append(vacancies, v)
	}
	return vacancies, nil
}

func (r *repository) Create(ctx context.Context, v *Vacancy) error {
	query := `
		INSERT INTO vacancies (title, slug, department, location, employment_type, schedule, 
		       salary_from, salary_to, salary_currency, salary_text, description, requirements, benefits, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, 
		v.Title, v.Slug, v.Department, v.Location, v.EmploymentType, v.Schedule,
		v.SalaryFrom, v.SalaryTo, v.SalaryCurrency, v.SalaryText, v.Description, v.Requirements, v.Benefits, v.Status,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create vacancy: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, v *Vacancy) error {
	query := `
		UPDATE vacancies
		SET title = $1, department = $2, location = $3, employment_type = $4, schedule = $5,
		    salary_from = $6, salary_to = $7, salary_currency = $8, salary_text = $9, description = $10,
		    requirements = $11, benefits = $12, status = $13, updated_at = NOW()
		WHERE id = $14
	`
	_, err := r.db.Exec(ctx, query,
		v.Title, v.Department, v.Location, v.EmploymentType, v.Schedule,
		v.SalaryFrom, v.SalaryTo, v.SalaryCurrency, v.SalaryText, v.Description,
		v.Requirements, v.Benefits, v.Status, v.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update vacancy: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM vacancies WHERE id = $1", id)
	return err
}
