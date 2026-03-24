ALTER TABLE user_refresh_sessions
    ADD COLUMN IF NOT EXISTS client_session_id VARCHAR(128);

UPDATE user_refresh_sessions
SET client_session_id = COALESCE(client_session_id, family_id)
WHERE client_session_id IS NULL;

ALTER TABLE user_refresh_sessions
    ALTER COLUMN client_session_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_client_session
    ON user_refresh_sessions(user_id, client_session_id, revoked_at);
