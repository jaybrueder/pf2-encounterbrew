CREATE TABLE IF NOT EXISTS parties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id INTEGER REFERENCES users(id)
);

ALTER TABLE users
ADD COLUMN active_party_id INTEGER REFERENCES parties(id);
