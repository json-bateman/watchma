package main

import (
	"log"

	"github.com/json-bateman/jellyfin-grabber/internal/app"
)

func main() {
	application := app.New()

	if err := application.Initialize(); err != nil {
		log.Fatal("Failed to initialize app:", err)
	}

	if err := application.Run(); err != nil {
		log.Fatal("App failed:", err)
	}
}

