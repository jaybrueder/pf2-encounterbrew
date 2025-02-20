-- First, drop the existing foreign key constraints
ALTER TABLE encounter_players
DROP CONSTRAINT encounter_players_encounter_id_fkey;

-- Then recreate them with CASCADE delete
ALTER TABLE encounter_players
ADD CONSTRAINT encounter_players_encounter_id_fkey
FOREIGN KEY (encounter_id)
REFERENCES encounters(id)
ON DELETE CASCADE;

-- Do the same for any other tables that reference encounters
-- For example, if you have encounter_monsters:
ALTER TABLE encounter_monsters
DROP CONSTRAINT encounter_monsters_encounter_id_fkey;

ALTER TABLE encounter_monsters
ADD CONSTRAINT encounter_monsters_encounter_id_fkey
FOREIGN KEY (encounter_id)
REFERENCES encounters(id)
ON DELETE CASCADE;
