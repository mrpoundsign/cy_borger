DROP TABLE IF EXISTS user_bans;
DROP TABLE IF EXISTS game_gms;
-- SQLite does not easily support dropping columns from existing tables, so we leave is_locked and system_id alone.
-- We also leave the systems table alone as it might be used elsewhere.
