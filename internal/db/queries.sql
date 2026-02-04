-- name: CreateSubscription :one
INSERT INTO subscriptions (discord_channel_id, lol_username, region)
VALUES ($1, $2, $3)
ON CONFLICT (discord_channel_id, lol_username, region) DO NOTHING
RETURNING *;

-- name: GetAllSubscriptions :many
SELECT * FROM subscriptions
ORDER BY created_at DESC
LIMIT $1;

-- name: GetEvalByGameAndSubscription :one
SELECT * FROM evals
WHERE game_id = $1 AND subscription_id = $2
LIMIT 1;

-- name: GetSubscriptionsByChannel :many
SELECT * FROM subscriptions
WHERE discord_channel_id = $1
ORDER BY created_at DESC;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1;

-- name: DeleteSubscription :execrows
DELETE FROM subscriptions
WHERE discord_channel_id = $1 AND lol_username = $2 AND region = $3;

-- name: UpdateSubscriptionLastEvaluatedAt :exec
UPDATE subscriptions
SET last_evaluated_at = NOW()
WHERE id = $1;

-- name: CreateTranslation :one
INSERT INTO translations (username, translation)
VALUES ($1, $2)
ON CONFLICT (username) DO UPDATE SET translation = $2
RETURNING *;

-- name: GetTranslation :one
SELECT * FROM translations
WHERE username = $1;

-- name: GetTranslations :many
SELECT * FROM translations
WHERE username = ANY($1::text[]);

-- name: CreateEval :one
INSERT INTO evals (subscription_id, eval_status, discord_message_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLatestEvalForSubscription :one
SELECT * FROM evals
WHERE subscription_id = $1
ORDER BY evaluated_at DESC
LIMIT 1;

-- name: CreateTranslationToEval :exec
INSERT INTO translation_to_evals (translation_id, eval_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetTranslationsForEval :many
SELECT t.*
FROM translations t
JOIN translation_to_evals tte ON t.id = tte.translation_id
WHERE tte.eval_id = $1;

-- name: CreateFeedback :one
INSERT INTO feedback (discord_message_id, feedback_text)
VALUES ($1, $2)
RETURNING *;

-- Account cache queries
-- name: GetCachedAccount :one
SELECT game_name, tag_line, region, puuid
FROM riot_account_cache
WHERE game_name = $1 AND tag_line = $2 AND region = $3 AND expires_at > NOW();

-- name: CacheAccount :exec
INSERT INTO riot_account_cache (game_name, tag_line, region, puuid, expires_at)
VALUES ($1, $2, $3, $4, NOW() + interval '24 hours')
ON CONFLICT (game_name, tag_line, region)
DO UPDATE SET puuid = $4, cached_at = NOW(), expires_at = NOW() + interval '24 hours';

-- Game cache queries
-- name: GetCachedGameStatus :one
SELECT puuid, region, in_game, game_id, participants
FROM riot_game_cache
WHERE puuid = $1 AND region = $2 AND expires_at > NOW();

-- name: CacheGameStatus :exec
INSERT INTO riot_game_cache (puuid, region, in_game, game_id, participants, expires_at)
VALUES ($1, $2, $3, $4, $5, NOW() + interval '2 minutes')
ON CONFLICT (puuid, region)
DO UPDATE SET in_game = $3, game_id = $4, participants = $5, cached_at = NOW(), expires_at = NOW() + interval '2 minutes';

-- Cleanup queries
-- name: DeleteExpiredAccountCache :exec
DELETE FROM riot_account_cache WHERE expires_at < NOW();

-- name: DeleteExpiredGameCache :exec
DELETE FROM riot_game_cache WHERE expires_at < NOW();
