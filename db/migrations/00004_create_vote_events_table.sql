-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vote_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    event_type TEXT NOT NULL CHECK(event_type IN ('draft_toggle', 'vote_toggle')),
    action TEXT NOT NULL CHECK(action IN ('selected', 'deselected')),
    movie_id TEXT NOT NULL,
    movie_name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vote_events;
-- +goose StatementEnd

