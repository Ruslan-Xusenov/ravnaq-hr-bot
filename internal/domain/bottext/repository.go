package bottext

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Get(ctx context.Context, id string) (string, error)
	Set(ctx context.Context, id, textContent string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context, id string) (string, error) {
	var textContent string
	err := r.db.QueryRow(ctx, "SELECT text_content FROM bot_texts WHERE id = $1", id).Scan(&textContent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return "", nil // Return empty string if not set yet
		}
		return "", err
	}
	return textContent, nil
}

func (r *repository) Set(ctx context.Context, id, textContent string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO bot_texts (id, text_content)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET text_content = EXCLUDED.text_content
	`, id, textContent)
	return err
}
