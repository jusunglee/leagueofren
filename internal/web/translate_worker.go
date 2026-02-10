package web

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
	"unicode"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/jobs"
	"github.com/jusunglee/leagueofren/internal/metrics"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/riverqueue/river"
)

type TranslateWorker struct {
	river.WorkerDefaults[jobs.TranslateUsernameArgs]
	repo       db.Repository
	riot       *riot.DirectClient
	translator *translation.Translator
	log        *slog.Logger
}

func NewTranslateWorker(repo db.Repository, riotClient *riot.DirectClient, translator *translation.Translator, log *slog.Logger) *TranslateWorker {
	return &TranslateWorker{repo: repo, riot: riotClient, translator: translator, log: log}
}

func (w *TranslateWorker) Work(ctx context.Context, job *river.Job[jobs.TranslateUsernameArgs]) error {
	username := job.Args.Username
	region := job.Args.Region

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		// Invalid format — no point retrying.
		return fmt.Errorf("invalid riot ID %q: %w", username, err)
	}

	account, err := w.riot.GetAccountByRiotID(gameName, tagLine, region)
	if err != nil {
		return fmt.Errorf("riot lookup for %q: %w", username, err)
	}
	puuid := account.PUUID

	llmStart := time.Now()
	translations, err := w.translator.TranslateUsernames(ctx, []string{gameName})
	metrics.LLMTranslationDuration.Observe(time.Since(llmStart).Seconds())
	if err != nil || len(translations) == 0 {
		metrics.TranslationSubmissions.WithLabelValues("failed").Inc()
		return fmt.Errorf("translation failed for %q: %w", username, err)
	}

	t := translations[0]
	language := detectLanguageFromName(gameName)

	err = w.repo.WithTx(ctx, func(txRepo db.Repository) error {
		_, err := txRepo.UpsertPlayer(ctx, db.UpsertPlayerParams{
			Username: username,
			Region:   region,
			Puuid:    sql.NullString{String: puuid, Valid: puuid != ""},
		})
		if err != nil {
			return err
		}

		params := db.UpsertPublicTranslationParams{
			Username:       username,
			Translation:    t.Translated,
			Language:       language,
			PlayerUsername: username,
			RiotVerified:   tagLine != "",
		}
		if t.Explanation != "" {
			params.Explanation = sql.NullString{String: t.Explanation, Valid: true}
		}

		_, err = txRepo.UpsertPublicTranslation(ctx, params)
		return err
	})
	if err != nil {
		metrics.TranslationSubmissions.WithLabelValues("failed").Inc()
		return fmt.Errorf("upserting translation for %q: %w", username, err)
	}

	metrics.TranslationSubmissions.WithLabelValues("success").Inc()
	w.log.InfoContext(ctx, "translated username", "username", username, "region", region, "translation", t.Translated)
	return nil
}

func detectLanguageFromName(name string) string {
	for _, r := range name {
		if unicode.Is(unicode.Hangul, r) {
			return "korean"
		}
	}
	return "chinese"
}
