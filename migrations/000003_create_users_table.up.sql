CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
    -- Add other user fields as needed
);

-- Add default seeds for development purposes
INSERT INTO users (name)
VALUES ('Default')
ON CONFLICT (id) DO NOTHING;
