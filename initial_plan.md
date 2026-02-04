Simple discord bot that can /subscribe <league_of_legends_username> a user and translates any (for now, more to come) korean and english summoner names in the game when subscribed user starts one. It uses a simple haiku LLM with a prefixed prompt and sends a message to the subscribed discord channel.

General design:

## Data access patterns (not db write-heavy discord-write moderate, db read heavy)

1. Insert subscription: Insert discord_channel_id, lol_username, other_discord_pid_maybe_server_id
   1.5: Ignores: subscription_id, username,
2. get all subscriptions
3. delete subscription
4. evals: id, subscription_id, evaluated_at, eval_status[OFFLINE, NEW_TRANSLATIONS, REUSE_TRANSLATIONS, NO_TRANSLATIONS], discord_message_id
5. translation: username, translation
6. translation_to_evals: translation_id, eval_id
7. Feedback: discord_message_id, text

## Data flow

1. Insert / delete subscriptions is trivial
2. Get all subscriptions, for each subscription
3. use riot api to look up if in game
4. if not, insert OFFLINE eval and continue
5. get all korean/chinese character usernames in game
6. Grab any existing translations and append to list to message
7. For remaining usernames, invoke LLM to translate with prompt and store in translation table
8. send concatenated translation message to subscription
9. Store evals and translation_to_evals
10. continue to next subscription

## other implementation details:

1. Use exponential backoff for ratelimiting (20qps/100qp2min for riot), (50 rps for discord)
2. start with 2 minute poll intervals
3. Use grafana / loki for logs and metrics and tracing, make sure to measure everything important like total duration, per sub duration, api call distribution latency, status codes, etc
4. cleanup job for evals older than 2 weeks

## bonus ideas

need to manage RL very nicely we have potentially a lot of lol usernames to look up. maybe use user's discord status playing lol if self-subscribed?

## Stack

Golang, psql, discord api, league api. deployed on railway

## name inspo

i saw a lot of 인 (in) and 人 (ren) in my games, which I knew is korean and chinese for "person", but I could almost never figure out the preceding2 characters so I always looked them up. this inspired leagueofren
