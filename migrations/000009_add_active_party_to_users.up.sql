ALTER TABLE users
ADD COLUMN active_party_id INTEGER REFERENCES parties(id);
