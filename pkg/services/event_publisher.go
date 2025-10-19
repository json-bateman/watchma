package services

import (
	"log/slog"

	"watchma/pkg/utils"

	"github.com/nats-io/nats.go"
)

type EventPublisher struct {
	nc     *nats.Conn
	logger *slog.Logger
}

func NewEventPublisher(nc *nats.Conn, logger *slog.Logger) *EventPublisher {
	return &EventPublisher{
		nc:     nc,
		logger: logger,
	}
}

func (ep *EventPublisher) Publish(subject string, data []byte) error {
	if err := ep.nc.Publish(subject, data); err != nil {
		ep.logger.Error("Failed to publish event", "subject", subject, "error", err)
		return err
	}
	ep.logger.Debug(utils.NATS_PUB, "subject", subject, "msg", string(data))
	return nil
}

// Helper methods for common patterns
func (ep *EventPublisher) PublishRoomEvent(roomName, event string) error {
	subject := utils.RoomSubject(roomName)
	ep.logger.Debug(utils.NATS_PUB, "subject", subject, "msg", event)
	return ep.Publish(subject, []byte(event))
}

func (ep *EventPublisher) PublishLobbyEvent(event string) error {
	ep.logger.Debug(utils.NATS_PUB, "msg", event)
	return ep.Publish(utils.NATS_LOBBY_ROOMS, []byte(event))
}
