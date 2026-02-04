-- Remove region-based unique constraint
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_channel_username_region_key;

-- Add back original unique constraint
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_discord_channel_id_lol_username_key
    UNIQUE (discord_channel_id, lol_username);

-- Drop region column
ALTER TABLE subscriptions DROP COLUMN IF EXISTS region;
