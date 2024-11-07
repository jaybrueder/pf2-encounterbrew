-- Index for users table
CREATE INDEX idx_users_id ON users(id);

-- Index for parties table
CREATE INDEX idx_parties_user_id ON parties(user_id);

-- Index for players table
CREATE INDEX idx_players_party_id ON players(party_id);

-- Index for encounters table
CREATE INDEX idx_encounters_user_id ON encounters(user_id);
