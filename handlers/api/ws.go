package api

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// For dev only. CORS stuff.
	// TODO: change this by prod.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func GameWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Read error:", "error", err)
			break
		}
		slog.Log(context.Background(), slog.Level(2), "hi from ws")

		// Echo message back to client (for now)
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			slog.Error("Write error:", "error", err)
			break
		}
	}
}
