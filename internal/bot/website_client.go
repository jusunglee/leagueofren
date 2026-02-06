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

// WebsiteClient submits translations to the companion website API.
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
	Username     string  `json:"username"`
	Translation  string  `json:"translation"`
	Explanation  *string `json:"explanation,omitempty"`
	Language     string  `json:"language"`
	Region       string  `json:"region"`
	SourceBotID  *string `json:"source_bot_id,omitempty"`
	RiotVerified bool    `json:"riot_verified"`
}

func (w *WebsiteClient) SubmitTranslations(ctx context.Context, translations []translation.Translation, language, region string) error {
	if !w.Enabled() {
		return nil
	}

	for _, t := range translations {
		body := websiteSubmission{
			Username:    t.Original,
			Translation: t.Translated,
			Language:    language,
			Region:      region,
		}
		if t.Explanation != "" {
			body.Explanation = &t.Explanation
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
			return fmt.Errorf("submitting translation for %s: %w", t.Original, err)
		}
		resp.Body.Close()
	}

	return nil
}
