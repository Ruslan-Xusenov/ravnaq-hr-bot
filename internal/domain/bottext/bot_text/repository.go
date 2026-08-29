package bot_text

import "context"

type Repository interface {
	Get(ctx context.Context, id string) (string, error)
	Set(ctx context.Context, id, textContent string) error
}
