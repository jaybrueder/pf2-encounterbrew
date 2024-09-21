CREATE TABLE IF NOT EXISTS conditions (
  id SERIAL PRIMARY KEY,
  data JSONB,
  CONSTRAINT data_not_null CHECK (data IS NOT NULL)
);

CREATE INDEX idx_conditions_data_gin ON conditions USING GIN (data);
