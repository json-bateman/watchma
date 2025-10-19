package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"watchma/pkg/config"
	"watchma/pkg/database"
	"watchma/pkg/database/repository"
	"watchma/pkg/handlers/web"
	"watchma/pkg/providers"
	"watchma/pkg/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
)

type App struct {
	Settings     *config.Settings
	Logger       *slog.Logger
	Router       *chi.Mux
	MovieService *services.MovieService
	NATS         *nats.Conn
	DB           *database.DB
	UserRepo     *repository.UserRepository
	SessionRepo  *repository.SessionRepository
	AuthService  *services.AuthService
}

func New() *App {
	return &App{
		Settings: config.LoadSettings(),
	}
}

func (a *App) Initialize() error {
	a.Logger = config.NewColorLog(a.Settings.LogLevel)
	slog.SetDefault(a.Logger)

	// Initialize database with goose migrations
	db, err := database.New("./watchma.db", a.Logger)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	a.DB = db
	a.UserRepo = repository.NewUserRepository(db.DB, a.Logger)
	a.SessionRepo = repository.NewSessionRepository(db.DB)
	a.AuthService = services.NewAuthService(a.UserRepo, a.SessionRepo, a.Logger)

	var movieProvider providers.MovieProvider
	if a.Settings.UseDummyData {
		movieProvider = providers.NewDummyMovieProvider()
	} else {

		movieProvider = providers.NewCachingMovieProvider(
			providers.NewJellyfinMovieProvider(
				a.Settings.JellyfinApiKey,
				a.Settings.JellyfinBaseURL,
				a.Logger),
			time.Minute)
	}
	a.MovieService = services.NewMovieService(movieProvider)

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)

	config.SetupFileServer(a.Logger, a.Router)

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return fmt.Errorf("connect NATS: %w", err)
	}
	// Verify round-trip (ensures we’re really up)
	if err := nc.FlushTimeout(2 * time.Second); err != nil {
		return fmt.Errorf("nats flush failed: %w", err)
	}
	a.NATS = nc
	a.Logger.Info("NATS connected",
		"url", nc.ConnectedUrl(), "server_id", nc.ConnectedServerId(),
		"max_payload", nc.MaxPayload())

	eventPublisher := services.NewEventPublisher(a.NATS, a.Logger)
	roomService := services.NewRoomService(eventPublisher, a.Logger)
	movieOfTheDayService := services.NewMovieOfTheDayService(a.MovieService)
	webHandler := web.NewWebHandler(a.Settings, a.MovieService, a.Logger, roomService, movieOfTheDayService, a.AuthService, a.NATS)
	webHandler.SetupRoutes(a.Router)

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Settings.Port)

	defer func() {
		if a.DB != nil {
			a.DB.Close()
		}
		if a.NATS != nil {
			a.NATS.Close()
		}
	}()

	port := fmt.Sprintf(":%d", a.Settings.Port)
	return http.ListenAndServe(port, a.Router)
}
