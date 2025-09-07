package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/nats-io/nats.go"
)

// NatsConnect establishes and returns a persistent NATS connection using the default URL.
// The caller is responsible for closing the connection during application shutdown.
func NatsConnect(l *slog.Logger) *nats.Conn {
	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.Name("jellyfin-grabber"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		l.Error("NATS Connection Error:",
			"Error", err,
		)
		os.Exit(1)
	}
	l.Info("Nats Configuration",
		"Name", utils.NATS_NAME,
		"URL", nats.DefaultURL,
	)
	return nc
}
