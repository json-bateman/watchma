package chat

import (
	"time"
)

type ChatMessage struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
