-- Create or replace the upsert_player_from_profile function
CREATE OR REPLACE FUNCTION upsert_player_from_profile(
    p_auth_user_id UUID,
    p_username TEXT
) RETURNS UUID AS $$
DECLARE
    v_player_id UUID;
BEGIN
    -- Try to update existing player
    UPDATE players
    SET 
        username = p_username,
        auth_user_id = p_auth_user_id,
        is_guest = false,
        updated_at = NOW()
    WHERE auth_user_id = p_auth_user_id
    RETURNING id INTO v_player_id;

    -- If no update occurred, insert new player
    IF v_player_id IS NULL THEN
        INSERT INTO players (id, username, auth_user_id, is_guest, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            p_username,
            p_auth_user_id,
            false,
            NOW(),
            NOW()
        )
        RETURNING id INTO v_player_id;
    END IF;

    RETURN v_player_id;
END;
$$ LANGUAGE plpgsql;
