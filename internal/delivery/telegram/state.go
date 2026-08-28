package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StateManager struct {
	redis *redis.Client
}

func NewStateManager(r *redis.Client) *StateManager {
	return &StateManager{redis: r}
}

func (s *StateManager) Get(ctx context.Context, telegramID int64) (string, error) {
	key := fmt.Sprintf("user_state:%d", telegramID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *StateManager) Set(ctx context.Context, telegramID int64, state string) error {
	key := fmt.Sprintf("user_state:%d", telegramID)
	// State expires in 1 hour if no interaction
	return s.redis.Set(ctx, key, state, time.Hour).Err()
}

func (s *StateManager) Clear(ctx context.Context, telegramID int64) error {
	key := fmt.Sprintf("user_state:%d", telegramID)
	return s.redis.Del(ctx, key).Err()
}

// Store temporary data for the current flow (like creating resume)
func (s *StateManager) SetData(ctx context.Context, telegramID int64, field, value string) error {
	key := fmt.Sprintf("user_data:%d", telegramID)
	return s.redis.HSet(ctx, key, field, value).Err()
}

func (s *StateManager) GetData(ctx context.Context, telegramID int64, field string) (string, error) {
	key := fmt.Sprintf("user_data:%d", telegramID)
	val, err := s.redis.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (s *StateManager) ClearData(ctx context.Context, telegramID int64) error {
	key := fmt.Sprintf("user_data:%d", telegramID)
	return s.redis.Del(ctx, key).Err()
}
