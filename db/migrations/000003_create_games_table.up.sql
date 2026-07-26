CREATE TABLE IF NOT EXISTS games (
    id TEXT PRIMARY KEY,
    gm_code TEXT NOT NULL,
    invite_code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);
