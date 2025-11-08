package room

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

const (
	MessageSentEvent    = "Message Sent Event"
	RoomUpdateEvent     = "Room Update Event"
	RoomStartEvent      = "Room Start Event"
	RoomVotingEvent     = "Room Voting Event"
	RoomAnnounceEvent   = "Room Announce Event"
	RoomFinishEvent     = "Room Finish Event"
	RoomListUpdateEvent = "Room List Update Event"
)

// NATS
const (
	NATSLobbyRooms = "app.lobby.rooms"
	NATSPub        = "NATS: Published"
	NATSSub        = "NATS: Subscribed"
)

// RoomSubject returns the NATS subject for a specific room
func RoomSubject(roomName string) string {
	return "app.room." + roomName
}

// EventPublisher handles publishing events to NATS
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
	ep.logger.Debug(NATSPub, "subject", subject, "msg", string(data))
	return nil
}

func (ep *EventPublisher) PublishRoomEvent(roomName, event string) error {
	subject := RoomSubject(roomName)
	ep.logger.Debug(NATSPub, "subject", subject, "msg", event)
	return ep.Publish(subject, []byte(event))
}

func (ep *EventPublisher) PublishLobbyEvent(event string) error {
	ep.logger.Debug(NATSPub, "msg", event)
	return ep.Publish(NATSLobbyRooms, []byte(event))
}
