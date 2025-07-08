-- Drop the CASCADE constraints
ALTER TABLE combatant_conditions 
DROP CONSTRAINT combatant_conditions_encounter_id_fkey;

ALTER TABLE combatant_conditions
ADD CONSTRAINT combatant_conditions_encounter_id_fkey 
FOREIGN KEY (encounter_id) 
REFERENCES encounters(id);

ALTER TABLE combatant_conditions 
DROP CONSTRAINT combatant_conditions_encounter_player_id_fkey;

ALTER TABLE combatant_conditions
ADD CONSTRAINT combatant_conditions_encounter_player_id_fkey 
FOREIGN KEY (encounter_player_id) 
REFERENCES encounter_players(id);

ALTER TABLE combatant_conditions 
DROP CONSTRAINT combatant_conditions_encounter_monster_id_fkey;

ALTER TABLE combatant_conditions
ADD CONSTRAINT combatant_conditions_encounter_monster_id_fkey 
FOREIGN KEY (encounter_monster_id) 
REFERENCES encounter_monsters(id);