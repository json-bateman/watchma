package config

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Connect establishes and returns a persistent NATS connection using the default URL.
// The caller is responsible for closing the connection during application shutdown.
func Connect() *nats.Conn {
	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.Name("Room-Chat"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	return nc
}
