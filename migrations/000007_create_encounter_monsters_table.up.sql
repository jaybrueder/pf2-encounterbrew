CREATE TABLE IF NOT EXISTS encounter_monsters (
    encounter_id INTEGER REFERENCES encounters(id),
    monster_id INTEGER REFERENCES monsters(id),
    adjustment INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (encounter_id, monster_id)
);
