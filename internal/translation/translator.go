package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jusunglee/leagueofren/internal/llm"
)

type Translator struct {
	llm llm.Client
}

type Translation struct {
	Original    string `json:"original"`
	Translated  string `json:"translated"`
	Explanation string `json:"explanation,omitempty"`
}

func NewTranslator(client llm.Client) *Translator {
	return &Translator{llm: client}
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

	var sb strings.Builder
	sb.WriteString("Translate these summoner names:\n")
	for _, name := range usernames {
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

	return translations, nil
}
