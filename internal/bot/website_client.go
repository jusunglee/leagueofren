package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jusunglee/leagueofren/internal/translation"
)

// WebsiteClient submits usernames to the companion website for server-side
// translation. The website validates the username via Riot API and runs its
// own LLM translation, so the bot only needs to send username + region.
// If URL is empty, all calls are no-ops (opt-out by default).
type WebsiteClient struct {
	url  string
	http *http.Client
}

func NewWebsiteClient(url string) *WebsiteClient {
	return &WebsiteClient{
		url:  url,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebsiteClient) Enabled() bool {
	return w.url != ""
}

type websiteSubmission struct {
	Username string `json:"username"`
	Region   string `json:"region"`
}

func (w *WebsiteClient) SubmitTranslations(ctx context.Context, translations []translation.Translation, riotIDs map[string]string, region string) error {
	if !w.Enabled() {
		return nil
	}

	for _, t := range translations {
		// Use full Riot ID (name#tag) if available, fall back to game name
		username := t.Original
		if fullID, ok := riotIDs[t.Original]; ok {
			username = fullID
		}
		body := websiteSubmission{
			Username: username,
			Region:   region,
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling submission: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", w.url+"/api/v1/translations", bytes.NewReader(jsonBody))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.http.Do(req)
		if err != nil {
			return fmt.Errorf("submitting username %s: %w", t.Original, err)
		}
		resp.Body.Close()
	}

	return nil
}
