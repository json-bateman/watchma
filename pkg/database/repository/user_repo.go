package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"watchma/pkg/database/sqlcgen"
)

// User is an alias to the sqlc generated User type for backward compatibility
type User = sqlcgen.User

type UserRepository struct {
	queries *sqlcgen.Queries
	l       *slog.Logger
}

func NewUserRepository(db *sql.DB, l *slog.Logger) *UserRepository {
	return &UserRepository{
		queries: sqlcgen.New(db),
		l:       l,
	}
}

func (r *UserRepository) CreateUser(username, passwordHash string) (*User, error) {
	now := time.Now()
	ctx := context.Background()

	user, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		r.l.Error("Could not CreateUser", "err", err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*User, error) {
	ctx := context.Background()

	user, err := r.queries.GetUserByUsername(ctx, username)
	if err == sql.ErrNoRows {
		r.l.Error("query user: Not found", "username", username)
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}

// GetUserBySessionToken fetches a user by their session token using an optimized JOIN
func (r *UserRepository) GetUserBySessionToken(token string) (*User, error) {
	ctx := context.Background()

	user, err := r.queries.GetUserBySessionToken(ctx, token)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("query user by session token: %w", err)
	}

	return &user, nil
}
