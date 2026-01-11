-- Migration: Add optimized indexes for game session queries
-- This migration adds indexes to improve query performance for active session lookups

-- Index for active session lookups by status
CREATE INDEX IF NOT EXISTS idx_game_sessions_status ON game_sessions(status);

-- Composite index for player lookups with status filter
CREATE INDEX IF NOT EXISTS idx_game_sessions_player1_status ON game_sessions(player1, status);
CREATE INDEX IF NOT EXISTS idx_game_sessions_player2_status ON game_sessions(player2, status);

-- Index for timeout queries (status + updated_at)
CREATE INDEX IF NOT EXISTS idx_game_sessions_status_updated ON game_sessions(status, updated_at);

-- Index for game history queries (status + end_time)
CREATE INDEX IF NOT EXISTS idx_game_sessions_status_end_time ON game_sessions(status, end_time DESC);

-- Index for created_at ordering
CREATE INDEX IF NOT EXISTS idx_game_sessions_created_at ON game_sessions(created_at DESC);
