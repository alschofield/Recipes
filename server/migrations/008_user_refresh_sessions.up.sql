CREATE TABLE IF NOT EXISTS user_refresh_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id UUID NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    token_hash VARCHAR(128) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    replaced_by_token_id UUID,
    user_agent TEXT,
    ip_address TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_user ON user_refresh_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_family ON user_refresh_sessions(family_id);
CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_active ON user_refresh_sessions(user_id, family_id, revoked_at);
