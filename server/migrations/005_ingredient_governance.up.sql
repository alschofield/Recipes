-- 005_ingredient_governance.up.sql
-- Candidate queue, voting, and governance support for ingredient quality.

CREATE TABLE IF NOT EXISTS ingredient_candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    raw_name VARCHAR(120) NOT NULL,
    normalized_name VARCHAR(120) NOT NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (source IN ('user', 'llm', 'system')),
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved_alias', 'approved_canonical', 'rejected')),
    confidence NUMERIC(4,3) NOT NULL DEFAULT 0,
    suggested_canonical_id UUID REFERENCES ingredients(id) ON DELETE SET NULL,
    proposed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resolution_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ingredient_candidates_status ON ingredient_candidates(status);
CREATE INDEX IF NOT EXISTS idx_ingredient_candidates_normalized ON ingredient_candidates(normalized_name);
CREATE INDEX IF NOT EXISTS idx_ingredient_candidates_source ON ingredient_candidates(source);

CREATE UNIQUE INDEX IF NOT EXISTS ux_ingredient_candidates_pending_name
ON ingredient_candidates(normalized_name)
WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS ingredient_candidate_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id UUID NOT NULL REFERENCES ingredient_candidates(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote SMALLINT NOT NULL CHECK (vote IN (-1, 1)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(candidate_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ingredient_candidate_votes_candidate ON ingredient_candidate_votes(candidate_id);

CREATE OR REPLACE TRIGGER ingredient_candidate_votes_updated_at
    BEFORE UPDATE ON ingredient_candidate_votes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
