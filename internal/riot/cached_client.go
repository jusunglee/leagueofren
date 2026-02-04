package riot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jusunglee/leagueofren/internal/db"
)

type CachedClient struct {
	client  *client
	queries *db.Queries
}

func NewCachedClient(apiKey string, queries *db.Queries) *CachedClient {
	return &CachedClient{
		client:  newClient(apiKey),
		queries: queries,
	}
}

func (c *CachedClient) GetAccountByRiotID(ctx context.Context, gameName, tagLine, region string) (Account, error) {
	cached, err := c.queries.GetCachedAccount(ctx, db.GetCachedAccountParams{
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
	if !errors.Is(err, pgx.ErrNoRows) {
		return Account{}, fmt.Errorf("account cache lookup failed: %w", err)
	}

	account, err := c.client.GetAccountByRiotID(gameName, tagLine, region)
	if err != nil {
		return Account{}, err
	}

	if err := c.queries.CacheAccount(ctx, db.CacheAccountParams{
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
	cached, err := c.queries.GetCachedGameStatus(ctx, db.GetCachedGameStatusParams{
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
	if !errors.Is(err, pgx.ErrNoRows) {
		return ActiveGame{}, fmt.Errorf("game cache lookup failed: %w", err)
	}

	game, err := c.client.GetActiveGame(puuid, region)
	if errors.Is(err, ErrNotInGame) {
		if cacheErr := c.queries.CacheGameStatus(ctx, db.CacheGameStatusParams{
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

	if err := c.queries.CacheGameStatus(ctx, db.CacheGameStatusParams{
		Puuid:        puuid,
		Region:       region,
		InGame:       true,
		GameID:       pgtype.Int8{Int64: game.GameID, Valid: true},
		Participants: participantsJSON,
	}); err != nil {
		return ActiveGame{}, fmt.Errorf("failed to cache game status: %w", err)
	}

	return game, nil
}
