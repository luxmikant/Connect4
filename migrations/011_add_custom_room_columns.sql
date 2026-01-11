-- Add missing columns to game_sessions table for custom room functionality
ALTER TABLE game_sessions
ADD COLUMN IF NOT EXISTS room_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS is_custom BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255);

-- Create index for room_code lookups
CREATE INDEX IF NOT EXISTS idx_game_sessions_room_code ON game_sessions(room_code);
CREATE INDEX IF NOT EXISTS idx_game_sessions_is_custom ON game_sessions(is_custom);
