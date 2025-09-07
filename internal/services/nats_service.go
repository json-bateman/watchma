package services

import (
	"strings"
	"sync"

	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/nats-io/nats.go"
)

type MessagingService struct {
	nats         *nats.Conn
	gameClients  map[string]map[chan string]bool
	roomMessages map[string][]string
	mu           *sync.RWMutex
}

func (ms *MessagingService) SetupSubscriptions() {
	ms.nats.Subscribe("chat.*", ms.handleChatMessage)
	ms.nats.Subscribe(utils.LEAVE_MSG+".*", ms.handleJoinMessage)
	ms.nats.Subscribe("join.*", ms.handleLeaveMessage)
}

func (ms *MessagingService) handleChatMessage(m *nats.Msg) {
	room := strings.TrimPrefix(m.Subject, "chat.")
	message := string(m.Data)

	ms.mu.Lock()
	// Store message in room history
	ms.roomMessages[room] = append(ms.roomMessages[room], message)
	gameClients := ms.gameClients[room]
	ms.mu.Unlock()

	for gameClient := range gameClients {
		select {
		case gameClient <- message:
		default: // Non-blocking send to prevent deadlock
		}
	}
}

func (ms *MessagingService) handleJoinMessage(m *nats.Msg) {
	room := strings.TrimPrefix(m.Subject, utils.JOIN_MSG+".")

	ms.mu.Lock()
	gameClients := ms.gameClients[room]
	ms.mu.Unlock()

	for gameClient := range gameClients {
		select {
		case gameClient <- utils.JOIN_MSG:
		default: // Non-blocking send to prevent deadlock
		}
	}
}

func (ms *MessagingService) handleLeaveMessage(m *nats.Msg) {
	room := strings.TrimPrefix(m.Subject, utils.LEAVE_MSG+".")

	ms.mu.Lock()
	gameClients := ms.gameClients[room]
	ms.mu.Unlock()

	for gameClient := range gameClients {
		select {
		case gameClient <- utils.LEAVE_MSG:
		default: // Non-blocking send to prevent deadlock
		}
	}
}
