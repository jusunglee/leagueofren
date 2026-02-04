-- Subscriptions table
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    discord_channel_id TEXT NOT NULL,
    server_id TEXT,
    lol_username TEXT NOT NULL,
    region TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (discord_channel_id, lol_username, region)
);

CREATE INDEX idx_subscriptions_last_evaluated_at ON subscriptions(last_evaluated_at);

-- Evals table (tracks each polling check)
CREATE TABLE evals (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    game_id BIGINT,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    eval_status TEXT NOT NULL,
    discord_message_id TEXT,
    CHECK (eval_status IN ('OFFLINE', 'NEW_TRANSLATIONS', 'REUSE_TRANSLATIONS', 'NO_TRANSLATIONS'))
);

CREATE INDEX idx_evals_subscription_id ON evals(subscription_id);
CREATE INDEX idx_evals_evaluated_at ON evals(evaluated_at);
CREATE INDEX idx_evals_subscription_game ON evals(subscription_id, game_id);

-- Translations table (cached username translations)
CREATE TABLE translations (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    translation TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Translation to evals junction table
CREATE TABLE translation_to_evals (
    translation_id BIGINT NOT NULL REFERENCES translations(id) ON DELETE CASCADE,
    eval_id BIGINT NOT NULL REFERENCES evals(id) ON DELETE CASCADE,
    PRIMARY KEY (translation_id, eval_id)
);

CREATE INDEX idx_translation_to_evals_eval ON translation_to_evals(eval_id);

-- Feedback table
CREATE TABLE feedback (
    id BIGSERIAL PRIMARY KEY,
    discord_message_id TEXT NOT NULL,
    feedback_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Riot account cache for GetAccountByRiotID
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

-- Riot game cache for spectator API
CREATE TABLE riot_game_cache (
    id BIGSERIAL PRIMARY KEY,
    puuid TEXT NOT NULL,
    region TEXT NOT NULL,
    in_game BOOLEAN NOT NULL,
    game_id BIGINT,
    participants JSONB,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE (puuid, region)
);

CREATE INDEX idx_riot_game_cache_expires ON riot_game_cache(expires_at);
