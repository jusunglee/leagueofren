-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    discord_channel_id TEXT NOT NULL,
    server_id TEXT NOT NULL,
    lol_username TEXT NOT NULL,
    region TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_evaluated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (discord_channel_id, lol_username, region)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_last_evaluated_at ON subscriptions(last_evaluated_at);

-- Evals table (tracks each polling check)
CREATE TABLE IF NOT EXISTS evals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    game_id INTEGER,
    evaluated_at TEXT NOT NULL DEFAULT (datetime('now')),
    eval_status TEXT NOT NULL CHECK (eval_status IN ('OFFLINE', 'NEW_TRANSLATIONS', 'REUSE_TRANSLATIONS', 'NO_TRANSLATIONS')),
    discord_message_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_evals_subscription_id ON evals(subscription_id);
CREATE INDEX IF NOT EXISTS idx_evals_evaluated_at ON evals(evaluated_at);
CREATE INDEX IF NOT EXISTS idx_evals_subscription_game ON evals(subscription_id, game_id);

-- Translations table (cached username translations)
CREATE TABLE IF NOT EXISTS translations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    translation TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Translation to evals junction table
CREATE TABLE IF NOT EXISTS translation_to_evals (
    translation_id INTEGER NOT NULL REFERENCES translations(id) ON DELETE CASCADE,
    eval_id INTEGER NOT NULL REFERENCES evals(id) ON DELETE CASCADE,
    PRIMARY KEY (translation_id, eval_id)
);

CREATE INDEX IF NOT EXISTS idx_translation_to_evals_eval ON translation_to_evals(eval_id);

-- Feedback table
CREATE TABLE IF NOT EXISTS feedback (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    discord_message_id TEXT NOT NULL,
    feedback_text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Riot account cache for GetAccountByRiotID
CREATE TABLE IF NOT EXISTS riot_account_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_name TEXT NOT NULL,
    tag_line TEXT NOT NULL,
    region TEXT NOT NULL,
    puuid TEXT NOT NULL,
    cached_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL,
    UNIQUE (game_name, tag_line, region)
);

CREATE INDEX IF NOT EXISTS idx_riot_account_cache_expires ON riot_account_cache(expires_at);

-- Riot game cache for spectator API
CREATE TABLE IF NOT EXISTS riot_game_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    puuid TEXT NOT NULL,
    region TEXT NOT NULL,
    in_game INTEGER NOT NULL,
    game_id INTEGER,
    participants TEXT,
    cached_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL,
    UNIQUE (puuid, region)
);

CREATE INDEX IF NOT EXISTS idx_riot_game_cache_expires ON riot_game_cache(expires_at);
