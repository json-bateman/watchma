package repository

import (
	"database/sql"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create stores a session token for a user
func (r *SessionRepository) Create(userID int64, token string) error {
	expiresAt := time.Now().AddDate(100, 0, 0) // far future (never expires)
	_, err := r.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	return err
}

// GetUserID returns the user ID for a given session token
func (r *SessionRepository) GetUserID(token string) (int64, error) {
	var userID int64
	err := r.db.QueryRow(
		"SELECT user_id FROM refresh_tokens WHERE token = ?",
		token,
	).Scan(&userID)
	return userID, err
}

// Delete removes a session (for logout)
func (r *SessionRepository) Delete(token string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE token = ?", token)
	return err
}
