package repository

import (
	"context"
	"database/sql"
	"time"

	"watchma/pkg/database/sqlcgen"
)

type SessionRepository struct {
	queries *sqlcgen.Queries
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		queries: sqlcgen.New(db),
	}
}

// Create stores a session token for a user
func (r *SessionRepository) Create(userID int64, token string) error {
	ctx := context.Background()
	expiresAt := time.Now().AddDate(100, 0, 0) // far future (never expires)

	return r.queries.CreateSession(ctx, sqlcgen.CreateSessionParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

// GetUserID returns the user ID for a given session token
func (r *SessionRepository) GetUserID(token string) (int64, error) {
	ctx := context.Background()
	return r.queries.GetUserIDByToken(ctx, token)
}

// Delete removes a session (for logout)
func (r *SessionRepository) Delete(token string) error {
	ctx := context.Background()
	return r.queries.DeleteSession(ctx, token)
}
