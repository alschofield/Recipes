-- 005_ingredient_governance.down.sql

DROP TRIGGER IF EXISTS ingredient_candidate_votes_updated_at ON ingredient_candidate_votes;
DROP TABLE IF EXISTS ingredient_candidate_votes;
DROP INDEX IF EXISTS ux_ingredient_candidates_pending_name;
DROP TABLE IF EXISTS ingredient_candidates;
