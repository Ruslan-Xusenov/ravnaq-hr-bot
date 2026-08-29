package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, ch *Channel) error
	GetAll(ctx context.Context) ([]Channel, error)
	DeleteByChatID(ctx context.Context, chatID int64) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, ch *Channel) error {
	query := `
		INSERT INTO mandatory_channels (chat_id, title, url)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, ch.ChatID, ch.Title, ch.URL).Scan(&ch.ID, &ch.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	return nil
}

func (r *repository) GetAll(ctx context.Context) ([]Channel, error) {
	query := `SELECT id, chat_id, title, url, created_at FROM mandatory_channels ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(&c.ID, &c.ChatID, &c.Title, &c.URL, &c.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, nil
}

func (r *repository) DeleteByChatID(ctx context.Context, chatID int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM mandatory_channels WHERE chat_id = $1", chatID)
	return err
}

func (r *repository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM mandatory_channels WHERE id = $1", id)
	return err
}
