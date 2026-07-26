package ws

import (
	"testing"
)

func TestHub(t *testing.T) {
	hub := &Hub{
		rooms: make(map[string]map[*Client]bool),
	}

	room := "test_room"
	hub.Broadcast(room, "hello")
}
