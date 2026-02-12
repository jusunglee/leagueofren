package db

import (
	"context"
	"database/sql"
	"time"
)

type Player struct {
	Username     string
	Region       string
	Rank         sql.NullString
	TopChampions sql.NullString
	Puuid        sql.NullString
	FirstSeen    time.Time
	LastUpdated  time.Time
}

type UpsertPlayerParams struct {
	Username     string
	Region       string
	Rank         sql.NullString
	TopChampions sql.NullString
	Puuid        sql.NullString
}

type UpdatePlayerStatsParams struct {
	Username     string
	Rank         sql.NullString
	TopChampions sql.NullString
}

type PublicTranslation struct {
	ID           int64
	Username     string
	Translation  string
	Explanation  sql.NullString
	Language     string
	Region       string
	SourceBotID  sql.NullString
	RiotVerified bool
	Rank         sql.NullString
	TopChampions sql.NullString
	Upvotes      int32
	Downvotes    int32
	CreatedAt    time.Time
	FirstSeen    time.Time
}

// Vote represents an IP-hashed vote on a public translation
type Vote struct {
	ID            int64
	TranslationID int64
	IpHash        string
	VisitorID     string
	Vote          int16
	CreatedAt     time.Time
}

// PublicFeedback represents user feedback on a public translation
type PublicFeedback struct {
	ID            int64
	TranslationID int64
	IpHash        string
	FeedbackText  string
	CreatedAt     time.Time
}

type UpsertPublicTranslationParams struct {
	Username       string
	Translation    string
	Explanation    sql.NullString
	Language       string
	PlayerUsername string
	SourceBotID    sql.NullString
	RiotVerified   bool
}

type ListPublicTranslationsNewParams struct {
	Region   string
	Language string
	Limit    int32
	Offset   int32
}

type ListPublicTranslationsTopParams struct {
	Region    string
	Language  string
	Limit     int32
	Offset    int32
	CreatedAt time.Time
}

type CountPublicTranslationsParams struct {
	Region   string
	Language string
}

type UpsertVoteParams struct {
	TranslationID int64
	IpHash        string
	VisitorID     string
	Vote          int16
}

type GetVoteParams struct {
	TranslationID int64
	VisitorID     string
}

type DeleteVoteParams struct {
	TranslationID int64
	VisitorID     string
}

type CreatePublicFeedbackParams struct {
	TranslationID int64
	IpHash        string
	FeedbackText  string
}

type ListPublicFeedbackParams struct {
	Limit  int32
	Offset int32
}

type ListPublicFeedbackRow struct {
	ID            int64
	TranslationID int64
	IpHash        string
	FeedbackText  string
	CreatedAt     time.Time
	Username      string
	Translation   string
}

// Subscription represents a user's subscription to track a LoL player
type Subscription struct {
	ID               int64
	DiscordChannelID string
	ServerID         string
	LolUsername      string
	Region           string
	CreatedAt        time.Time
	LastEvaluatedAt  time.Time
}

// Eval represents an evaluation of a subscription (checking if player is in game)
type Eval struct {
	ID               int64
	SubscriptionID   int64
	GameID           sql.NullInt64
	EvaluatedAt      time.Time
	EvalStatus       string
	DiscordMessageID sql.NullString
}

// Translation represents a cached translation of a username
type Translation struct {
	ID          int64
	Username    string
	Translation string
	Provider    string
	Model       string
	CreatedAt   time.Time
}

// Feedback represents user feedback on a translation
type Feedback struct {
	ID               int64
	DiscordMessageID string
	FeedbackText     string
	CreatedAt        time.Time
}

// RiotAccountCache represents cached Riot account info
type RiotAccountCache struct {
	ID        int64
	GameName  string
	TagLine   string
	Region    string
	Puuid     string
	CachedAt  time.Time
	ExpiresAt time.Time
}

// RiotGameCache represents cached game status
type RiotGameCache struct {
	ID           int64
	Puuid        string
	Region       string
	InGame       bool
	GameID       sql.NullInt64
	Participants []byte
	CachedAt     time.Time
	ExpiresAt    time.Time
}

// FindSubscriptionsWithExpiredNewestOnlineEvalRow is the result of FindSubscriptionsWithExpiredNewestOnlineEval
type FindSubscriptionsWithExpiredNewestOnlineEvalRow struct {
	SubscriptionID   int64
	NewestOnlineEval time.Time
}

// Parameter structs for repository methods

type CreateSubscriptionParams struct {
	DiscordChannelID string
	LolUsername      string
	Region           string
	ServerID         string
}

type DeleteSubscriptionParams struct {
	DiscordChannelID string
	LolUsername      string
	Region           string
}

type CreateEvalParams struct {
	SubscriptionID   int64
	EvalStatus       string
	DiscordMessageID sql.NullString
	GameID           sql.NullInt64
}

type GetEvalByGameAndSubscriptionParams struct {
	GameID         sql.NullInt64
	SubscriptionID int64
}

type CreateTranslationParams struct {
	Username    string
	Translation string
	Provider    string
	Model       string
}

type CreateTranslationToEvalParams struct {
	TranslationID int64
	EvalID        int64
}

type CreateFeedbackParams struct {
	DiscordMessageID string
	FeedbackText     string
}

type GetCachedAccountParams struct {
	GameName string
	TagLine  string
	Region   string
}

type GetCachedAccountRow struct {
	GameName string
	TagLine  string
	Region   string
	Puuid    string
}

