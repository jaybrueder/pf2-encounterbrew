-- Drop constraints first
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_name_party_id_unique;
ALTER TABLE parties DROP CONSTRAINT IF EXISTS parties_name_user_id_unique;
ALTER TABLE monsters DROP CONSTRAINT IF EXISTS monsters_name_unique;
ALTER TABLE conditions DROP CONSTRAINT IF EXISTS conditions_name_unique;

-- Drop indexes
DROP INDEX IF EXISTS idx_monsters_name;
DROP INDEX IF EXISTS idx_conditions_name;

-- Drop columns
ALTER TABLE monsters DROP COLUMN IF EXISTS name;
ALTER TABLE conditions DROP COLUMN IF EXISTS name;
