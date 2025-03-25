CREATE TABLE IF NOT EXISTS encounter_players (
    id SERIAL PRIMARY KEY,
    encounter_id INTEGER REFERENCES encounters(id) ON DELETE CASCADE,
    player_id INTEGER REFERENCES players(id) ON DELETE CASCADE,
    initiative INTEGER DEFAULT 0,
    hp INTEGER DEFAULT 1
);
