package riot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrNotFound = errors.New("account not found")
var ErrInvalidRegion = errors.New("invalid region")

var ValidRegions = []string{
	"NA",
	"EUW",
	"EUNE",
	"KR",
	"JP",
	"BR",
	"LAN",
	"LAS",
	"OCE",
	"TR",
	"RU",
}

var regionToRoutingURL = map[string]string{
	"NA":   "https://americas.api.riotgames.com",
	"BR":   "https://americas.api.riotgames.com",
	"LAN":  "https://americas.api.riotgames.com",
	"LAS":  "https://americas.api.riotgames.com",
	"EUW":  "https://europe.api.riotgames.com",
	"EUNE": "https://europe.api.riotgames.com",
	"TR":   "https://europe.api.riotgames.com",
	"RU":   "https://europe.api.riotgames.com",
	"KR":   "https://asia.api.riotgames.com",
	"JP":   "https://asia.api.riotgames.com",
	"OCE":  "https://sea.api.riotgames.com",
}

type Account struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func IsValidRegion(region string) bool {
	for _, r := range ValidRegions {
		if r == region {
			return true
		}
	}
	return false
}

func getRegionalURL(region string) (string, error) {
	url, ok := regionToRoutingURL[region]
	if !ok {
		return "", ErrInvalidRegion
	}
	return url, nil
}

func (c *Client) GetAccountByRiotID(gameName, tagLine, region string) (*Account, error) {
	baseURL, err := getRegionalURL(region)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/riot/account/v1/accounts/by-riot-id/%s/%s",
		baseURL,
		url.PathEscape(gameName),
		url.PathEscape(tagLine),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &account, nil
}

func ParseRiotID(input string) (gameName, tagLine string, err error) {
	input = strings.TrimSpace(input)

	parts := strings.Split(input, "#")
	if len(parts) != 2 {
		return "", "", errors.New("invalid format: expected 'name#tag' or 'name #tag'")
	}

	gameName = strings.TrimSpace(parts[0])
	tagLine = strings.TrimSpace(parts[1])

	if gameName == "" || tagLine == "" {
		return "", "", errors.New("invalid format: name and tag cannot be empty")
	}

	return gameName, tagLine, nil
}
