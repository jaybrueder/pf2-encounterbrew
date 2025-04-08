-- Add name column and unique constraint to conditions
ALTER TABLE conditions ADD COLUMN name VARCHAR(255);
UPDATE conditions SET name = data->>'name'; -- Attempt to populate from existing JSON data's 'name' field
ALTER TABLE conditions ALTER COLUMN name SET NOT NULL;
ALTER TABLE conditions ADD CONSTRAINT conditions_name_unique UNIQUE (name);
CREATE INDEX idx_conditions_name ON conditions(name); -- Index for faster lookups

-- Add name column and unique constraint to monsters
ALTER TABLE monsters ADD COLUMN name VARCHAR(255);
UPDATE monsters SET name = data->>'name'; -- Attempt to populate from existing JSON data's 'name' field
ALTER TABLE monsters ALTER COLUMN name SET NOT NULL;
ALTER TABLE monsters ADD CONSTRAINT monsters_name_unique UNIQUE (name);
CREATE INDEX idx_monsters_name ON monsters(name); -- Index for faster lookups

-- Add unique constraint for party name per user
-- Assuming user_id cannot be NULL based on schema intent
ALTER TABLE parties ADD CONSTRAINT parties_name_user_id_unique UNIQUE (name, user_id);

-- Add unique constraint for player name within a party
-- Assuming party_id cannot be NULL based on schema intent
ALTER TABLE players ADD CONSTRAINT players_name_party_id_unique UNIQUE (name, party_id);
