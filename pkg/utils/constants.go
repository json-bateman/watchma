package utils

const (
	ROOM_UPDATE_EVENT      = "Room Update Event"
	MESSAGE_SENT_EVENT     = "Message Sent Event"
	ROOM_START_EVENT       = "Room Start Event"
	ROOM_FINISH_EVENT      = "Room Finish Event"
	ROOM_LIST_UPDATE_EVENT = "Room List Update Event"
	USERNAME_COOKIE        = "Watchma_Username"
)

// NATS
const (
	NATS_LOBBY_ROOMS = "app.lobby.rooms"
	NATS_PUB         = "NATS: Published"
	NATS_SUB         = "NATS: Subscribed"
)

// RoomSubject returns the NATS subject for a specific room
func RoomSubject(roomName string) string {
	return "app.room." + roomName
}
