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
	Settings *config.Settings
	Logger   *slog.Logger
	Router   *chi.Mux
	NATS     *nats.Conn
	DB       *database.DB
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

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return fmt.Errorf("connect NATS: %w", err)
	}

	// Verify round-trip (ensures we’re really up)
	if err := nc.FlushTimeout(2 * time.Second); err != nil {
		return fmt.Errorf("nats flush failed: %w", err)
	}

	a.NATS = nc
	a.Logger.Info(
		"NATS connected",
		"url", nc.ConnectedUrl(),
		"server_id", nc.ConnectedServerId(),
		"max_payload", nc.MaxPayload(),
	)

	a.DB = db
	userRepo := repository.NewUserRepository(db.DB, a.Logger)
	sessionRepo := repository.NewSessionRepository(db.DB)

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

	eventPublisher := services.NewEventPublisher(a.NATS, a.Logger)
	authService := services.NewAuthService(userRepo, sessionRepo, a.Logger)
	movieService := services.NewMovieService(movieProvider)
	roomService := services.NewRoomService(eventPublisher, a.Logger)
	movieOfTheDayService := services.NewMovieOfTheDayService(movieService)

	webHandler := web.NewWebHandler(
		a.Settings,
		a.Logger,
		a.NATS,
		&web.WebHandlerServices{
			MovieService:         movieService,
			RoomService:          roomService,
			MovieOfTheDayService: movieOfTheDayService,
			AuthService:          authService,
		},
	)

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)
	webHandler.SetupRoutes(a.Router)
	config.SetupFileServer(a.Logger, a.Router)

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
