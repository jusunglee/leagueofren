-- Subscriptions table
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    discord_channel_id TEXT NOT NULL,
    server_id TEXT NOT NULL,
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
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
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

-- ===========================================
-- Companion Website Tables
-- ===========================================

-- Players table (player metadata, refreshed by worker)
CREATE TABLE players (
    username TEXT PRIMARY KEY,
    region TEXT NOT NULL,
    rank TEXT,
    top_champions TEXT,
    puuid TEXT,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_region ON players(region);

-- Public translations submitted by bot users (opt-in)
CREATE TABLE public_translations (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    translation TEXT NOT NULL,
    explanation TEXT,
    language TEXT NOT NULL,
    player_username TEXT NOT NULL REFERENCES players(username),
    source_bot_id TEXT,
    riot_verified BOOLEAN NOT NULL DEFAULT false,
    upvotes INT NOT NULL DEFAULT 0,
    downvotes INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_public_translations_username ON public_translations(username);
CREATE INDEX idx_public_translations_hot ON public_translations(upvotes, downvotes, created_at);
CREATE INDEX idx_public_translations_created ON public_translations(created_at);

-- IP-based vote tracking (no login required, one vote per IP per translation)
CREATE TABLE votes (
    id BIGSERIAL PRIMARY KEY,
    translation_id BIGINT NOT NULL REFERENCES public_translations(id) ON DELETE CASCADE,
    ip_hash TEXT NOT NULL,
    visitor_id TEXT NOT NULL,
    vote SMALLINT NOT NULL CHECK (vote IN (-1, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(translation_id, visitor_id)
);

CREATE INDEX idx_votes_translation ON votes(translation_id);

-- Public feedback on translations (visible in admin panel only)
CREATE TABLE public_feedback (
    id BIGSERIAL PRIMARY KEY,
    translation_id BIGINT NOT NULL REFERENCES public_translations(id) ON DELETE CASCADE,
    ip_hash TEXT NOT NULL,
    feedback_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_public_feedback_translation ON public_feedback(translation_id);
CREATE INDEX idx_public_feedback_created ON public_feedback(created_at);
