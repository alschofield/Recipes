DROP INDEX IF EXISTS idx_user_refresh_sessions_client_session;

ALTER TABLE user_refresh_sessions
    DROP COLUMN IF EXISTS client_session_id;
