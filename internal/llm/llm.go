package llm

import (
	"context"
	"strings"
)

type Client interface {
	Complete(ctx context.Context, system, prompt string) (string, error)
}

// StripMarkdownCodeBlocks removes ```...``` wrappers from LLM responses
func StripMarkdownCodeBlocks(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		if idx := strings.Index(text, "\n"); idx != -1 {
			text = text[idx+1:]
		}
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
		text = strings.TrimSpace(text)
	}
	return text
}
