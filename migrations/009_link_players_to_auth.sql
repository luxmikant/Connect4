-- Link existing players table to Supabase auth (when available)
-- NOTE: This migration works with or without Supabase auth schema

DO $$
BEGIN
  -- Add auth_user_id column
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                 WHERE table_name = 'players' AND column_name = 'auth_user_id') THEN
    -- Check if auth schema exists for the foreign key
    IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'auth') THEN
      -- Supabase environment: Add column with FK reference
      ALTER TABLE players ADD COLUMN auth_user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL;
      RAISE NOTICE 'Added auth_user_id column with auth.users reference';
    ELSE
      -- Standard PostgreSQL: Add column without FK reference
      ALTER TABLE players ADD COLUMN auth_user_id UUID;
      RAISE NOTICE 'Added auth_user_id column without FK (non-Supabase environment)';
    END IF;
  END IF;

  -- Add is_guest column
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                 WHERE table_name = 'players' AND column_name = 'is_guest') THEN
    ALTER TABLE players ADD COLUMN is_guest BOOLEAN DEFAULT false;
    RAISE NOTICE 'Added is_guest column';
  END IF;
END $$;

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_players_auth_user_id ON players(auth_user_id);
CREATE INDEX IF NOT EXISTS idx_players_is_guest ON players(is_guest);

-- Update existing players to be marked as guests (since no auth exists yet)
UPDATE players SET is_guest = true WHERE auth_user_id IS NULL AND is_guest IS NOT true;

-- Add constraint: authenticated players must have auth_user_id
-- (This will be enforced in application logic, not DB constraint to allow guests)
