-- Drop the existing foreign key constraint
ALTER TABLE combatant_conditions 
DROP CONSTRAINT combatant_conditions_encounter_id_fkey;

-- Add it back with ON DELETE CASCADE
ALTER TABLE combatant_conditions
ADD CONSTRAINT combatant_conditions_encounter_id_fkey 
FOREIGN KEY (encounter_id) 
REFERENCES encounters(id) 
ON DELETE CASCADE;

-- Also add CASCADE to the encounter_player_id and encounter_monster_id constraints
ALTER TABLE combatant_conditions 
DROP CONSTRAINT combatant_conditions_encounter_player_id_fkey;

ALTER TABLE combatant_conditions
ADD CONSTRAINT combatant_conditions_encounter_player_id_fkey 
FOREIGN KEY (encounter_player_id) 
REFERENCES encounter_players(id) 
ON DELETE CASCADE;

ALTER TABLE combatant_conditions 
DROP CONSTRAINT combatant_conditions_encounter_monster_id_fkey;

ALTER TABLE combatant_conditions
ADD CONSTRAINT combatant_conditions_encounter_monster_id_fkey 
FOREIGN KEY (encounter_monster_id) 
REFERENCES encounter_monsters(id) 
ON DELETE CASCADE;