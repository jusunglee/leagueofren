package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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
	// UserChannelPermissions returns the permissions a user has in a channel
	UserChannelPermissions(userID, channelID string, options ...discordgo.RequestOption) (int64, error)
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

// MessageServer defines the interface for sending messages to a messaging platform
type MessageServer interface {
	SendMessage(ctx context.Context, job sendMessageJob) (*discordgo.Message, error)
	ReplyToMessage(channelID, messageID, content string) error
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

// discordMessageServer implements MessageServer using a Discord session
type discordMessageServer struct {
	session DiscordSession
}

func (d *discordMessageServer) SendMessage(ctx context.Context, job sendMessageJob) (*discordgo.Message, error) {
	embed := formatTranslationEmbed(job.username, job.translations)
	return d.session.ChannelMessageSendComplex(job.channelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Good ✓",
						CustomID: "feedback_good",
						Style:    discordgo.SuccessButton,
					},
					discordgo.Button{
						Label:    "Suggest Fix",
						CustomID: "feedback_fix",
						Style:    discordgo.SecondaryButton,
					},
				},
			},
		},
	})
}

func (d *discordMessageServer) ReplyToMessage(channelID, messageID, content string) error {
	_, err := d.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:   content,
		Reference: &discordgo.MessageReference{MessageID: messageID},
	})
	return err
}

// NewMessageServer creates a MessageServer that uses Discord
func NewMessageServer(session DiscordSession) MessageServer {
	return &discordMessageServer{session: session}
}

func formatTranslationEmbed(username string, translations []translation.Translation) *discordgo.MessageEmbed {
	const maxInlineEntries = 8
	fields := make([]*discordgo.MessageEmbedField, 0, 25)

	inlineCount := min(len(translations), maxInlineEntries)
	for i := 0; i < inlineCount; i++ {
		t := translations[i]
		fields = append(fields,
			&discordgo.MessageEmbedField{Name: "Original", Value: t.Original, Inline: true},
			&discordgo.MessageEmbedField{Name: "Translation", Value: t.Translated, Inline: true},
		)
		if i < inlineCount-1 {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "\u200b", Value: "\u200b", Inline: false})
		}
	}

	if len(translations) > maxInlineEntries {
		var sb strings.Builder
		for _, t := range translations[maxInlineEntries:] {
			sb.WriteString(fmt.Sprintf("**%s** → %s\n", t.Original, t.Translated))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "\u200b",
			Value: sb.String(),
		})
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s is in a game!", username),
		Color:       0x5865F2,
		Description: "Translations for players in this match:",
		Fields:      fields,
	}
}
