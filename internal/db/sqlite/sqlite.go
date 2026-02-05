package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jusunglee/leagueofren/internal/db"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Repository implements db.Repository using SQLite
type Repository struct {
	db *sql.DB
}

// New creates a new SQLite repository
func New(ctx context.Context, dbPath string) (*Repository, error) {
	// Strip sqlite:// prefix if present
	dbPath = strings.TrimPrefix(dbPath, "sqlite://")

	isNew := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		isNew = true
	}

	sqliteDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening SQLite database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := sqliteDB.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		sqliteDB.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := sqliteDB.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		sqliteDB.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	repo := &Repository{db: sqliteDB}

	if isNew {
		if _, err := sqliteDB.ExecContext(ctx, schemaSQL); err != nil {
			sqliteDB.Close()
			return nil, fmt.Errorf("initializing schema: %w", err)
		}
		slog.Info("created new SQLite database", "path", dbPath)
	}

	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

// Subscription methods

func (r *Repository) CreateSubscription(ctx context.Context, arg db.CreateSubscriptionParams) (db.Subscription, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO subscriptions (discord_channel_id, lol_username, region, server_id)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (discord_channel_id, lol_username, region) DO NOTHING
	`, arg.DiscordChannelID, arg.LolUsername, arg.Region, arg.ServerID)
	if err != nil {
		return db.Subscription{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return db.Subscription{}, err
	}
	if rowsAffected == 0 {
		return db.Subscription{}, db.ErrNoRows
	}

	id, err := result.LastInsertId()
	if err != nil {
		return db.Subscription{}, err
	}

	return r.GetSubscriptionByID(ctx, id)
}

func (r *Repository) GetAllSubscriptions(ctx context.Context, limit int32) ([]db.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, discord_channel_id, server_id, lol_username, region, created_at, last_evaluated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSubscriptions(rows)
}

func (r *Repository) GetSubscriptionsByChannel(ctx context.Context, discordChannelID string) ([]db.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, discord_channel_id, server_id, lol_username, region, created_at, last_evaluated_at
		FROM subscriptions
		WHERE discord_channel_id = ?
		ORDER BY created_at DESC
	`, discordChannelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSubscriptions(rows)
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id int64) (db.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, discord_channel_id, server_id, lol_username, region, created_at, last_evaluated_at
		FROM subscriptions
		WHERE id = ?
	`, id)

	return scanSubscription(row)
}

func (r *Repository) CountSubscriptionsByServer(ctx context.Context, serverID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM subscriptions WHERE server_id = ?
	`, serverID).Scan(&count)
	return count, err
}

func (r *Repository) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM subscriptions
		WHERE discord_channel_id = ? AND lol_username = ? AND region = ?
	`, arg.DiscordChannelID, arg.LolUsername, arg.Region)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) DeleteSubscriptions(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM subscriptions WHERE id IN (%s)", strings.Join(placeholders, ","))
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) UpdateSubscriptionLastEvaluatedAt(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions SET last_evaluated_at = datetime('now') WHERE id = ?
	`, id)
	return err
}

// Eval methods

func (r *Repository) CreateEval(ctx context.Context, arg db.CreateEvalParams) (db.Eval, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO evals (subscription_id, eval_status, discord_message_id, game_id)
		VALUES (?, ?, ?, ?)
	`, arg.SubscriptionID, arg.EvalStatus, nullString(arg.DiscordMessageID), nullInt64(arg.GameID))
	if err != nil {
		return db.Eval{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return db.Eval{}, err
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, subscription_id, game_id, evaluated_at, eval_status, discord_message_id
		FROM evals WHERE id = ?
	`, id)

	return scanEval(row)
}

func (r *Repository) GetEvalByGameAndSubscription(ctx context.Context, arg db.GetEvalByGameAndSubscriptionParams) (db.Eval, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, subscription_id, game_id, evaluated_at, eval_status, discord_message_id
		FROM evals
		WHERE game_id = ? AND subscription_id = ?
		LIMIT 1
	`, nullInt64(arg.GameID), arg.SubscriptionID)

	return scanEval(row)
}

func (r *Repository) GetLatestEvalForSubscription(ctx context.Context, subscriptionID int64) (db.Eval, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, subscription_id, game_id, evaluated_at, eval_status, discord_message_id
		FROM evals
		WHERE subscription_id = ?
		ORDER BY evaluated_at DESC
		LIMIT 1
	`, subscriptionID)

	return scanEval(row)
}

