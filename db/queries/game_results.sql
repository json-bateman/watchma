-- name: CreateGameResult :one
INSERT INTO game_results (
    room_name,
    winning_movie_id,
    winning_movie_name,
    winning_vote_count,
    total_players
) VALUES (?,?,?,?,?)
RETURNING *;

-- name: CreateGameParticipant :one
INSERT INTO game_participants (
    game_id,
    user_id
) VALUES (?,?)
RETURNING *;

-- name: GetGameResultsByUser :many
SELECT gr.* FROM game_results gr
JOIN game_participants gp ON gr.id = gp.game_id
WHERE gp.user_id = ?
ORDER BY gr.completed_at DESC;

-- name: GetMostPopularWinningMovies :many
SELECT
    winning_movie_id,
    winning_movie_name,
    COUNT(*) as win_count,
    COALESCE(AVG(winning_vote_count), 0.0) as avg_votes
FROM game_results
GROUP BY winning_movie_id, winning_movie_name
ORDER BY win_count DESC
LIMIT ?;
