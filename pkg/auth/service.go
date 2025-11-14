package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"log/slog"
	"watchma/db/repository"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "watchma_session"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	logger      *slog.Logger
	IsDev       bool
}

func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, logger *slog.Logger, isDev bool) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
		IsDev:       isDev,
	}
}

func (s *AuthService) LoginOrCreate(username, password string) (*repository.User, string, error) {
	user, err := s.userRepo.GetByUsername(username)

	if err == sql.ErrNoRows {
		// User doesn't exist - create them
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user, err = s.userRepo.CreateUser(username, string(hash))
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
	if err := s.sessionRepo.Create(user.ID, token); err != nil {
		return nil, "", err
	}

	return user, token, nil

}

// GetUserBySessionToken fetches a user by their session token
func (s *AuthService) GetUserBySessionToken(token string) (*repository.User, error) {
	return s.userRepo.GetUserBySessionToken(token)
}

func generateRandomToken() string {
	b := make([]byte, 32) // 32 bytes = 256 bits of randomness
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
