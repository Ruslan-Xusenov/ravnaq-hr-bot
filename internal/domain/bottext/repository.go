package bottext

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Get(ctx context.Context, id string) (*BotText, error)
	Set(ctx context.Context, id, textContent string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context, id string) (*BotText, error) {
	query := `SELECT id, text_content FROM bot_texts WHERE id = $1`
	var t BotText
	err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.TextContent)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) Set(ctx context.Context, id, textContent string) error {
	query := `
		INSERT INTO bot_texts (id, text_content)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET text_content = EXCLUDED.text_content
	`
	_, err := r.db.Exec(ctx, query, id, textContent)
	return err
}
