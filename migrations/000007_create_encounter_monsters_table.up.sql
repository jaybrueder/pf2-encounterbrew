CREATE TABLE IF NOT EXISTS encounter_monsters (
    id SERIAL PRIMARY KEY,
    encounter_id INTEGER REFERENCES encounters(id),
    monster_id INTEGER REFERENCES monsters(id),
    initiative INTEGER DEFAULT 0,
    level_adjustment INTEGER DEFAULT 0,
    hp INTEGER DEFAULT 1
);
