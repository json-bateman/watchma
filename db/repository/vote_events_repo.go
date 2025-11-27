package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"watchma/db/sqlcgen"
)

type VoteEvent = sqlcgen.VoteEvent

type VoteRepository struct {
	queries *sqlcgen.Queries
	l       *slog.Logger
}

func NewVoteRepository(db *sql.DB, l *slog.Logger) *VoteRepository {
	return &VoteRepository{
		queries: sqlcgen.New(db),
		l:       l,
	}
}

func (r *VoteRepository) AddVoteEvent(userId int64, eventType, action, movieId, movieName string) (*VoteEvent, error) {
	ctx := context.Background()
	voteEvent, err := r.queries.CreateVoteEvent(ctx, sqlcgen.CreateVoteEventParams{
		UserID:    userId,
		EventType: eventType,
		Action:    action,
		MovieID:   movieId,
		MovieName: movieName,
	})
	if err != nil {
		r.l.Error("Could not AddVoteEvent", "err", err)
		return nil, err
	}

	return &voteEvent, nil
}
