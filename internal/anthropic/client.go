package anthropic

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jusunglee/leagueofren/internal/llm"
)

// Re-export Model type and constants for external use
type Model = anthropic.Model

const (
	ModelClaudeSonnet4_5 Model = anthropic.ModelClaudeSonnet4_5_20250929
	ModelClaudeHaiku4_5  Model = anthropic.ModelClaudeHaiku4_5_20251001
	ModelClaudeOpus4_5   Model = anthropic.ModelClaudeOpus4_5_20251101
)

var DefaultModel Model = ModelClaudeSonnet4_5

type Client struct {
	client anthropic.Client
	model  Model
}

func NewClient(apiKey string, model Model) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

func (c *Client) Complete(ctx context.Context, system, prompt string) (string, error) {
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic API call failed: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic")
	}

	var text string
	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			text = textBlock.Text
			break
		}
	}

	if text == "" {
		return "", fmt.Errorf("no text content in response")
	}

	return llm.StripMarkdownCodeBlocks(text), nil
}
