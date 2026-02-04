-- Account cache for GetAccountByRiotID
CREATE TABLE riot_account_cache (
    id BIGSERIAL PRIMARY KEY,
    game_name TEXT NOT NULL,
    tag_line TEXT NOT NULL,
    region TEXT NOT NULL,
    puuid TEXT NOT NULL,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE (game_name, tag_line, region)
);

CREATE INDEX idx_riot_account_cache_expires ON riot_account_cache(expires_at);

-- Game cache for spectator API
CREATE TABLE riot_game_cache (
    id BIGSERIAL PRIMARY KEY,
    puuid TEXT NOT NULL,
    region TEXT NOT NULL,
    in_game BOOLEAN NOT NULL,
    game_id TEXT,
    participants JSONB,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE (puuid, region)
);

CREATE INDEX idx_riot_game_cache_expires ON riot_game_cache(expires_at);
