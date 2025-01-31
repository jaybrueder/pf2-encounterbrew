CREATE TABLE IF NOT EXISTS encounter_players (
    id SERIAL PRIMARY KEY,
    encounter_id INTEGER REFERENCES encounters(id),
    player_id INTEGER REFERENCES players(id),
    initiative INTEGER DEFAULT 0
);
