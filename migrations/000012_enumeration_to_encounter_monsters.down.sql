-- Remove the index first
DROP INDEX IF EXISTS idx_encounter_monsters_enum;
-- Remove the enumeration column
ALTER TABLE encounter_monsters DROP COLUMN IF EXISTS enumeration;
