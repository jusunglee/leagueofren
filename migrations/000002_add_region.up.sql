-- Add region column to subscriptions
ALTER TABLE subscriptions ADD COLUMN region TEXT NOT NULL DEFAULT 'NA';

-- Remove the default after adding (force explicit region on new inserts)
ALTER TABLE subscriptions ALTER COLUMN region DROP DEFAULT;

-- Drop old unique constraint and create new one with region
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_discord_channel_id_lol_username_key;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_channel_username_region_key
    UNIQUE (discord_channel_id, lol_username, region);
