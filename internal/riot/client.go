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

var regionToPlatformURL = map[string]string{
	"NA":   "https://na1.api.riotgames.com",
	"BR":   "https://br1.api.riotgames.com",
	"LAN":  "https://la1.api.riotgames.com",
	"LAS":  "https://la2.api.riotgames.com",
	"EUW":  "https://euw1.api.riotgames.com",
	"EUNE": "https://eun1.api.riotgames.com",
	"TR":   "https://tr1.api.riotgames.com",
	"RU":   "https://ru.api.riotgames.com",
	"KR":   "https://kr.api.riotgames.com",
	"JP":   "https://jp1.api.riotgames.com",
	"OCE":  "https://oc1.api.riotgames.com",
}

type Account struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

var ErrNotInGame = errors.New("player not in game")

type ActiveGame struct {
	GameID       int64         `json:"gameId"`
	Participants []Participant `json:"participants"`
}

type Participant struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"riotId"`
}

type client struct {
	apiKey     string
	httpClient *http.Client
}

func newClient(apiKey string) *client {
	return &client{
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

func getPlatformURL(region string) (string, error) {
	url, ok := regionToPlatformURL[region]
	if !ok {
		return "", ErrInvalidRegion
	}
	return url, nil
}

func (c *client) GetAccountByRiotID(gameName, tagLine, region string) (Account, error) {
	baseURL, err := getRegionalURL(region)
	if err != nil {
		return Account{}, err
	}

	endpoint := fmt.Sprintf("%s/riot/account/v1/accounts/by-riot-id/%s/%s",
		baseURL,
		url.PathEscape(gameName),
		url.PathEscape(tagLine),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return Account{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Account{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Account{}, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return Account{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return Account{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return account, nil
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

func (c *client) GetActiveGame(puuid, region string) (ActiveGame, error) {
	baseURL, err := getPlatformURL(region)
	if err != nil {
		return ActiveGame{}, err
	}

	endpoint := fmt.Sprintf("%s/lol/spectator/v5/active-games/by-summoner/%s",
		baseURL,
		url.PathEscape(puuid),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return ActiveGame{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ActiveGame{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ActiveGame{}, ErrNotInGame
	}

	if resp.StatusCode != http.StatusOK {
		return ActiveGame{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var game ActiveGame
	if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
		return ActiveGame{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return game, nil
}
