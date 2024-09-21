CREATE TABLE IF NOT EXISTS monsters (
  id SERIAL PRIMARY KEY,
  data JSONB,
  CONSTRAINT data_not_null CHECK (data IS NOT NULL)
);

CREATE INDEX idx_monster_data_gin ON monsters USING GIN (data);
