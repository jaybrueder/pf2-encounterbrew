CREATE TABLE IF NOT EXISTS encounters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id INTEGER REFERENCES users(id)
);

-- Add default seeds for development purposes
INSERT INTO encounters (name, user_id)
VALUES ('Default', (SELECT id FROM users WHERE name = 'Default' ORDER BY id LIMIT 1))
ON CONFLICT (id) DO NOTHING
RETURNING id;
