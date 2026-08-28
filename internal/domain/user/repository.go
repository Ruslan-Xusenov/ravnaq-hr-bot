package user

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, 
		       language_code, primary_phone, status, last_activity_at, created_at, updated_at
		FROM users
		WHERE telegram_id = $1 AND deleted_at IS NULL
	`
	var u User
	err := r.db.QueryRow(ctx, query, telegramID).Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.TelegramFirstName, &u.TelegramLastName,
		&u.LanguageCode, &u.PrimaryPhone, &u.Status, &u.LastActivityAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}
	return &u, nil
}

func (r *repository) GetAll(ctx context.Context) ([]User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, 
		       language_code, primary_phone, status, last_activity_at, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.TelegramID, &u.TelegramUsername, &u.TelegramFirstName, &u.TelegramLastName,
			&u.LanguageCode, &u.PrimaryPhone, &u.Status, &u.LastActivityAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *repository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (telegram_id, telegram_username, telegram_first_name, telegram_last_name, 
		                   language_code, primary_phone, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, last_activity_at, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		u.TelegramID, u.TelegramUsername, u.TelegramFirstName, u.TelegramLastName,
		u.LanguageCode, u.PrimaryPhone, u.Status,
	).Scan(&u.ID, &u.LastActivityAt, &u.CreatedAt, &u.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users 
		SET telegram_username = $2, telegram_first_name = $3, telegram_last_name = $4,
		    language_code = $5, primary_phone = $6, status = $7, 
		    last_activity_at = $8, updated_at = $9
		WHERE id = $1
	`
	u.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		u.ID, u.TelegramUsername, u.TelegramFirstName, u.TelegramLastName,
		u.LanguageCode, u.PrimaryPhone, u.Status,
		u.LastActivityAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
