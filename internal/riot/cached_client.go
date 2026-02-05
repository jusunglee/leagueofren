package riot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jusunglee/leagueofren/internal/db"
)

type CachedClient struct {
	client *client
	repo   db.Repository
}

func NewCachedClient(apiKey string, repo db.Repository) *CachedClient {
	return &CachedClient{
		client: newClient(apiKey),
		repo:   repo,
	}
}

func (c *CachedClient) GetAccountByRiotID(ctx context.Context, gameName, tagLine, region string) (Account, error) {
	cached, err := c.repo.GetCachedAccount(ctx, db.GetCachedAccountParams{
		GameName: gameName,
		TagLine:  tagLine,
		Region:   region,
	})
	if err == nil {
		return Account{
			PUUID:    cached.Puuid,
			GameName: cached.GameName,
			TagLine:  cached.TagLine,
		}, nil
	}
	if !db.IsNoRows(err) {
		return Account{}, fmt.Errorf("account cache lookup failed: %w", err)
	}

	account, err := c.client.GetAccountByRiotID(gameName, tagLine, region)
	if err != nil {
		return Account{}, err
	}

	if err := c.repo.CacheAccount(ctx, db.CacheAccountParams{
		GameName: account.GameName,
		TagLine:  account.TagLine,
		Region:   region,
		Puuid:    account.PUUID,
	}); err != nil {
		return Account{}, fmt.Errorf("failed to cache account: %w", err)
	}

	return account, nil
}

func (c *CachedClient) GetActiveGame(ctx context.Context, puuid, region string) (ActiveGame, error) {
	cached, err := c.repo.GetCachedGameStatus(ctx, db.GetCachedGameStatusParams{
		Puuid:  puuid,
		Region: region,
	})
	if err == nil {
		if !cached.InGame {
			return ActiveGame{}, ErrNotInGame
		}

		var participants []Participant
		if cached.Participants != nil {
			if err := json.Unmarshal(cached.Participants, &participants); err != nil {
				return ActiveGame{}, fmt.Errorf("failed to unmarshal cached participants: %w", err)
			}
		}

		return ActiveGame{
			GameID:       cached.GameID.Int64,
			Participants: participants,
		}, nil
	}
	if !db.IsNoRows(err) {
		return ActiveGame{}, fmt.Errorf("game cache lookup failed: %w", err)
	}

	game, err := c.client.GetActiveGame(puuid, region)
	if errors.Is(err, ErrNotInGame) {
		if cacheErr := c.repo.CacheGameStatus(ctx, db.CacheGameStatusParams{
			Puuid:  puuid,
			Region: region,
			InGame: false,
		}); cacheErr != nil {
			return ActiveGame{}, fmt.Errorf("failed to cache not-in-game game status: %w", cacheErr)
		}
		return ActiveGame{}, ErrNotInGame
	}
	if err != nil {
		return ActiveGame{}, err
	}

	participantsJSON, err := json.Marshal(game.Participants)
	if err != nil {
		return ActiveGame{}, fmt.Errorf("failed to marshal participants: %w", err)
	}

	if err := c.repo.CacheGameStatus(ctx, db.CacheGameStatusParams{
		Puuid:        puuid,
		Region:       region,
		InGame:       true,
		GameID:       sql.NullInt64{Int64: game.GameID, Valid: true},
		Participants: participantsJSON,
	}); err != nil {
		return ActiveGame{}, fmt.Errorf("failed to cache game status: %w", err)
	}

	return game, nil
}
