package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoomService(t *testing.T) {
	rs := NewRoomService()
	assert.NotNil(t, rs)
	assert.NotNil(t, rs.Rooms)
	assert.Equal(t, 0, len(rs.Rooms))
}

func TestRoomService_AddRoom(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "testuser",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}

	rs.AddRoom("testroom", game)

	assert.True(t, rs.RoomExists("testroom"))
	room, exists := rs.GetRoom("testroom")
	require.True(t, exists)
	assert.Equal(t, "testroom", room.Name)
	assert.Equal(t, game, room.Game)
	assert.NotNil(t, room.Users)
	assert.NotNil(t, room.RoomMessages)
}

func TestRoomService_ConcurrentAddRoom(t *testing.T) {
	rs := NewRoomService()
	const numGoroutines = 1000
	var wg sync.WaitGroup

	// Add rooms concurrently
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			game := &types.GameSession{
				Host:        "testuser",
				MovieNumber: 5,
				MaxPlayers:  4,
				Votes:       make(map[*types.JellyfinItem]int),
			}
			rs.AddRoom(fmt.Sprintf("room%d", id), game)
		}(i)
	}

	wg.Wait()

	// Verify all rooms were added
	for i := range numGoroutines {
		roomName := fmt.Sprintf("room%d", i)
		assert.True(t, rs.RoomExists(roomName), "Room %s should exist", roomName)
	}
}

func TestRoomService_DeleteRoom(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "testuser",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}

	rs.AddRoom("testroom", game)
	assert.True(t, rs.RoomExists("testroom"))

	rs.DeleteRoom("testroom")
	assert.False(t, rs.RoomExists("testroom"))

	_, exists := rs.GetRoom("testroom")
	assert.False(t, exists)
}

func TestRoomService_ConcurrentDeleteRoom(t *testing.T) {
	rs := NewRoomService()
	const numRooms = 1000

	// Add rooms first
	for i := range numRooms {
		game := &types.GameSession{
			Host:        "testuser",
			MovieNumber: 5,
			MaxPlayers:  4,
			Votes:       make(map[*types.JellyfinItem]int),
		}
		rs.AddRoom(fmt.Sprintf("room%d", i), game)
	}

	var wg sync.WaitGroup

	// Delete rooms concurrently
	for i := range numRooms {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rs.DeleteRoom(fmt.Sprintf("room%d", id))
		}(i)
	}

	wg.Wait()

	// Verify all rooms were deleted
	for i := range numRooms {
		roomName := fmt.Sprintf("room%d", i)
		assert.False(t, rs.RoomExists(roomName), "Room %s should not exist", roomName)
	}
}

func TestRoom_AddUser(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "host",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}
	rs.AddRoom("testroom", game)

	room, _ := rs.GetRoom("testroom")
	user := room.AddUser("testuser")

	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Name)
	assert.False(t, user.JoinedAt.IsZero())

	retrievedUser, exists := room.GetUser("testuser")
	assert.True(t, exists)
	assert.Equal(t, user, retrievedUser)
}

func TestRoom_RemoveUser(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "host",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}
	rs.AddRoom("testroom", game)

	room, _ := rs.GetRoom("testroom")
	room.AddUser("testuser")

	_, exists := room.GetUser("testuser")
	assert.True(t, exists)

	room.RemoveUser("testuser")

	_, exists = room.GetUser("testuser")
	assert.False(t, exists)
}

func TestRoom_ConcurrentUserOperations(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "host",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}
	rs.AddRoom("testroom", game)

	room, _ := rs.GetRoom("testroom")
	const numUsers = 2000
	var wg sync.WaitGroup

	// Add users concurrently
	for i := range numUsers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			username := fmt.Sprintf("user%d", id)
			room.AddUser(username)
		}(i)
	}

	wg.Wait()

	// Verify all users were added
	users := room.GetAllUsers()
	assert.Equal(t, numUsers, len(users))

	// Remove users concurrently
	for i := range numUsers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			username := fmt.Sprintf("user%d", id)
			room.RemoveUser(username)
		}(i)
	}

	wg.Wait()

	// Verify all users were removed
	users = room.GetAllUsers()
	assert.Equal(t, 0, len(users))
}

func TestRoom_UsersByJoinTime(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "host",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}
	rs.AddRoom("testroom", game)

	room, _ := rs.GetRoom("testroom")

	// Add users with slight delays to ensure different join times
	usernames := []string{"user1", "user2", "user3"}
	for _, username := range usernames {
		room.AddUser(username)
		time.Sleep(time.Millisecond * 5)
	}

	sortedUsers := room.UsersByJoinTime()
	assert.Equal(t, 3, len(sortedUsers))

	// Verify users are sorted by join time (earliest first)
	for i := 0; i < len(sortedUsers)-1; i++ {
		assert.True(t, sortedUsers[i].JoinedAt.Before(sortedUsers[i+1].JoinedAt) ||
			sortedUsers[i].JoinedAt.Equal(sortedUsers[i+1].JoinedAt))
	}

	// Verify the correct order based on our insertion
	assert.Equal(t, "user1", sortedUsers[0].Name)
	assert.Equal(t, "user2", sortedUsers[1].Name)
	assert.Equal(t, "user3", sortedUsers[2].Name)
}

func TestRoom_GetAllUsers(t *testing.T) {
	rs := NewRoomService()
	game := &types.GameSession{
		Host:        "host",
		MovieNumber: 5,
		MaxPlayers:  4,
		Votes:       make(map[*types.JellyfinItem]int),
	}
	rs.AddRoom("testroom", game)

	room, _ := rs.GetRoom("testroom")

	// Empty room
	users := room.GetAllUsers()
	assert.Equal(t, 0, len(users))

	// Add some users
	room.AddUser("user1")
	room.AddUser("user2")

	users = room.GetAllUsers()
	assert.Equal(t, 2, len(users))

	// Verify users are included (order may vary due to map iteration)
	usernames := make([]string, len(users))
	for i, user := range users {
		usernames[i] = user.Name
	}
	assert.Contains(t, usernames, "user1")
	assert.Contains(t, usernames, "user2")
}

