package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	repo, err := New(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })
	return repo
}

func TestSubscriptionCRUD(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	sub, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1",
		LolUsername:      "Player#NA1",
		Region:           "NA",
		ServerID:         "server-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "Player#NA1", sub.LolUsername)
	assert.Equal(t, "NA", sub.Region)

	got, err := repo.GetSubscriptionByID(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, sub.ID, got.ID)

	byChan, err := repo.GetSubscriptionsByChannel(ctx, "chan-1")
	require.NoError(t, err)
	assert.Len(t, byChan, 1)

	count, err := repo.CountSubscriptionsByServer(ctx, "server-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	all, err := repo.GetAllSubscriptions(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// Duplicate returns ErrNoRows
	_, err = repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1",
		LolUsername:      "Player#NA1",
		Region:           "NA",
		ServerID:         "server-1",
	})
	assert.True(t, db.IsNoRows(err))
}

func TestDeleteSubscription(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1",
		LolUsername:      "Player#NA1",
		Region:           "NA",
		ServerID:         "server-1",
	})
	require.NoError(t, err)

	rows, err := repo.DeleteSubscription(ctx, db.DeleteSubscriptionParams{
		DiscordChannelID: "chan-1",
		LolUsername:      "Player#NA1",
		Region:           "NA",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	all, err := repo.GetAllSubscriptions(ctx, 100)
	require.NoError(t, err)
	assert.Empty(t, all)
}

func TestEvalCRUD(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	sub, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1",
		LolUsername:      "Player#NA1",
		Region:           "NA",
		ServerID:         "server-1",
	})
	require.NoError(t, err)

	eval, err := repo.CreateEval(ctx, db.CreateEvalParams{
		SubscriptionID:   sub.ID,
		EvalStatus:       "NEW_TRANSLATIONS",
		DiscordMessageID: sql.NullString{String: "msg-1", Valid: true},
		GameID:           sql.NullInt64{Int64: 999, Valid: true},
	})
	require.NoError(t, err)
	assert.Equal(t, "NEW_TRANSLATIONS", eval.EvalStatus)

	got, err := repo.GetEvalByGameAndSubscription(ctx, db.GetEvalByGameAndSubscriptionParams{
		GameID:         sql.NullInt64{Int64: 999, Valid: true},
		SubscriptionID: sub.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, eval.ID, got.ID)

	latest, err := repo.GetLatestEvalForSubscription(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, eval.ID, latest.ID)

	// Not found
	_, err = repo.GetEvalByGameAndSubscription(ctx, db.GetEvalByGameAndSubscriptionParams{
		GameID:         sql.NullInt64{Int64: 0, Valid: true},
		SubscriptionID: sub.ID,
	})
	assert.True(t, db.IsNoRows(err))
}

func TestTranslationCRUD(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	tr, err := repo.CreateTranslation(ctx, db.CreateTranslationParams{
		Username:    "玩家",
		Translation: "Player",
		Provider:    "anthropic",
		Model:       "claude-3",
	})
	require.NoError(t, err)
	assert.Equal(t, "Player", tr.Translation)

	got, err := repo.GetTranslation(ctx, "玩家")
	require.NoError(t, err)
	assert.Equal(t, tr.ID, got.ID)

	batch, err := repo.GetTranslations(ctx, []string{"玩家", "nonexistent"})
	require.NoError(t, err)
	assert.Len(t, batch, 1)

	empty, err := repo.GetTranslations(ctx, []string{})
	require.NoError(t, err)
	assert.Empty(t, empty)

	// Upsert
	updated, err := repo.CreateTranslation(ctx, db.CreateTranslationParams{
		Username:    "玩家",
		Translation: "Gamer",
		Provider:    "google",
		Model:       "gemini",
	})
	require.NoError(t, err)
	assert.Equal(t, "Gamer", updated.Translation)
	assert.Equal(t, "google", updated.Provider)
}

func TestFeedback(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	fb, err := repo.CreateFeedback(ctx, db.CreateFeedbackParams{
		DiscordMessageID: "msg-1",
		FeedbackText:     "👍",
	})
	require.NoError(t, err)
	assert.Equal(t, "msg-1", fb.DiscordMessageID)
	assert.Equal(t, "👍", fb.FeedbackText)
}

