-- name: CreateSubscription :one
INSERT INTO subscriptions (discord_channel_id, lol_username, region, server_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (discord_channel_id, lol_username, region) DO NOTHING
RETURNING *;

-- name: GetAllSubscriptions :many
SELECT * FROM subscriptions
ORDER BY created_at DESC;

-- name: GetEvalByGameAndSubscription :one
SELECT * FROM evals
WHERE game_id = $1 AND subscription_id = $2
LIMIT 1;

-- name: DeleteEvals :execrows
DELETE FROM evals
WHERE evaluated_at < $1;

-- name: FindSubscriptionsWithExpiredNewestOnlineEval :many
SELECT subscription_id, MAX(evaluated_at) as newest_online_eval
FROM evals
WHERE eval_status != 'OFFLINE'
GROUP BY subscription_id
HAVING MAX(evaluated_at) < $1;

-- name: GetSubscriptionsByChannel :many
SELECT * FROM subscriptions
WHERE discord_channel_id = $1
ORDER BY created_at DESC;

-- name: CountSubscriptionsByServer :one
SELECT COUNT(*)
FROM subscriptions
WHERE server_id = $1;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1;

-- name: DeleteSubscription :execrows
DELETE FROM subscriptions
WHERE discord_channel_id = $1 AND lol_username = $2 AND region = $3;

-- name: DeleteSubscriptions :execrows
DELETE FROM subscriptions
WHERE id=ANY($1::bigint[]);

-- name: DeleteSubscriptionsByServer :execrows
DELETE FROM subscriptions
WHERE server_id = $1;

-- name: UpdateSubscriptionLastEvaluatedAt :exec
UPDATE subscriptions
SET last_evaluated_at = NOW()
WHERE id = $1;

-- name: CreateTranslation :one
INSERT INTO translations (username, translation, provider, model)
VALUES ($1, $2, $3, $4)
ON CONFLICT (username) DO UPDATE SET translation = $2, provider = $3, model = $4
RETURNING *;

-- name: GetTranslation :one
SELECT * FROM translations
WHERE username = $1;

-- name: GetTranslations :many
SELECT * FROM translations
WHERE username = ANY($1::text[]);

-- name: CreateEval :one
INSERT INTO evals (subscription_id, eval_status, discord_message_id, game_id)
VALUES ($1, $2, $3, $4)
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

-- name: DeleteOldTranslations :execrows
DELETE FROM translations WHERE created_at < $1;

-- name: DeleteOldFeedback :execrows
DELETE FROM feedback WHERE created_at < $1;

-- ===========================================
-- Companion Website Queries
-- ===========================================

-- Player queries

-- name: UpsertPlayer :one
INSERT INTO players (username, region, rank, top_champions, puuid)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (username) DO UPDATE SET
    region = EXCLUDED.region,
    rank = COALESCE(EXCLUDED.rank, players.rank),
    top_champions = COALESCE(EXCLUDED.top_champions, players.top_champions),
    puuid = COALESCE(EXCLUDED.puuid, players.puuid),
    last_updated = NOW()
RETURNING *;

-- name: GetPlayer :one
SELECT * FROM players WHERE username = $1;

-- name: ListAllPlayers :many
SELECT * FROM players ORDER BY username;

-- name: UpdatePlayerStats :exec
UPDATE players SET rank = $2, top_champions = $3, last_updated = NOW()
WHERE username = $1;

-- Public translation queries (JOIN against players for region/rank/top_champions)

-- name: UpsertPublicTranslation :one
INSERT INTO public_translations (username, translation, explanation, language, player_username, source_bot_id, riot_verified)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (username) DO UPDATE SET
    translation = EXCLUDED.translation,
    explanation = EXCLUDED.explanation,
    language = EXCLUDED.language,
    player_username = EXCLUDED.player_username,
    source_bot_id = EXCLUDED.source_bot_id,
    riot_verified = EXCLUDED.riot_verified
RETURNING *;

-- name: GetPublicTranslation :one
SELECT pt.id, pt.username, pt.translation, pt.explanation, pt.language,
       p.region, pt.source_bot_id, pt.riot_verified, p.rank, p.top_champions,
       pt.upvotes, pt.downvotes, pt.created_at, p.first_seen
FROM public_translations pt
JOIN players p ON pt.player_username = p.username
WHERE pt.id = $1;

-- name: GetPublicTranslationByUsername :one
SELECT pt.id, pt.username, pt.translation, pt.explanation, pt.language,
       p.region, pt.source_bot_id, pt.riot_verified, p.rank, p.top_champions,
       pt.upvotes, pt.downvotes, pt.created_at, p.first_seen
FROM public_translations pt
JOIN players p ON pt.player_username = p.username
WHERE pt.username = $1;

-- name: ListPublicTranslationsNew :many
SELECT pt.id, pt.username, pt.translation, pt.explanation, pt.language,
       p.region, pt.source_bot_id, pt.riot_verified, p.rank, p.top_champions,
       pt.upvotes, pt.downvotes, pt.created_at, p.first_seen
FROM public_translations pt
JOIN players p ON pt.player_username = p.username
WHERE ($1::text = '' OR p.region = $1)
  AND ($2::text = '' OR pt.language = $2)
ORDER BY pt.created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListPublicTranslationsTop :many
SELECT pt.id, pt.username, pt.translation, pt.explanation, pt.language,
       p.region, pt.source_bot_id, pt.riot_verified, p.rank, p.top_champions,
       pt.upvotes, pt.downvotes, pt.created_at, p.first_seen
FROM public_translations pt
JOIN players p ON pt.player_username = p.username
WHERE ($1::text = '' OR p.region = $1)
  AND ($2::text = '' OR pt.language = $2)
  AND pt.created_at > $5
ORDER BY (pt.upvotes - pt.downvotes) DESC, pt.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountPublicTranslations :one
SELECT COUNT(*)
FROM public_translations pt
JOIN players p ON pt.player_username = p.username
WHERE ($1::text = '' OR p.region = $1)
  AND ($2::text = '' OR pt.language = $2);

-- name: IncrementUpvotes :exec
UPDATE public_translations SET upvotes = upvotes + 1 WHERE id = $1;

-- name: DecrementUpvotes :exec
UPDATE public_translations SET upvotes = upvotes - 1 WHERE id = $1;

-- name: IncrementDownvotes :exec
UPDATE public_translations SET downvotes = downvotes + 1 WHERE id = $1;

-- name: DecrementDownvotes :exec
UPDATE public_translations SET downvotes = downvotes - 1 WHERE id = $1;

-- name: UpsertVote :one
INSERT INTO votes (translation_id, ip_hash, visitor_id, vote)
VALUES ($1, $2, $3, $4)
ON CONFLICT (translation_id, visitor_id) DO UPDATE SET vote = $4
RETURNING *;

-- name: GetVote :one
SELECT * FROM votes WHERE translation_id = $1 AND visitor_id = $2;

-- name: DeleteVote :execrows
DELETE FROM votes WHERE translation_id = $1 AND visitor_id = $2;

-- name: CountVotesByIP :one
SELECT COUNT(*) FROM votes WHERE ip_hash = $1;

-- name: CreatePublicFeedback :one
INSERT INTO public_feedback (translation_id, ip_hash, feedback_text)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListPublicFeedback :many
SELECT pf.*, pt.username, pt.translation
FROM public_feedback pf
JOIN public_translations pt ON pt.id = pf.translation_id
ORDER BY pf.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPublicFeedback :one
SELECT COUNT(*) FROM public_feedback;
