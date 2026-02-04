package google

import (
	"context"
	"fmt"

	"github.com/jusunglee/leagueofren/internal/llm"
	"google.golang.org/genai"
)

// Model represents a Google AI model identifier
type Model string

const (
	ModelGemma3_27B   Model = "gemma-3-27b-it"
	ModelGemini2Flash Model = "gemini-2.0-flash"
	ModelGemini2_5Pro Model = "gemini-2.5-pro"
)

var DefaultModel Model = ModelGemma3_27B

type Client struct {
	client *genai.Client
	model  Model
}

func NewClient(ctx context.Context, apiKey string, model Model) (*Client, error) {
	if model == "" {
		model = DefaultModel
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create google client: %w", err)
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

func (c *Client) Complete(ctx context.Context, system, prompt string) (string, error) {
	// Gemma doesn't support system instructions natively, prepend to user message
	fullPrompt := system + "\n\n" + prompt

	result, err := c.client.Models.GenerateContent(ctx, string(c.model),
		[]*genai.Content{{Parts: []*genai.Part{{Text: fullPrompt}}}},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("google API call failed: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from google")
	}

	text := result.Candidates[0].Content.Parts[0].Text

	return llm.StripMarkdownCodeBlocks(text), nil
}
