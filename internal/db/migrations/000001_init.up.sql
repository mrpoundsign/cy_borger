CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT DEFAULT '',
    salt TEXT DEFAULT '',
    handle TEXT DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS characters (
    id TEXT PRIMARY KEY,
    edit_code TEXT NOT NULL,
    game_id TEXT DEFAULT '',
    owner_id TEXT DEFAULT '',
    data_json TEXT NOT NULL,
    is_saved INTEGER DEFAULT 0,
    is_dead INTEGER DEFAULT 0,
    death_note TEXT DEFAULT '',
    died_at TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS games (
    id TEXT PRIMARY KEY,
    gm_code TEXT NOT NULL,
    invite_code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    owner_id TEXT DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
