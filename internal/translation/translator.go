package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/llm"
)

type Translator struct {
	llm      llm.Client
	repo     db.Repository
	provider string
	model    string
}

type Translation struct {
	Original    string `json:"original"`
	Translated  string `json:"translated"`
	Explanation string `json:"explanation,omitempty"`
}

func NewTranslator(client llm.Client, repo db.Repository, provider, model string) *Translator {
	return &Translator{
		llm:      client,
		repo:     repo,
		provider: provider,
		model:    model,
	}
}

const systemPrompt = `You are translating League of Legends summoner names from Korean and Chinese to English.

For each name, provide:
1. The English translation or transliteration
2. Brief context if it's a cultural reference, pun, pro player name, or gaming term

Respond ONLY with a JSON array, no other text. Example:
[
  {"original": "不知火舞", "translated": "Mai Shiranui", "explanation": "Fighting game character from Fatal Fury/KOF"},
  {"original": "人人人", "translated": "Person Person Person", "explanation": ""}
]`

func (t *Translator) TranslateUsernames(ctx context.Context, usernames []string) ([]Translation, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	cached, err := t.repo.GetTranslations(ctx, usernames)
	if err != nil {
		return nil, fmt.Errorf("cache lookup failed: %w", err)
	}

	cachedMap := make(map[string]string, len(cached))
	for _, c := range cached {
		cachedMap[c.Username] = c.Translation
	}

	var results []Translation
	var uncached []string

	for _, username := range usernames {
		if translation, ok := cachedMap[username]; ok {
			results = append(results, Translation{
				Original:   username,
				Translated: translation,
			})
		} else {
			uncached = append(uncached, username)
		}
	}

	if len(uncached) == 0 {
		return results, nil
	}

	var sb strings.Builder
	sb.WriteString("Translate these summoner names:\n")
	for _, name := range uncached {
		sb.WriteString("- ")
		sb.WriteString(name)
		sb.WriteString("\n")
	}

	text, err := t.llm.Complete(ctx, systemPrompt, sb.String())
	if err != nil {
		return nil, err
	}

	var translations []Translation
	if err := json.Unmarshal([]byte(text), &translations); err != nil {
		return nil, fmt.Errorf("failed to parse translation response: %w (response: %s)", err, text)
	}

	for _, tr := range translations {
		composed := composeTranslation(tr)
		_, err := t.repo.CreateTranslation(ctx, db.CreateTranslationParams{
			Username:    tr.Original,
			Translation: composed,
			Provider:    t.provider,
			Model:       t.model,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to cache translation for %s: %w", tr.Original, err)
		}
		results = append(results, Translation{
			Original:   tr.Original,
			Translated: composed,
		})
	}

	return results, nil
}

func composeTranslation(tr Translation) string {
	if tr.Explanation == "" {
		return tr.Translated
	}
	return fmt.Sprintf("%s (%s)", tr.Translated, tr.Explanation)
}
