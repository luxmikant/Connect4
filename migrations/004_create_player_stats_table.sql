-- Create player_stats table
CREATE TABLE IF NOT EXISTS player_stats (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    games_played INTEGER NOT NULL DEFAULT 0,
    games_won INTEGER NOT NULL DEFAULT 0,
    win_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0,
    avg_game_time INTEGER NOT NULL DEFAULT 0,
    last_played TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index on username
CREATE UNIQUE INDEX IF NOT EXISTS idx_player_stats_username ON player_stats(username);

-- Create index for leaderboard queries (games_won DESC for ranking)
CREATE INDEX IF NOT EXISTS idx_player_stats_wins ON player_stats(games_won DESC);
CREATE INDEX IF NOT EXISTS idx_player_stats_win_rate ON player_stats(win_rate DESC);

-- Add trigger to update updated_at column
CREATE TRIGGER update_player_stats_updated_at BEFORE UPDATE ON player_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();