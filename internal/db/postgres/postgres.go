package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/db/sqlc"
)

// Repository implements db.Repository using PostgreSQL via pgx
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// New creates a new PostgreSQL repository
func New(ctx context.Context, databaseURL string) (*Repository, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	config.MaxConns = 5
	config.MinConns = 2
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 30 * time.Second
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Repository{
		pool:    pool,
		queries: sqlc.New(pool),
	}, nil
}

func (r *Repository) Close() error {
	r.pool.Close()
	return nil
}

func (r *Repository) WithTx(ctx context.Context, fn func(repo db.Repository) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	// If fn() panics, the normal err-check rollback below won't run.
	// recover() catches the panic so we can roll back the tx (releasing the db connection), then re-panic.
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
	}()

	txRepo := &Repository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}

	err = fn(txRepo)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// Subscription methods

func (r *Repository) CreateSubscription(ctx context.Context, arg db.CreateSubscriptionParams) (db.Subscription, error) {
	result, err := r.queries.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
		DiscordChannelID: arg.DiscordChannelID,
		LolUsername:      arg.LolUsername,
		Region:           arg.Region,
		ServerID:         arg.ServerID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Subscription{}, db.ErrNoRows
		}
		return db.Subscription{}, err
	}
	return convertSubscription(result), nil
}

func (r *Repository) GetAllSubscriptions(ctx context.Context, limit int32) ([]db.Subscription, error) {
	results, err := r.queries.GetAllSubscriptions(ctx, limit)
	if err != nil {
		return nil, err
	}
	return convertSubscriptions(results), nil
}

func (r *Repository) GetSubscriptionsByChannel(ctx context.Context, discordChannelID string) ([]db.Subscription, error) {
	results, err := r.queries.GetSubscriptionsByChannel(ctx, discordChannelID)
	if err != nil {
		return nil, err
	}
	return convertSubscriptions(results), nil
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id int64) (db.Subscription, error) {
	result, err := r.queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Subscription{}, db.ErrNoRows
		}
		return db.Subscription{}, err
	}
	return convertSubscription(result), nil
}

func (r *Repository) CountSubscriptionsByServer(ctx context.Context, serverID string) (int64, error) {
	return r.queries.CountSubscriptionsByServer(ctx, serverID)
}

func (r *Repository) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) (int64, error) {
	return r.queries.DeleteSubscription(ctx, sqlc.DeleteSubscriptionParams{
		DiscordChannelID: arg.DiscordChannelID,
		LolUsername:      arg.LolUsername,
		Region:           arg.Region,
	})
}

func (r *Repository) DeleteSubscriptions(ctx context.Context, ids []int64) (int64, error) {
	return r.queries.DeleteSubscriptions(ctx, ids)
}

func (r *Repository) UpdateSubscriptionLastEvaluatedAt(ctx context.Context, id int64) error {
	return r.queries.UpdateSubscriptionLastEvaluatedAt(ctx, id)
}

// Eval methods

func (r *Repository) CreateEval(ctx context.Context, arg db.CreateEvalParams) (db.Eval, error) {
	result, err := r.queries.CreateEval(ctx, sqlc.CreateEvalParams{
		SubscriptionID:   arg.SubscriptionID,
		EvalStatus:       arg.EvalStatus,
		DiscordMessageID: toPgText(arg.DiscordMessageID),
		GameID:           toPgInt8(arg.GameID),
	})
	if err != nil {
		return db.Eval{}, err
	}
	return convertEval(result), nil
}

func (r *Repository) GetEvalByGameAndSubscription(ctx context.Context, arg db.GetEvalByGameAndSubscriptionParams) (db.Eval, error) {
	result, err := r.queries.GetEvalByGameAndSubscription(ctx, sqlc.GetEvalByGameAndSubscriptionParams{
		GameID:         toPgInt8(arg.GameID),
		SubscriptionID: arg.SubscriptionID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Eval{}, db.ErrNoRows
		}
		return db.Eval{}, err
	}
	return convertEval(result), nil
}

