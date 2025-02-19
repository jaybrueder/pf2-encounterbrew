CREATE TABLE IF NOT EXISTS encounters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id INTEGER REFERENCES users(id),
    party_id INTEGER REFERENCES parties(id)
);
