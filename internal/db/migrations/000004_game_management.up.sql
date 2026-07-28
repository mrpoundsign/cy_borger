CREATE TABLE IF NOT EXISTS systems (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO systems (id, name) VALUES ('0', 'CY_BORG') ON CONFLICT(id) DO NOTHING;

ALTER TABLE games ADD COLUMN is_locked INTEGER DEFAULT 0;
ALTER TABLE games ADD COLUMN system_id TEXT DEFAULT '0';
ALTER TABLE characters ADD COLUMN system_id TEXT DEFAULT '0';

CREATE TABLE IF NOT EXISTS game_gms (
    game_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (game_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_bans (
    owner_id TEXT NOT NULL,
    banned_user_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (owner_id, banned_user_id)
);