func (r *Repository) DeleteEvals(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM evals WHERE evaluated_at < ?
	`, before.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) FindSubscriptionsWithExpiredNewestOnlineEval(ctx context.Context, before time.Time) ([]db.FindSubscriptionsWithExpiredNewestOnlineEvalRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT subscription_id, MAX(evaluated_at) as newest_online_eval
		FROM evals
		WHERE eval_status != 'OFFLINE'
		GROUP BY subscription_id
		HAVING MAX(evaluated_at) < ?
	`, before.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []db.FindSubscriptionsWithExpiredNewestOnlineEvalRow
	for rows.Next() {
		var row db.FindSubscriptionsWithExpiredNewestOnlineEvalRow
		var newestEvalStr string
		if err := rows.Scan(&row.SubscriptionID, &newestEvalStr); err != nil {
			return nil, err
		}
		row.NewestOnlineEval, _ = time.Parse(time.RFC3339, newestEvalStr)
		results = append(results, row)
	}
	return results, rows.Err()
}

// Translation methods

func (r *Repository) CreateTranslation(ctx context.Context, arg db.CreateTranslationParams) (db.Translation, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO translations (username, translation, provider, model)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (username) DO UPDATE SET translation = ?, provider = ?, model = ?
	`, arg.Username, arg.Translation, arg.Provider, arg.Model, arg.Translation, arg.Provider, arg.Model)
	if err != nil {
		return db.Translation{}, err
	}

	return r.GetTranslation(ctx, arg.Username)
}

func (r *Repository) GetTranslation(ctx context.Context, username string) (db.Translation, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, translation, provider, model, created_at
		FROM translations WHERE username = ?
	`, username)

	return scanTranslation(row)
}

func (r *Repository) GetTranslations(ctx context.Context, usernames []string) ([]db.Translation, error) {
	if len(usernames) == 0 {
		return []db.Translation{}, nil
	}

	placeholders := make([]string, len(usernames))
	args := make([]interface{}, len(usernames))
	for i, u := range usernames {
		placeholders[i] = "?"
		args[i] = u
	}

	query := fmt.Sprintf(`
		SELECT id, username, translation, provider, model, created_at
		FROM translations WHERE username IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTranslations(rows)
}

func (r *Repository) GetTranslationsForEval(ctx context.Context, evalID int64) ([]db.Translation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.username, t.translation, t.provider, t.model, t.created_at
		FROM translations t
		JOIN translation_to_evals tte ON t.id = tte.translation_id
		WHERE tte.eval_id = ?
	`, evalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTranslations(rows)
}

func (r *Repository) CreateTranslationToEval(ctx context.Context, arg db.CreateTranslationToEvalParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO translation_to_evals (translation_id, eval_id)
		VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`, arg.TranslationID, arg.EvalID)
	return err
}

// Feedback methods

func (r *Repository) CreateFeedback(ctx context.Context, arg db.CreateFeedbackParams) (db.Feedback, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback (discord_message_id, feedback_text)
		VALUES (?, ?)
	`, arg.DiscordMessageID, arg.FeedbackText)
	if err != nil {
		return db.Feedback{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return db.Feedback{}, err
	}

	var f db.Feedback
	var createdAtStr string
	err = r.db.QueryRowContext(ctx, `
		SELECT id, discord_message_id, feedback_text, created_at FROM feedback WHERE id = ?
	`, id).Scan(&f.ID, &f.DiscordMessageID, &f.FeedbackText, &createdAtStr)
	if err != nil {
		return db.Feedback{}, err
	}
	f.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return f, nil
}

// Cache methods

func (r *Repository) GetCachedAccount(ctx context.Context, arg db.GetCachedAccountParams) (db.GetCachedAccountRow, error) {
	var row db.GetCachedAccountRow
	err := r.db.QueryRowContext(ctx, `
		SELECT game_name, tag_line, region, puuid
		FROM riot_account_cache
		WHERE game_name = ? AND tag_line = ? AND region = ? AND expires_at > datetime('now')
	`, arg.GameName, arg.TagLine, arg.Region).Scan(&row.GameName, &row.TagLine, &row.Region, &row.Puuid)
	if err == sql.ErrNoRows {
		return db.GetCachedAccountRow{}, db.ErrNoRows
	}
	return row, err
}

func (r *Repository) CacheAccount(ctx context.Context, arg db.CacheAccountParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO riot_account_cache (game_name, tag_line, region, puuid, expires_at)
		VALUES (?, ?, ?, ?, datetime('now', '+24 hours'))
		ON CONFLICT (game_name, tag_line, region)
		DO UPDATE SET puuid = ?, cached_at = datetime('now'), expires_at = datetime('now', '+24 hours')
	`, arg.GameName, arg.TagLine, arg.Region, arg.Puuid, arg.Puuid)
	return err
}

func (r *Repository) GetCachedGameStatus(ctx context.Context, arg db.GetCachedGameStatusParams) (db.GetCachedGameStatusRow, error) {
	var row db.GetCachedGameStatusRow
	var inGame int
	err := r.db.QueryRowContext(ctx, `
		SELECT puuid, region, in_game, game_id, participants
		FROM riot_game_cache
		WHERE puuid = ? AND region = ? AND expires_at > datetime('now')
	`, arg.Puuid, arg.Region).Scan(&row.Puuid, &row.Region, &inGame, &row.GameID, &row.Participants)
	if err == sql.ErrNoRows {
		return db.GetCachedGameStatusRow{}, db.ErrNoRows
	}
	row.InGame = inGame != 0
	return row, err
}

func (r *Repository) CacheGameStatus(ctx context.Context, arg db.CacheGameStatusParams) error {
	inGame := 0
	if arg.InGame {
		inGame = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO riot_game_cache (puuid, region, in_game, game_id, participants, expires_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', '+2 minutes'))
		ON CONFLICT (puuid, region)
		DO UPDATE SET in_game = ?, game_id = ?, participants = ?, cached_at = datetime('now'), expires_at = datetime('now', '+2 minutes')
	`, arg.Puuid, arg.Region, inGame, nullInt64(arg.GameID), arg.Participants, inGame, nullInt64(arg.GameID), arg.Participants)
	return err
}