func (r *Repository) GetLatestEvalForSubscription(ctx context.Context, subscriptionID int64) (db.Eval, error) {
	result, err := r.queries.GetLatestEvalForSubscription(ctx, subscriptionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Eval{}, db.ErrNoRows
		}
		return db.Eval{}, err
	}
	return convertEval(result), nil
}

func (r *Repository) DeleteEvals(ctx context.Context, before time.Time) (int64, error) {
	return r.queries.DeleteEvals(ctx, pgtype.Timestamptz{Valid: true, Time: before})
}

func (r *Repository) FindSubscriptionsWithExpiredNewestOnlineEval(ctx context.Context, before time.Time) ([]db.FindSubscriptionsWithExpiredNewestOnlineEvalRow, error) {
	results, err := r.queries.FindSubscriptionsWithExpiredNewestOnlineEval(ctx, pgtype.Timestamptz{Valid: true, Time: before})
	if err != nil {
		return nil, err
	}
	rows := make([]db.FindSubscriptionsWithExpiredNewestOnlineEvalRow, len(results))
	for i, result := range results {
		var newestEval time.Time
		if t, ok := result.NewestOnlineEval.(time.Time); ok {
			newestEval = t
		}
		rows[i] = db.FindSubscriptionsWithExpiredNewestOnlineEvalRow{
			SubscriptionID:   result.SubscriptionID,
			NewestOnlineEval: newestEval,
		}
	}
	return rows, nil
}

// Translation methods

func (r *Repository) CreateTranslation(ctx context.Context, arg db.CreateTranslationParams) (db.Translation, error) {
	result, err := r.queries.CreateTranslation(ctx, sqlc.CreateTranslationParams{
		Username:    arg.Username,
		Translation: arg.Translation,
		Provider:    arg.Provider,
		Model:       arg.Model,
	})
	if err != nil {
		return db.Translation{}, err
	}
	return convertTranslation(result), nil
}

func (r *Repository) GetTranslation(ctx context.Context, username string) (db.Translation, error) {
	result, err := r.queries.GetTranslation(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Translation{}, db.ErrNoRows
		}
		return db.Translation{}, err
	}
	return convertTranslation(result), nil
}

func (r *Repository) GetTranslations(ctx context.Context, usernames []string) ([]db.Translation, error) {
	results, err := r.queries.GetTranslations(ctx, usernames)
	if err != nil {
		return nil, err
	}
	return convertTranslations(results), nil
}

func (r *Repository) GetTranslationsForEval(ctx context.Context, evalID int64) ([]db.Translation, error) {
	results, err := r.queries.GetTranslationsForEval(ctx, evalID)
	if err != nil {
		return nil, err
	}
	return convertTranslations(results), nil
}

func (r *Repository) CreateTranslationToEval(ctx context.Context, arg db.CreateTranslationToEvalParams) error {
	return r.queries.CreateTranslationToEval(ctx, sqlc.CreateTranslationToEvalParams{
		TranslationID: arg.TranslationID,
		EvalID:        arg.EvalID,
	})
}

// Feedback methods

func (r *Repository) CreateFeedback(ctx context.Context, arg db.CreateFeedbackParams) (db.Feedback, error) {
	result, err := r.queries.CreateFeedback(ctx, sqlc.CreateFeedbackParams{
		DiscordMessageID: arg.DiscordMessageID,
		FeedbackText:     arg.FeedbackText,
	})
	if err != nil {
		return db.Feedback{}, err
	}
	return db.Feedback{
		ID:               result.ID,
		DiscordMessageID: result.DiscordMessageID,
		FeedbackText:     result.FeedbackText,
		CreatedAt:        result.CreatedAt.Time,
	}, nil
}

// Cache methods

