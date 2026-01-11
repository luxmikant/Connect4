-- Analytics snapshots table for storing aggregated metrics
-- This table stores point-in-time snapshots of game analytics (Requirement 10.5)

CREATE TABLE IF NOT EXISTS analytics_snapshots (
    id VARCHAR(36) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    games_completed_hour BIGINT NOT NULL DEFAULT 0,
    games_completed_day BIGINT NOT NULL DEFAULT 0,
    avg_game_duration_sec BIGINT NOT NULL DEFAULT 0,
    min_game_duration_sec BIGINT NOT NULL DEFAULT 0,
    max_game_duration_sec BIGINT NOT NULL DEFAULT 0,
    total_moves BIGINT NOT NULL DEFAULT 0,
    unique_players_hour BIGINT NOT NULL DEFAULT 0,
    active_games BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for time-based queries
CREATE INDEX IF NOT EXISTS idx_analytics_snapshots_timestamp ON analytics_snapshots(timestamp);

-- Index for efficient daily/hourly aggregation queries
CREATE INDEX IF NOT EXISTS idx_analytics_snapshots_created_at ON analytics_snapshots(created_at);

-- Comment on table
COMMENT ON TABLE analytics_snapshots IS 'Stores point-in-time snapshots of game analytics metrics';
COMMENT ON COLUMN analytics_snapshots.games_completed_hour IS 'Number of games completed in the last hour';
COMMENT ON COLUMN analytics_snapshots.games_completed_day IS 'Number of games completed in the last day';
COMMENT ON COLUMN analytics_snapshots.avg_game_duration_sec IS 'Average game duration in seconds';
COMMENT ON COLUMN analytics_snapshots.unique_players_hour IS 'Number of unique players in the last hour';
