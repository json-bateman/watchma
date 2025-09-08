package services

import (
	"encoding/json"
	"sync"

	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/nats-io/nats.go"
)

type MessagingService struct {
	nats         *nats.Conn
	gameClients  map[string]map[chan string]bool // roomName -> client
	roomMessages map[string][]string             // room -> messages
	mu           *sync.RWMutex
}

func NewMessageService(n *nats.Conn, gC map[string]map[chan string]bool, rM map[string][]string) *MessagingService {
	return &MessagingService{
		nats:         n,
		gameClients:  gC,
		roomMessages: rM,
	}
}

func (ms *MessagingService) SetupSubscriptions() {
	ms.nats.Subscribe("chat.*", ms.handleChatMessage)
	// ms.nats.Subscribe(utils.JOIN_MSG+".*", ms.handleJoinMessage)
	// ms.nats.Subscribe(utils.LEAVE_MSG+".*", ms.handleLeaveMessage)
}

func (ms *MessagingService) handleChatMessage(m *nats.Msg) {
	var req types.NatsPublishRequest
	json.Unmarshal(m.Data, &req)

	ms.mu.Lock()
	// Store message in room history
	ms.roomMessages[req.Room] = append(ms.roomMessages[req.Room], req.Message)
	gameClients := ms.gameClients[req.Room]
	ms.mu.Unlock()

	for gameClient := range gameClients {
		select {
		case gameClient <- req.Message:
		default: // Non-blocking send to prevent deadlock
		}
	}
}

// func (ms *MessagingService) handleJoinMessage(m *nats.Msg) {
// 	room := strings.TrimPrefix(m.Subject, utils.JOIN_MSG+".")
// 	message := string(m.Data)
//
// 	ms.mu.Lock()
// 	gameClients := ms.gameClients[room]
// 	ms.mu.Unlock()
//
// 	for gameClient := range gameClients {
// 		select {
// 		case gameClient <- utils.JOIN_MSG:
// 		default: // Non-blocking send to prevent deadlock
// 		}
// 	}
// }
//
// func (ms *MessagingService) handleLeaveMessage(m *nats.Msg) {
// 	room := strings.TrimPrefix(m.Subject, utils.LEAVE_MSG+".")
//
// 	ms.mu.Lock()
// 	gameClients := ms.gameClients[room]
// 	ms.mu.Unlock()
//
// 	for gameClient := range gameClients {
// 		select {
// 		case gameClient <- utils.LEAVE_MSG:
// 		default: // Non-blocking send to prevent deadlock
// 		}
// 	}
// }