func (r *Repository) GetCachedAccount(ctx context.Context, arg db.GetCachedAccountParams) (db.GetCachedAccountRow, error) {
	result, err := r.queries.GetCachedAccount(ctx, sqlc.GetCachedAccountParams{
		GameName: arg.GameName,
		TagLine:  arg.TagLine,
		Region:   arg.Region,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.GetCachedAccountRow{}, db.ErrNoRows
		}
		return db.GetCachedAccountRow{}, err
	}
	return db.GetCachedAccountRow{
		GameName: result.GameName,
		TagLine:  result.TagLine,
		Region:   result.Region,
		Puuid:    result.Puuid,
	}, nil
}

func (r *Repository) CacheAccount(ctx context.Context, arg db.CacheAccountParams) error {
	return r.queries.CacheAccount(ctx, sqlc.CacheAccountParams{
		GameName: arg.GameName,
		TagLine:  arg.TagLine,
		Region:   arg.Region,
		Puuid:    arg.Puuid,
	})
}

func (r *Repository) GetCachedGameStatus(ctx context.Context, arg db.GetCachedGameStatusParams) (db.GetCachedGameStatusRow, error) {
	result, err := r.queries.GetCachedGameStatus(ctx, sqlc.GetCachedGameStatusParams{
		Puuid:  arg.Puuid,
		Region: arg.Region,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.GetCachedGameStatusRow{}, db.ErrNoRows
		}
		return db.GetCachedGameStatusRow{}, err
	}
	return db.GetCachedGameStatusRow{
		Puuid:        result.Puuid,
		Region:       result.Region,
		InGame:       result.InGame,
		GameID:       fromPgInt8(result.GameID),
		Participants: result.Participants,
	}, nil
}

func (r *Repository) CacheGameStatus(ctx context.Context, arg db.CacheGameStatusParams) error {
	return r.queries.CacheGameStatus(ctx, sqlc.CacheGameStatusParams{
		Puuid:        arg.Puuid,
		Region:       arg.Region,
		InGame:       arg.InGame,
		GameID:       toPgInt8(arg.GameID),
		Participants: arg.Participants,
	})
}

func (r *Repository) DeleteOldTranslations(ctx context.Context, before time.Time) (int64, error) {
	return r.queries.DeleteOldTranslations(ctx, pgtype.Timestamptz{Valid: true, Time: before})
}

func (r *Repository) DeleteOldFeedback(ctx context.Context, before time.Time) (int64, error) {
	return r.queries.DeleteOldFeedback(ctx, pgtype.Timestamptz{Valid: true, Time: before})
}

func (r *Repository) DeleteExpiredAccountCache(ctx context.Context) error {
	return r.queries.DeleteExpiredAccountCache(ctx)
}

func (r *Repository) DeleteExpiredGameCache(ctx context.Context) error {
	return r.queries.DeleteExpiredGameCache(ctx)
}

// Public Translation methods

func (r *Repository) UpsertPublicTranslation(ctx context.Context, arg db.UpsertPublicTranslationParams) (db.PublicTranslation, error) {
	result, err := r.queries.UpsertPublicTranslation(ctx, sqlc.UpsertPublicTranslationParams{
		Username:     arg.Username,
		Translation:  arg.Translation,
		Explanation:  toPgText(arg.Explanation),
		Language:     arg.Language,
		Region:       arg.Region,
		SourceBotID:  toPgText(arg.SourceBotID),
		RiotVerified: arg.RiotVerified,
	})
	if err != nil {
		return db.PublicTranslation{}, err
	}
	return convertPublicTranslation(result), nil
}

func (r *Repository) GetPublicTranslation(ctx context.Context, id int64) (db.PublicTranslation, error) {
	result, err := r.queries.GetPublicTranslation(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.PublicTranslation{}, db.ErrNoRows
		}
		return db.PublicTranslation{}, err
	}
	return convertPublicTranslation(result), nil
}

func (r *Repository) GetPublicTranslationByUsername(ctx context.Context, username string) (db.PublicTranslation, error) {
	result, err := r.queries.GetPublicTranslationByUsername(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.PublicTranslation{}, db.ErrNoRows
		}
		return db.PublicTranslation{}, err
	}
	return convertPublicTranslation(result), nil
}

