package channel

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ID        uuid.UUID `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
