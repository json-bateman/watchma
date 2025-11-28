-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS game_results (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  room_name TEXT NOT NULL,
  winning_movie_id TEXT NOT NULL,
  winning_movie_name TEXT NOT NULL,
  winning_vote_count INTEGER NOT NULL,
  total_players INTEGER NOT NULL,
  completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS game_participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    FOREIGN KEY (game_id) REFERENCES game_results (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_results;
DROP TABLE IF EXISTS game_participants;
-- +goose StatementEnd
