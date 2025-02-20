-- Modify the parties table's foreign key constraint
ALTER TABLE parties
DROP CONSTRAINT IF EXISTS parties_user_id_fkey;

ALTER TABLE parties
ADD CONSTRAINT parties_user_id_fkey
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;

-- Modify the players table's foreign key constraint
ALTER TABLE players
DROP CONSTRAINT IF EXISTS players_party_id_fkey;

ALTER TABLE players
ADD CONSTRAINT players_party_id_fkey
FOREIGN KEY (party_id)
REFERENCES parties(id)
ON DELETE CASCADE;

-- If you have encounter_players table, modify its constraint too
ALTER TABLE encounter_players
DROP CONSTRAINT IF EXISTS encounter_players_player_id_fkey;

ALTER TABLE encounter_players
ADD CONSTRAINT encounter_players_player_id_fkey
FOREIGN KEY (player_id)
REFERENCES players(id)
ON DELETE CASCADE;

-- If you have encounters table with party_id, modify its constraint
ALTER TABLE encounters
DROP CONSTRAINT IF EXISTS encounters_party_id_fkey;

ALTER TABLE encounters
ADD CONSTRAINT encounters_party_id_fkey
FOREIGN KEY (party_id)
REFERENCES parties(id)
ON DELETE CASCADE;
