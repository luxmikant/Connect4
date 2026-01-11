-- Link existing players table to Supabase auth
ALTER TABLE players ADD COLUMN IF NOT EXISTS auth_user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL;
ALTER TABLE players ADD COLUMN IF NOT EXISTS is_guest BOOLEAN DEFAULT false;

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_players_auth_user_id ON players(auth_user_id);
CREATE INDEX IF NOT EXISTS idx_players_is_guest ON players(is_guest);

-- Update existing players to be marked as guests (since no auth exists yet)
UPDATE players SET is_guest = true WHERE auth_user_id IS NULL;

-- Add constraint: authenticated players must have auth_user_id
-- (This will be enforced in application logic, not DB constraint to allow guests)
