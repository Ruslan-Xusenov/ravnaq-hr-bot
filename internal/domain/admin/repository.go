package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*Admin, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*Admin, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM admins WHERE email = $1`
	var a Admin
	err := r.db.QueryRow(ctx, query, email).Scan(
		&a.ID, &a.Email, &a.Password, &a.Role, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get admin by email: %w", err)
	}
	return &a, nil
}
