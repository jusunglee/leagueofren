package bot

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
)

// Logger defines the logging interface used by Bot
type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// DiscordSession defines the Discord session interface used by Bot
type DiscordSession interface {
	AddHandler(handler interface{}) func()
	Open() error
	Close() error
	ApplicationCommandBulkOverwrite(appID, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
	// GetUserID returns the bot's user ID
	GetUserID() string
}

// RiotClient defines the Riot API client interface used by Bot
type RiotClient interface {
	GetAccountByRiotID(ctx context.Context, gameName, tagLine, region string) (riot.Account, error)
	GetActiveGame(ctx context.Context, puuid, region string) (riot.ActiveGame, error)
}

// Translator defines the translation interface used by Bot
type Translator interface {
	TranslateUsernames(ctx context.Context, usernames []string) ([]translation.Translation, error)
}

// slogAdapter wraps *slog.Logger to return our Logger interface from With()
type slogAdapter struct {
	*slog.Logger
}

func (l *slogAdapter) With(args ...any) Logger {
	return &slogAdapter{Logger: l.Logger.With(args...)}
}

// NewLogger wraps a *slog.Logger to implement the Logger interface
func NewLogger(log *slog.Logger) Logger {
	return &slogAdapter{Logger: log}
}

// discordSessionAdapter wraps *discordgo.Session to implement DiscordSession
type discordSessionAdapter struct {
	*discordgo.Session
}

func (s *discordSessionAdapter) GetUserID() string {
	return s.State.User.ID
}

// NewDiscordSession wraps a *discordgo.Session to implement the DiscordSession interface
func NewDiscordSession(session *discordgo.Session) DiscordSession {
	return &discordSessionAdapter{Session: session}
}

// riotClientAdapter wraps *riot.CachedClient to implement RiotClient
type riotClientAdapter struct {
	*riot.CachedClient
}

// NewRiotClient wraps a *riot.CachedClient to implement the RiotClient interface
func NewRiotClient(client *riot.CachedClient) RiotClient {
	return &riotClientAdapter{CachedClient: client}
}

// translatorAdapter wraps *translation.Translator to implement Translator
type translatorAdapter struct {
	*translation.Translator
}

// NewTranslator wraps a *translation.Translator to implement the Translator interface
func NewTranslator(translator *translation.Translator) Translator {
	return &translatorAdapter{Translator: translator}
}
