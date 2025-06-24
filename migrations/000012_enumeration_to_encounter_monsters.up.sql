-- Add enumeration column to encounter_monsters table
ALTER TABLE encounter_monsters
ADD COLUMN enumeration INTEGER DEFAULT 0;
-- Index to improve query performance
CREATE INDEX idx_encounter_monsters_enum ON encounter_monsters(encounter_id, monster_id, enumeration);
