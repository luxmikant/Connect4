-- Create moves table
CREATE TABLE IF NOT EXISTS moves (
    id VARCHAR(255) PRIMARY KEY,
    game_id VARCHAR(255) NOT NULL,
    player VARCHAR(10) NOT NULL,
    col INTEGER NOT NULL CHECK (col >= 0 AND col < 7),
    row INTEGER NOT NULL CHECK (row >= 0 AND row < 6),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_moves_game_id ON moves(game_id);
CREATE INDEX IF NOT EXISTS idx_moves_timestamp ON moves(timestamp);

-- Add foreign key constraint to game_sessions (only if it doesn't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_moves_game_id' 
        AND table_name = 'moves'
    ) THEN
        ALTER TABLE moves 
        ADD CONSTRAINT fk_moves_game_id 
        FOREIGN KEY (game_id) REFERENCES game_sessions(id) ON DELETE CASCADE;
    END IF;
END $$;