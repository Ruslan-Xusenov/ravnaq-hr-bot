package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeBroadcastMessage = "broadcast:message"
)

type BroadcastMessagePayload struct {
	Text     string `json:"text"`
	LangCode string `json:"lang_code"`
}

func NewBroadcastMessageTask(text string, langCode string) (*asynq.Task, error) {
	payload, err := json.Marshal(BroadcastMessagePayload{
		Text:     text,
		LangCode: langCode,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeBroadcastMessage, payload), nil
}
