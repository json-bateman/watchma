package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"log/slog"
	"time"
	"watchma/db/sqlcgen"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "watchma_session"
)

type AuthService struct {
	queries *sqlcgen.Queries
	logger  *slog.Logger
	IsDev   bool
}

func NewAuthService(queries *sqlcgen.Queries, logger *slog.Logger, isDev bool) *AuthService {
	return &AuthService{
		queries: queries,
		logger:  logger,
		IsDev:   isDev,
	}
}

func (s *AuthService) LoginOrCreate(username, password string) (*sqlcgen.User, string, error) {
	ctx := context.Background()
	user, err := s.queries.GetUserByUsername(ctx, username)

	if err == sql.ErrNoRows {
		// User doesn't exist - create them
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user, err = s.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
			Username:     username,
			PasswordHash: string(hash),
		})
		if err != nil {
			return nil, "", err
		}
		s.logger.Info("New user created", "username", username)
	} else if err != nil {
		return nil, "", err
	} else {
		// User exists - verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash),
			[]byte(password)); err != nil {
			return nil, "", errors.New("Invalid password for existing user: " + user.Username)
		}
		s.logger.Info("User logged in", "username", username)
	}

	token := generateRandomToken()
	if err := s.queries.CreateSession(ctx, sqlcgen.CreateSessionParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}); err != nil {
		return nil, "", err
	}

	return &user, token, nil

}

// GetUserBySessionToken fetches a user by their session token
func (s *AuthService) GetUserBySessionToken(token string) (*sqlcgen.User, error) {
	ctx := context.Background()
	user, err := s.queries.GetUserBySessionToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func generateRandomToken() string {
	b := make([]byte, 32) // 32 bytes = 256 bits of randomness
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
