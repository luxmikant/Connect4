-- Create profiles table for additional user data
-- NOTE: This migration is designed to work with or without Supabase auth schema
-- When auth schema exists (Supabase), it creates full integration
-- When auth schema doesn't exist (standard PostgreSQL), it creates a standalone profiles table

-- Check if we're in a Supabase environment (auth schema exists)
DO $$
BEGIN
  -- Only create the profiles table with auth.users reference if auth schema exists
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'auth') THEN
    -- Supabase environment: Create with auth.users reference
    CREATE TABLE IF NOT EXISTS public.profiles (
      id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
      username VARCHAR(50) UNIQUE NOT NULL,
      avatar_url TEXT,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
      
      CONSTRAINT username_length CHECK (char_length(username) >= 3 AND char_length(username) <= 20),
      CONSTRAINT username_format CHECK (username ~ '^[a-zA-Z0-9_]+$')
    );

    -- Enable Row Level Security
    ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

    -- RLS Policies (only work with Supabase auth)
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'profiles' AND policyname = 'Profiles are viewable by everyone') THEN
      CREATE POLICY "Profiles are viewable by everyone"
        ON public.profiles FOR SELECT
        USING (true);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'profiles' AND policyname = 'Users can insert own profile') THEN
      CREATE POLICY "Users can insert own profile"
        ON public.profiles FOR INSERT
        WITH CHECK (auth.uid() = id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'profiles' AND policyname = 'Users can update own profile') THEN
      CREATE POLICY "Users can update own profile"
        ON public.profiles FOR UPDATE
        USING (auth.uid() = id);
    END IF;

    -- Function to handle new user signup
    CREATE OR REPLACE FUNCTION public.handle_new_user()
    RETURNS TRIGGER AS $func$
    BEGIN
      INSERT INTO public.profiles (id, username, avatar_url)
      VALUES (
        new.id,
        COALESCE(
          new.raw_user_meta_data->>'username',
          split_part(new.email, '@', 1)
        ),
        new.raw_user_meta_data->>'avatar_url'
      );
      RETURN new;
    END;
    $func$ LANGUAGE plpgsql SECURITY DEFINER;

    -- Trigger to auto-create profile on user signup
    DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
    CREATE TRIGGER on_auth_user_created
      AFTER INSERT ON auth.users
      FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

    RAISE NOTICE 'Profiles table created with Supabase auth integration';
  ELSE
    -- Standard PostgreSQL environment: Skip this migration
    -- The players table already handles user data
    RAISE NOTICE 'Skipping profiles migration - auth schema not found (non-Supabase environment)';
  END IF;
END $$;

-- Create index on username for fast lookups (only if table exists)
CREATE INDEX IF NOT EXISTS idx_profiles_username ON public.profiles(username);

-- Function to update updated_at timestamp (general purpose, always create)
CREATE OR REPLACE FUNCTION public.handle_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  new.updated_at = CURRENT_TIMESTAMP;
  RETURN new;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at (only if table exists)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'profiles') THEN
    DROP TRIGGER IF EXISTS on_profile_updated ON public.profiles;
    CREATE TRIGGER on_profile_updated
      BEFORE UPDATE ON public.profiles
      FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();
  END IF;
END $$;
