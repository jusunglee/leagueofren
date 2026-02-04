-- subscriptions table
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    discord_channel_id TEXT NOT NULL,
    lol_username TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(discord_channel_id, lol_username)
);

-- evals table (tracks each polling check)
CREATE TABLE evals (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    eval_status TEXT NOT NULL, -- OFFLINE | NEW_TRANSLATIONS | REUSE_TRANSLATIONS | NO_TRANSLATIONS
    discord_message_id TEXT,
    CHECK (eval_status IN ('OFFLINE', 'NEW_TRANSLATIONS', 'REUSE_TRANSLATIONS', 'NO_TRANSLATIONS'))
);

-- translations table (cached username translations)
CREATE TABLE translations (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    translation TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- translation_to_evals junction table
CREATE TABLE translation_to_evals (
    translation_id BIGINT NOT NULL REFERENCES translations(id) ON DELETE CASCADE,
    eval_id BIGINT NOT NULL REFERENCES evals(id) ON DELETE CASCADE,
    PRIMARY KEY (translation_id, eval_id)
);

-- feedback table
CREATE TABLE feedback (
    id BIGSERIAL PRIMARY KEY,
    discord_message_id TEXT NOT NULL,
    feedback_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_evals_subscription_id ON evals(subscription_id);
CREATE INDEX idx_evals_evaluated_at ON evals(evaluated_at);
CREATE INDEX idx_translation_to_evals_eval ON translation_to_evals(eval_id);
