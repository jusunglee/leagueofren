ALTER TABLE subscriptions ADD COLUMN last_evaluated_at TIMESTAMPTZ;

CREATE INDEX idx_subscriptions_last_evaluated_at ON subscriptions(last_evaluated_at);
