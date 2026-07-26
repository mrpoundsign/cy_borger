CREATE TABLE IF NOT EXISTS characters (
    id TEXT PRIMARY KEY,
    edit_code TEXT NOT NULL,
    game_id TEXT DEFAULT '',
    owner_id TEXT DEFAULT '',
    data_json TEXT NOT NULL
);