func (r *Repository) DeleteExpiredAccountCache(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM riot_account_cache WHERE expires_at < datetime('now')`)
	return err
}

func (r *Repository) DeleteExpiredGameCache(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM riot_game_cache WHERE expires_at < datetime('now')`)
	return err
}

// Helper functions

func scanSubscription(row *sql.Row) (db.Subscription, error) {
	var s db.Subscription
	var createdAtStr, lastEvaluatedAtStr string
	err := row.Scan(&s.ID, &s.DiscordChannelID, &s.ServerID, &s.LolUsername, &s.Region, &createdAtStr, &lastEvaluatedAtStr)
	if err == sql.ErrNoRows {
		return db.Subscription{}, db.ErrNoRows
	}
	if err != nil {
		return db.Subscription{}, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	s.LastEvaluatedAt, _ = time.Parse(time.RFC3339, lastEvaluatedAtStr)
	return s, nil
}

func scanSubscriptions(rows *sql.Rows) ([]db.Subscription, error) {
	var subs []db.Subscription
	for rows.Next() {
		var s db.Subscription
		var createdAtStr, lastEvaluatedAtStr string
		if err := rows.Scan(&s.ID, &s.DiscordChannelID, &s.ServerID, &s.LolUsername, &s.Region, &createdAtStr, &lastEvaluatedAtStr); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		s.LastEvaluatedAt, _ = time.Parse(time.RFC3339, lastEvaluatedAtStr)
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func scanEval(row *sql.Row) (db.Eval, error) {
	var e db.Eval
	var evaluatedAtStr string
	err := row.Scan(&e.ID, &e.SubscriptionID, &e.GameID, &evaluatedAtStr, &e.EvalStatus, &e.DiscordMessageID)
	if err == sql.ErrNoRows {
		return db.Eval{}, db.ErrNoRows
	}
	if err != nil {
		return db.Eval{}, err
	}
	e.EvaluatedAt, _ = time.Parse(time.RFC3339, evaluatedAtStr)
	return e, nil
}

func scanTranslation(row *sql.Row) (db.Translation, error) {
	var t db.Translation
	var createdAtStr string
	err := row.Scan(&t.ID, &t.Username, &t.Translation, &t.Provider, &t.Model, &createdAtStr)
	if err == sql.ErrNoRows {
		return db.Translation{}, db.ErrNoRows
	}
	if err != nil {
		return db.Translation{}, err
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return t, nil
}

func scanTranslations(rows *sql.Rows) ([]db.Translation, error) {
	var translations []db.Translation
	for rows.Next() {
		var t db.Translation
		var createdAtStr string
		if err := rows.Scan(&t.ID, &t.Username, &t.Translation, &t.Provider, &t.Model, &createdAtStr); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		translations = append(translations, t)
	}
	return translations, rows.Err()
}

func nullString(s sql.NullString) interface{} {
	if s.Valid {
		return s.String
	}
	return nil
}

func nullInt64(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
}
