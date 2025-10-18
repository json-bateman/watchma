package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository struct {
	db *sql.DB
	l  *slog.Logger
}

func NewUserRepository(db *sql.DB, l *slog.Logger) *UserRepository {
	return &UserRepository{db: db, l: l}
}

func (r *UserRepository) CreateUser(username, passwordHash string) (*User, error) {
	now := time.Now()
	query := `
                INSERT INTO users (username, password_hash, created_at, updated_at)
				VALUES (?, ?, ?, ?)
        `
	result, err := r.db.Exec(query, username, passwordHash, now, now)
	if err != nil {
		r.l.Error("Could not CreateUser", "err", err)
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the complete user object
	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (r *UserRepository) GetByUsername(username string) (*User, error) {

	query := `
                SELECT id, username, password_hash, created_at, updated_at
                FROM users
                WHERE username = ?
        `

	var user User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		s := fmt.Sprintf("query user: %s --- Not found", username)

		r.l.Error(s)
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil

}
