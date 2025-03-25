CREATE TABLE IF NOT EXISTS combatant_conditions (
    id SERIAL PRIMARY KEY,
    encounter_id INTEGER REFERENCES encounters(id),
    encounter_player_id INTEGER REFERENCES encounter_players(id),
    encounter_monster_id INTEGER REFERENCES encounter_monsters(id),
    condition_id INTEGER REFERENCES conditions(id),
    condition_value INTEGER,
    CONSTRAINT chk_encounter_player_or_monster CHECK (
        (encounter_player_id IS NOT NULL AND encounter_monster_id IS NULL) OR
        (encounter_player_id IS NULL AND encounter_monster_id IS NOT NULL)
    )
);
