package app

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// StartEmbeddedNATS starts an embedded NATS server and returns a client connection
func StartEmbeddedNATS(logger *slog.Logger) (*server.Server, *nats.Conn, error) {
	// Configure the embedded NATS server
	opts := &server.Options{
		ServerName:    "watchma-nats-embedded",
		Host:          "127.0.0.1",
		Port:          4222,
		HTTPPort:      8222, // Monitoring port
		NoLog:         false,
		NoSigs:        true,        // Disable signal handling (parent app handles it)
		MaxPayload:    1024 * 1024, // 1MB
		MaxConn:       500,
		WriteDeadline: 10 * time.Second,
	}

	// Enable debug/trace logging if desired
	if logger != nil {
		opts.Debug = true
		opts.Trace = true
	}

	// Create and start the NATS server
	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("create NATS server: %w", err)
	}

	// Start the server in a goroutine
	go ns.Start()

	// Wait for the server to be ready
	if !ns.ReadyForConnections(4 * time.Second) {
		return nil, nil, fmt.Errorf("NATS server not ready after 4 seconds")
	}

	logger.Info("Embedded NATS server started",
		"host", opts.Host,
		"port", opts.Port,
		"http_port", opts.HTTPPort,
	)

	// Connect a client to the embedded server
	nc, err := nats.Connect(
		fmt.Sprintf("nats://%s:%d", opts.Host, opts.Port),
		nats.MaxReconnects(-1), // Unlimited reconnects
	)
	if err != nil {
		ns.Shutdown()
		return nil, nil, fmt.Errorf("connect to embedded NATS: %w", err)
	}

	// Verify the connection
	if err := nc.FlushTimeout(2 * time.Second); err != nil {
		ns.Shutdown()
		nc.Close()
		return nil, nil, fmt.Errorf("NATS flush failed: %w", err)
	}

	logger.Info("Connected to embedded NATS",
		"url", nc.ConnectedUrl(),
		"server_id", nc.ConnectedServerId(),
		"max_payload", nc.MaxPayload(),
	)

	return ns, nc, nil
}