func TestDeleteOldTranslations(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.CreateTranslation(ctx, db.CreateTranslationParams{
		Username: "old_user", Translation: "Old", Provider: "test", Model: "test",
	})
	require.NoError(t, err)

	_, err = repo.CreateTranslation(ctx, db.CreateTranslationParams{
		Username: "new_user", Translation: "New", Provider: "test", Model: "test",
	})
	require.NoError(t, err)

	// Delete translations older than 1 second in the future (should delete all)
	deleted, err := repo.DeleteOldTranslations(ctx, time.Now().Add(time.Second))
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	remaining, err := repo.GetTranslations(ctx, []string{"old_user", "new_user"})
	require.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestDeleteOldFeedback(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.CreateFeedback(ctx, db.CreateFeedbackParams{
		DiscordMessageID: "msg-1", FeedbackText: "good",
	})
	require.NoError(t, err)

	deleted, err := repo.DeleteOldFeedback(ctx, time.Now().Add(time.Second))
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}

func TestWithTxCommit(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.WithTx(ctx, func(txRepo db.Repository) error {
		_, err := txRepo.CreateSubscription(ctx, db.CreateSubscriptionParams{
			DiscordChannelID: "chan-tx",
			LolUsername:      "TxPlayer#NA1",
			Region:           "NA",
			ServerID:         "server-tx",
		})
		return err
	})
	require.NoError(t, err)

	subs, err := repo.GetSubscriptionsByChannel(ctx, "chan-tx")
	require.NoError(t, err)
	assert.Len(t, subs, 1)
}

func TestWithTxRollback(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.WithTx(ctx, func(txRepo db.Repository) error {
		_, err := txRepo.CreateSubscription(ctx, db.CreateSubscriptionParams{
			DiscordChannelID: "chan-rollback",
			LolUsername:      "RollbackPlayer#NA1",
			Region:           "NA",
			ServerID:         "server-rb",
		})
		if err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	require.Error(t, err)

	subs, err := repo.GetSubscriptionsByChannel(ctx, "chan-rollback")
	require.NoError(t, err)
	assert.Empty(t, subs)
}

func TestAccountCache(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.CacheAccount(ctx, db.CacheAccountParams{
		GameName: "Player", TagLine: "NA1", Region: "NA", Puuid: "puuid-123",
	})
	require.NoError(t, err)

	got, err := repo.GetCachedAccount(ctx, db.GetCachedAccountParams{
		GameName: "Player", TagLine: "NA1", Region: "NA",
	})
	require.NoError(t, err)
	assert.Equal(t, "puuid-123", got.Puuid)

	// Miss
	_, err = repo.GetCachedAccount(ctx, db.GetCachedAccountParams{
		GameName: "Nobody", TagLine: "XX", Region: "NA",
	})
	assert.True(t, db.IsNoRows(err))
}

func TestGameCache(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	participants, _ := json.Marshal([]map[string]string{{"puuid": "p1", "riotId": "P1#NA1"}})
	err := repo.CacheGameStatus(ctx, db.CacheGameStatusParams{
		Puuid:        "puuid-1",
		Region:       "NA",
		InGame:       true,
		GameID:       sql.NullInt64{Int64: 42, Valid: true},
		Participants: participants,
	})
	require.NoError(t, err)

	got, err := repo.GetCachedGameStatus(ctx, db.GetCachedGameStatusParams{
		Puuid: "puuid-1", Region: "NA",
	})
	require.NoError(t, err)
	assert.True(t, got.InGame)
	assert.Equal(t, int64(42), got.GameID.Int64)

	// Miss
	_, err = repo.GetCachedGameStatus(ctx, db.GetCachedGameStatusParams{
		Puuid: "nonexistent", Region: "NA",
	})
	assert.True(t, db.IsNoRows(err))
}

func TestDeleteExpiredCaches(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.CacheAccount(ctx, db.CacheAccountParams{
		GameName: "Player", TagLine: "NA1", Region: "NA", Puuid: "puuid-1",
	})
	require.NoError(t, err)

	// Should not delete non-expired entries
	require.NoError(t, repo.DeleteExpiredAccountCache(ctx))
	_, err = repo.GetCachedAccount(ctx, db.GetCachedAccountParams{
		GameName: "Player", TagLine: "NA1", Region: "NA",
	})
	require.NoError(t, err)

	require.NoError(t, repo.DeleteExpiredGameCache(ctx))
}

func TestDeleteEvals(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	sub, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1",
		LolUsername:      "Player#NA1",
		Region:           "NA",
		ServerID:         "server-1",
	})
	require.NoError(t, err)

	_, err = repo.CreateEval(ctx, db.CreateEvalParams{
		SubscriptionID:   sub.ID,
		EvalStatus:       "NEW_TRANSLATIONS",
		DiscordMessageID: sql.NullString{String: "msg-1", Valid: true},
		GameID:           sql.NullInt64{Int64: 100, Valid: true},
	})
	require.NoError(t, err)

	deleted, err := repo.DeleteEvals(ctx, time.Now().Add(time.Second))
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}

func TestTranslationToEval(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	sub, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1", LolUsername: "P#1", Region: "NA", ServerID: "s-1",
	})
	require.NoError(t, err)

	eval, err := repo.CreateEval(ctx, db.CreateEvalParams{
		SubscriptionID: sub.ID, EvalStatus: "NEW_TRANSLATIONS",
		DiscordMessageID: sql.NullString{String: "msg-1", Valid: true},
		GameID:           sql.NullInt64{Int64: 1, Valid: true},
	})
	require.NoError(t, err)

	tr, err := repo.CreateTranslation(ctx, db.CreateTranslationParams{
		Username: "玩家", Translation: "Player", Provider: "test", Model: "test",
	})
	require.NoError(t, err)

	err = repo.CreateTranslationToEval(ctx, db.CreateTranslationToEvalParams{
		TranslationID: tr.ID, EvalID: eval.ID,
	})
	require.NoError(t, err)

	translations, err := repo.GetTranslationsForEval(ctx, eval.ID)
	require.NoError(t, err)
	assert.Len(t, translations, 1)
	assert.Equal(t, "玩家", translations[0].Username)
}

func TestUpdateSubscriptionLastEvaluatedAt(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	sub, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1", LolUsername: "P#1", Region: "NA", ServerID: "s-1",
	})
	require.NoError(t, err)

	err = repo.UpdateSubscriptionLastEvaluatedAt(ctx, sub.ID)
	require.NoError(t, err)
}

func TestDeleteSubscriptions(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	sub1, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1", LolUsername: "P1#1", Region: "NA", ServerID: "s-1",
	})
	require.NoError(t, err)

	sub2, err := repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: "chan-1", LolUsername: "P2#1", Region: "NA", ServerID: "s-1",
	})
	require.NoError(t, err)

	deleted, err := repo.DeleteSubscriptions(ctx, []int64{sub1.ID, sub2.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	// Empty slice
	deleted, err = repo.DeleteSubscriptions(ctx, []int64{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}