type CacheAccountParams struct {
	GameName string
	TagLine  string
	Region   string
	Puuid    string
}

type GetCachedGameStatusParams struct {
	Puuid  string
	Region string
}

type GetCachedGameStatusRow struct {
	Puuid        string
	Region       string
	InGame       bool
	GameID       sql.NullInt64
	Participants []byte
}

type CacheGameStatusParams struct {
	Puuid        string
	Region       string
	InGame       bool
	GameID       sql.NullInt64
	Participants []byte
}

// Repository defines the interface for database operations
type Repository interface {
	// Subscriptions
	CreateSubscription(ctx context.Context, arg CreateSubscriptionParams) (Subscription, error)
	GetAllSubscriptions(ctx context.Context) ([]Subscription, error)
	GetSubscriptionsByChannel(ctx context.Context, discordChannelID string) ([]Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int64) (Subscription, error)
	CountSubscriptionsByServer(ctx context.Context, serverID string) (int64, error)
	DeleteSubscription(ctx context.Context, arg DeleteSubscriptionParams) (int64, error)
	DeleteSubscriptions(ctx context.Context, ids []int64) (int64, error)
	DeleteSubscriptionsByServer(ctx context.Context, serverID string) (int64, error)
	UpdateSubscriptionLastEvaluatedAt(ctx context.Context, id int64) error

	// Evals
	CreateEval(ctx context.Context, arg CreateEvalParams) (Eval, error)
	GetEvalByGameAndSubscription(ctx context.Context, arg GetEvalByGameAndSubscriptionParams) (Eval, error)
	GetLatestEvalForSubscription(ctx context.Context, subscriptionID int64) (Eval, error)
	DeleteEvals(ctx context.Context, before time.Time) (int64, error)
	FindSubscriptionsWithExpiredNewestOnlineEval(ctx context.Context, before time.Time) ([]FindSubscriptionsWithExpiredNewestOnlineEvalRow, error)

	// Translations
	CreateTranslation(ctx context.Context, arg CreateTranslationParams) (Translation, error)
	GetTranslation(ctx context.Context, username string) (Translation, error)
	GetTranslations(ctx context.Context, usernames []string) ([]Translation, error)
	GetTranslationsForEval(ctx context.Context, evalID int64) ([]Translation, error)
	CreateTranslationToEval(ctx context.Context, arg CreateTranslationToEvalParams) error

	// Feedback
	CreateFeedback(ctx context.Context, arg CreateFeedbackParams) (Feedback, error)

	// Riot Account Cache
	GetCachedAccount(ctx context.Context, arg GetCachedAccountParams) (GetCachedAccountRow, error)
	CacheAccount(ctx context.Context, arg CacheAccountParams) error

	// Riot Game Cache
	GetCachedGameStatus(ctx context.Context, arg GetCachedGameStatusParams) (GetCachedGameStatusRow, error)
	CacheGameStatus(ctx context.Context, arg CacheGameStatusParams) error

	// Retention/Cleanup
	DeleteOldTranslations(ctx context.Context, before time.Time) (int64, error)
	DeleteOldFeedback(ctx context.Context, before time.Time) (int64, error)
	DeleteExpiredAccountCache(ctx context.Context) error
	DeleteExpiredGameCache(ctx context.Context) error

	// Players
	UpsertPlayer(ctx context.Context, arg UpsertPlayerParams) (Player, error)
	GetPlayer(ctx context.Context, username string) (Player, error)
	ListAllPlayers(ctx context.Context) ([]Player, error)
	UpdatePlayerStats(ctx context.Context, arg UpdatePlayerStatsParams) error

	// Public Translations (companion website)
	UpsertPublicTranslation(ctx context.Context, arg UpsertPublicTranslationParams) (PublicTranslation, error)
	GetPublicTranslation(ctx context.Context, id int64) (PublicTranslation, error)
	GetPublicTranslationByUsername(ctx context.Context, username string) (PublicTranslation, error)
	ListPublicTranslationsNew(ctx context.Context, arg ListPublicTranslationsNewParams) ([]PublicTranslation, error)
	ListPublicTranslationsTop(ctx context.Context, arg ListPublicTranslationsTopParams) ([]PublicTranslation, error)
	CountPublicTranslations(ctx context.Context, arg CountPublicTranslationsParams) (int64, error)
	IncrementUpvotes(ctx context.Context, id int64) error
	DecrementUpvotes(ctx context.Context, id int64) error
	IncrementDownvotes(ctx context.Context, id int64) error
	DecrementDownvotes(ctx context.Context, id int64) error

	// Votes
	UpsertVote(ctx context.Context, arg UpsertVoteParams) (Vote, error)
	GetVote(ctx context.Context, arg GetVoteParams) (Vote, error)
	DeleteVote(ctx context.Context, arg DeleteVoteParams) (int64, error)
	CountVotesByIP(ctx context.Context, ipHash string) (int64, error)

	// Public Feedback
	CreatePublicFeedback(ctx context.Context, arg CreatePublicFeedbackParams) (PublicFeedback, error)
	ListPublicFeedback(ctx context.Context, arg ListPublicFeedbackParams) ([]ListPublicFeedbackRow, error)
	CountPublicFeedback(ctx context.Context) (int64, error)

	// Transaction support
	WithTx(ctx context.Context, fn func(repo Repository) error) error

	// Lifecycle
	Close() error
}