func (r *Repository) ListPublicTranslationsNew(ctx context.Context, arg db.ListPublicTranslationsNewParams) ([]db.PublicTranslation, error) {
	results, err := r.queries.ListPublicTranslationsNew(ctx, sqlc.ListPublicTranslationsNewParams{
		Column1: arg.Region,
		Column2: arg.Language,
		Limit:   arg.Limit,
		Offset:  arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	return convertPublicTranslations(results), nil
}

func (r *Repository) ListPublicTranslationsTop(ctx context.Context, arg db.ListPublicTranslationsTopParams) ([]db.PublicTranslation, error) {
	results, err := r.queries.ListPublicTranslationsTop(ctx, sqlc.ListPublicTranslationsTopParams{
		Column1:   arg.Region,
		Column2:   arg.Language,
		Limit:     arg.Limit,
		Offset:    arg.Offset,
		CreatedAt: pgtype.Timestamptz{Valid: true, Time: arg.CreatedAt},
	})
	if err != nil {
		return nil, err
	}
	return convertPublicTranslations(results), nil
}

func (r *Repository) CountPublicTranslations(ctx context.Context, arg db.CountPublicTranslationsParams) (int64, error) {
	return r.queries.CountPublicTranslations(ctx, sqlc.CountPublicTranslationsParams{
		Column1: arg.Region,
		Column2: arg.Language,
	})
}

func (r *Repository) IncrementUpvotes(ctx context.Context, id int64) error {
	return r.queries.IncrementUpvotes(ctx, id)
}

func (r *Repository) DecrementUpvotes(ctx context.Context, id int64) error {
	return r.queries.DecrementUpvotes(ctx, id)
}

func (r *Repository) IncrementDownvotes(ctx context.Context, id int64) error {
	return r.queries.IncrementDownvotes(ctx, id)
}

func (r *Repository) DecrementDownvotes(ctx context.Context, id int64) error {
	return r.queries.DecrementDownvotes(ctx, id)
}

// Vote methods

func (r *Repository) UpsertVote(ctx context.Context, arg db.UpsertVoteParams) (db.Vote, error) {
	result, err := r.queries.UpsertVote(ctx, sqlc.UpsertVoteParams{
		TranslationID: arg.TranslationID,
		IpHash:        arg.IpHash,
		Vote:          arg.Vote,
	})
	if err != nil {
		return db.Vote{}, err
	}
	return convertVote(result), nil
}

func (r *Repository) GetVote(ctx context.Context, arg db.GetVoteParams) (db.Vote, error) {
	result, err := r.queries.GetVote(ctx, sqlc.GetVoteParams{
		TranslationID: arg.TranslationID,
		IpHash:        arg.IpHash,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Vote{}, db.ErrNoRows
		}
		return db.Vote{}, err
	}
	return convertVote(result), nil
}

func (r *Repository) DeleteVote(ctx context.Context, arg db.DeleteVoteParams) (int64, error) {
	return r.queries.DeleteVote(ctx, sqlc.DeleteVoteParams{
		TranslationID: arg.TranslationID,
		IpHash:        arg.IpHash,
	})
}

// Public Feedback methods

func (r *Repository) CreatePublicFeedback(ctx context.Context, arg db.CreatePublicFeedbackParams) (db.PublicFeedback, error) {
	result, err := r.queries.CreatePublicFeedback(ctx, sqlc.CreatePublicFeedbackParams{
		TranslationID: arg.TranslationID,
		IpHash:        arg.IpHash,
		FeedbackText:  arg.FeedbackText,
	})
	if err != nil {
		return db.PublicFeedback{}, err
	}
	return db.PublicFeedback{
		ID:            result.ID,
		TranslationID: result.TranslationID,
		IpHash:        result.IpHash,
		FeedbackText:  result.FeedbackText,
		CreatedAt:     result.CreatedAt.Time,
	}, nil
}

func (r *Repository) ListPublicFeedback(ctx context.Context, arg db.ListPublicFeedbackParams) ([]db.ListPublicFeedbackRow, error) {
	results, err := r.queries.ListPublicFeedback(ctx, sqlc.ListPublicFeedbackParams{
		Limit:  arg.Limit,
		Offset: arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	rows := make([]db.ListPublicFeedbackRow, len(results))
	for i, r := range results {
		rows[i] = db.ListPublicFeedbackRow{
			ID:            r.ID,
			TranslationID: r.TranslationID,
			IpHash:        r.IpHash,
			FeedbackText:  r.FeedbackText,
			CreatedAt:     r.CreatedAt.Time,
			Username:      r.Username,
			Translation:   r.Translation,
		}
	}
	return rows, nil
}

func (r *Repository) CountPublicFeedback(ctx context.Context) (int64, error) {
	return r.queries.CountPublicFeedback(ctx)
}

// Type conversion helpers

func convertSubscription(s sqlc.Subscription) db.Subscription {
	return db.Subscription{
		ID:               s.ID,
		DiscordChannelID: s.DiscordChannelID,
		ServerID:         s.ServerID,
		LolUsername:      s.LolUsername,
		Region:           s.Region,
		CreatedAt:        s.CreatedAt.Time,
		LastEvaluatedAt:  s.LastEvaluatedAt.Time,
	}
}

func convertSubscriptions(subs []sqlc.Subscription) []db.Subscription {
	result := make([]db.Subscription, len(subs))
	for i, s := range subs {
		result[i] = convertSubscription(s)
	}
	return result
}

func convertEval(e sqlc.Eval) db.Eval {
	return db.Eval{
		ID:               e.ID,
		SubscriptionID:   e.SubscriptionID,
		GameID:           fromPgInt8(e.GameID),
		EvaluatedAt:      e.EvaluatedAt.Time,
		EvalStatus:       e.EvalStatus,
		DiscordMessageID: fromPgText(e.DiscordMessageID),
	}
}

func convertTranslation(t sqlc.Translation) db.Translation {
	return db.Translation{
		ID:          t.ID,
		Username:    t.Username,
		Translation: t.Translation,
		Provider:    t.Provider,
		Model:       t.Model,
		CreatedAt:   t.CreatedAt.Time,
	}
}

func convertTranslations(translations []sqlc.Translation) []db.Translation {
	result := make([]db.Translation, len(translations))
	for i, t := range translations {
		result[i] = convertTranslation(t)
	}
	return result
}

func convertPublicTranslation(t sqlc.PublicTranslation) db.PublicTranslation {
	return db.PublicTranslation{
		ID:           t.ID,
		Username:     t.Username,
		Translation:  t.Translation,
		Explanation:  fromPgText(t.Explanation),
		Language:     t.Language,
		Region:       t.Region,
		SourceBotID:  fromPgText(t.SourceBotID),
		RiotVerified: t.RiotVerified,
		Upvotes:      t.Upvotes,
		Downvotes:    t.Downvotes,
		CreatedAt:    t.CreatedAt.Time,
	}
}

func convertPublicTranslations(translations []sqlc.PublicTranslation) []db.PublicTranslation {
	result := make([]db.PublicTranslation, len(translations))
	for i, t := range translations {
		result[i] = convertPublicTranslation(t)
	}
	return result
}

func convertVote(v sqlc.Vote) db.Vote {
	return db.Vote{
		ID:            v.ID,
		TranslationID: v.TranslationID,
		IpHash:        v.IpHash,
		Vote:          v.Vote,
		CreatedAt:     v.CreatedAt.Time,
	}
}

func toPgInt8(n sql.NullInt64) pgtype.Int8 {
	return pgtype.Int8{Int64: n.Int64, Valid: n.Valid}
}

func fromPgInt8(n pgtype.Int8) sql.NullInt64 {
	return sql.NullInt64{Int64: n.Int64, Valid: n.Valid}
}

func toPgText(s sql.NullString) pgtype.Text {
	return pgtype.Text{String: s.String, Valid: s.Valid}
}

func fromPgText(t pgtype.Text) sql.NullString {
	return sql.NullString{String: t.String, Valid: t.Valid}
}
