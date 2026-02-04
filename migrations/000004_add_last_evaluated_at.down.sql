DROP INDEX IF EXISTS idx_subscriptions_last_evaluated_at;

ALTER TABLE subscriptions DROP COLUMN IF EXISTS last_evaluated_at;
