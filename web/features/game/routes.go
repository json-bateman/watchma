package game

import (
	"log/slog"

	"watchma/pkg/movie"
	"watchma/pkg/openai"
	"watchma/pkg/room"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

func SetupRoutes(
	r chi.Router,
	roomService *room.Service,
	movieService *movie.Service,
	openAiProvider *openai.Provider,
	logger *slog.Logger,
	nats *nats.Conn,
) error {
	handlers := newHandlers(roomService, movieService, openAiProvider, logger, nats)

	// Lobby
	r.Get("/room/{roomName}/lobby", handlers.singleRoom)
	r.Get("/sse/{roomName}", handlers.singleRoomSSE)
	r.Post("/message", handlers.publishChatMessage)
	r.Post("/room/{roomName}/ready", handlers.ready)
	r.Post("/room/{roomName}/start", handlers.startGame)
	r.Post("/room/{roomName}/leave", handlers.leaveRoom)

	// Draft
	r.Get("/room/{roomName}/draft", handlers.draft)
	r.Post("/draft/{roomName}/submit", handlers.draftSubmit)
	r.Post("/draft/{roomName}/query", handlers.queryMovies)
	r.Patch("/draft/{roomName}/{id}", handlers.toggleDraftMovie)
	r.Delete("/draft/{roomName}/{id}", handlers.deleteFromSelectedMovies)

	// Voting
	r.Get("/room/{roomName}/voting", handlers.voting)
	r.Post("/voting/{roomName}/submit", handlers.votingSubmit)
	r.Patch("/voting/{roomName}/{id}", handlers.toggleVotingMovie)

	// Results
	r.Get("/room/{roomName}/results", handlers.results)

	return nil
}
