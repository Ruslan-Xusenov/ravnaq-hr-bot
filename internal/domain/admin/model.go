package admin

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
