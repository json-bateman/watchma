package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"watchma/db"
	"watchma/db/repository"
	"watchma/pkg/auth"
	"watchma/pkg/jellyfin"
	"watchma/pkg/movie"
	"watchma/pkg/openai"
	"watchma/pkg/room"
	"watchma/web/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type App struct {
	Settings   *Settings
	Logger     *slog.Logger
	Router     *chi.Mux
	NATS       *nats.Conn
	NATSServer *server.Server // Embedded NATS server instance
	DB         *db.DB
}

func New() *App {
	return &App{
		Settings: LoadSettings(),
	}
}

func (a *App) Initialize() error {
	a.Logger = NewColorLog(a.Settings.LogLevel)
	slog.SetDefault(a.Logger)

	// Initialize database with goose migrations
	db, err := db.New("./watchma.db", a.Logger)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	ns, nc, err := StartEmbeddedNATS(a.Logger)
	if err != nil {
		return fmt.Errorf("start embedded NATS: %w", err)
	}

	var openAiProvider *openai.Provider
	if a.Settings.OpenAIApiKey != "" {
		openAiProvider = openai.NewProvider(
			a.Settings.OpenAIApiKey,
			a.Logger,
		)
		a.Logger.Info("OpenAI provider initialized")
	} else {
		a.Logger.Warn("OpenAI API key not found, AI features disabled")
	}

	a.NATS = nc
	a.NATSServer = ns

	a.DB = db
	userRepo := repository.NewUserRepository(db.DB, a.Logger)
	sessionRepo := repository.NewSessionRepository(db.DB)

	var movieProvider movie.Provider
	if a.Settings.UseDummyData {
		movieProvider = movie.NewDummyProvider()
	} else {
		movieProvider = movie.NewCachingProvider(
			jellyfin.NewJellyfinMovieProvider(
				a.Settings.JellyfinApiKey,
				a.Settings.JellyfinBaseURL,
				a.Logger),
			time.Minute)
	}

	eventPublisher := room.NewEventPublisher(a.NATS, a.Logger)
	authService := auth.NewAuthService(userRepo, sessionRepo, a.Logger, a.Settings.IsDev)
	movieService := movie.NewService(movieProvider, a.Logger)
	roomService := room.NewService(eventPublisher, a.Logger)

	webHandler := router.NewWebHandler(
		a.Settings.JellyfinBaseURL,
		a.Settings.JellyfinApiKey,
		a.Logger,
		a.NATS,
		&router.WebHandlerServices{
			MovieService:   movieService,
			RoomService:    roomService,
			AuthService:    authService,
			OpenAiProvider: openAiProvider,
		},
	)

	a.Router = chi.NewRouter()

	if a.Settings.IsDev {
		a.Router.Use(middleware.Logger)
	}

	webHandler.SetupRoutes(a.Router)
	SetupFileServer(a.Logger, a.Router)

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Settings.Port)

	defer func() {
		a.Logger.Info("Shutting down gracefully...")

		// Close NATS client connection first
		if a.NATS != nil {
			a.NATS.Close()
			a.Logger.Info("NATS client connection closed")
		}

		// Shutdown embedded NATS server
		if a.NATSServer != nil {
			a.NATSServer.Shutdown()
			a.NATSServer.WaitForShutdown()
			a.Logger.Info("Embedded NATS server shutdown")
		}

		// Close database
		if a.DB != nil {
			a.DB.Close()
			a.Logger.Info("Database closed")
		}
	}()

	port := fmt.Sprintf(":%d", a.Settings.Port)
	return http.ListenAndServe(port, a.Router)
}
