package utils

const (
	MESSAGE_SENT_EVENT     = "Message Sent Event"
	ROOM_UPDATE_EVENT      = "Room Update Event"
	ROOM_START_EVENT       = "Room Start Event"
	ROOM_VOTING_EVENT      = "Room Voting Event"
	ROOM_FINISH_EVENT      = "Room Finish Event"
	ROOM_LIST_UPDATE_EVENT = "Room List Update Event"
	SESSION_COOKIE_NAME    = "watchma_session"
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
