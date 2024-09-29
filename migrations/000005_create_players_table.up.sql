CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    level INTEGER DEFAULT 1,
    hp INTEGER DEFAULT 10,
    ac INTEGER DEFAULT 12,
    party_id INTEGER REFERENCES parties(id)
);

-- Add default seeds for development purposes
WITH gold_party AS (
    SELECT id FROM parties WHERE name = 'Gold' ORDER BY id LIMIT 1
)
INSERT INTO players (name, party_id)
VALUES
    ('Poppy', (SELECT id FROM gold_party)),
    ('Pad Lannis', (SELECT id FROM gold_party)),
    ('Farouk', (SELECT id FROM gold_party)),
    ('Torchwood', (SELECT id FROM gold_party)),
    ('Dendrobium', (SELECT id FROM gold_party))
ON CONFLICT (id) DO NOTHING;

WITH red_party AS (
    SELECT id FROM parties WHERE name = 'Red' ORDER BY id LIMIT 1
)
INSERT INTO players (name, party_id)
VALUES
    ('Fridhild', (SELECT id FROM red_party)),
    ('Donny Deephelm', (SELECT id FROM red_party)),
    ('Ardon Grayle', (SELECT id FROM red_party)),
    ('Mnpingo', (SELECT id FROM red_party))
ON CONFLICT (id) DO NOTHING;
