-- name: CreateSubscription :one
INSERT INTO subscriptions (discord_channel_id, lol_username)
VALUES ($1, $2)
ON CONFLICT (discord_channel_id, lol_username) DO NOTHING
RETURNING *;

-- name: GetAllSubscriptions :many
SELECT * FROM subscriptions
ORDER BY created_at DESC;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE discord_channel_id = $1 AND lol_username = $2;

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
