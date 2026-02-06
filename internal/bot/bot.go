package bot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	MaxSubscriptionsPerServer    int64
	EvaluateSubscriptionsTimeout time.Duration
	EvalExpirationDuration       time.Duration
	OfflineActivityThreshold     time.Duration
	NumConsumers                 int64
	GuildID                      string
}

type Bot struct {
	log           Logger
	session       DiscordSession
	messageServer MessageServer
	repo          db.Repository
	riotClient    RiotClient
	translator    Translator
	config        Config
	rateLimiter   *RateLimiter
}

func New(
	log Logger,
	session DiscordSession,
	messageServer MessageServer,
	repo db.Repository,
	riotClient RiotClient,
	translator Translator,
	config Config,
) *Bot {
	return &Bot{
		log:           log,
		session:       session,
		messageServer: messageServer,
		repo:          repo,
		riotClient:    riotClient,
		translator:    translator,
		config:        config,
		rateLimiter:   NewRateLimiter(),
	}
}

// TODO: Support ignore lists
// TODO: metrics into grafana/lokiw
// TODO: When https://github.com/golangci/golangci-lint/pull/6271 merges, enable exhaustruct and errcheck in golangci-lint

func (b *Bot) Run(ctx context.Context, cancel context.CancelCauseFunc) error {
	b.session.AddHandler(b.handleInteraction)
	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		b.log.InfoContext(ctx, "connected to Discord", "username", r.User.Username, "discriminator", r.User.Discriminator)
	})

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("opening Discord connection: %w", err)
	}

	if err := b.registerCommands(ctx); err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}

	ch := make(chan sendMessageJob, 20)
	var wg sync.WaitGroup

	wg.Add(1)
	go b.runProducer(ctx, ch, &wg)

	b.log.InfoContext(ctx, "starting consumers", "count", b.config.NumConsumers)
	wg.Add(int(b.config.NumConsumers))
	for i := range b.config.NumConsumers {
		go b.runConsumer(ctx, ch, &wg, i)
	}

	wg.Add(1)
	go b.runCleaner(ctx, &wg)

	b.log.InfoContext(ctx, "bot is running, press Ctrl+C to stop")

	<-ctx.Done()
	b.log.Info("shutdown signal received")
	wg.Wait()
	b.session.Close()
	b.log.InfoContext(ctx, "shut down complete")

	return nil
}

func (b *Bot) registerCommands(ctx context.Context) error {
	guildID := b.config.GuildID
	if guildID != "" {
		b.log.InfoContext(ctx, "registering commands to guild", "guild_id", guildID)
		_, err := b.session.ApplicationCommandBulkOverwrite(b.session.GetUserID(), "", []*discordgo.ApplicationCommand{})
		if err != nil {
			b.log.WarnContext(ctx, "failed to clear global commands", "error", err)
		} else {
			b.log.InfoContext(ctx, "cleared global commands")
		}
	} else {
		b.log.InfoContext(ctx, "registering commands globally (may take up to 1 hour to propagate)")
	}

	_, err := b.session.ApplicationCommandBulkOverwrite(b.session.GetUserID(), guildID, commands)
	if err != nil {
		return fmt.Errorf("bulk overwrite commands: %w", err)
	}
	b.log.InfoContext(ctx, "registered commands", "count", len(commands))
	return nil
}

func (b *Bot) runProducer(ctx context.Context, ch chan<- sendMessageJob, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)
	for ctx.Err() == nil {
		produceCtx, cancel := context.WithTimeout(ctx, b.config.EvaluateSubscriptionsTimeout)
		jobs, err := b.produceTranslationMessages(produceCtx)
		// Best effort send jobs even if there's an error, just make sure to log it
		b.log.Info("produced jobs", slog.Int("num_jobs", len(jobs)))

		// Send jobs even if ctx cancelled - give them a chance to drain
		for _, job := range jobs {
			select {
			case <-time.After(time.Minute):
				// Timeout waiting to send - channel full and consumers stopped
				b.log.Warn("timeout sending job, dropping")
				cancel()
				return
			case ch <- job:
				b.log.InfoContext(ctx, "sent job", slog.String("channel_id", job.channelID), slog.Int64("game_id", job.gameID), slog.Int64("subscription_id", job.subscriptionID))
			}
		}
		cancel()

		if err != nil {
			b.log.ErrorContext(ctx, "running eval", "error", err)
		}

		sleepWithContext(ctx, time.Minute)
	}
}

func (b *Bot) runConsumer(ctx context.Context, ch <-chan sendMessageJob, wg *sync.WaitGroup, id int64) {
	defer wg.Done()
	log := b.log.With("consumer_id", id)
	for job := range ch {
		// Use Background so shutdown doesn't cancel in-flight work
		processCtx, cancel := context.WithTimeout(context.Background(),
			time.Minute)
		err := b.consumeTranslationMessages(processCtx, job)
		if err != nil {
			log.Error("consuming", "error", err)
		}
		cancel()
	}
	log.Info("consumer stopped")
}

func (b *Bot) runCleaner(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for ctx.Err() == nil {
		cleanupCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		err := b.cleanupOldData(cleanupCtx)
		cancel()
		if err != nil {
			b.log.Error("deleting old data", "error", err)
		}
		sleepWithContext(ctx, time.Hour)
	}
}

func sleepWithContext(ctx context.Context, dur time.Duration) {
	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		return
	}
}

func buildRegionChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(riot.ValidRegions))
	for i, region := range riot.ValidRegions {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  region,
			Value: region,
		}
	}
	return choices
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "subscribe",
		Description: "Subscribe to League of Legends summoner translations",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "Riot ID (e.g., name#tag)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "region",
				Description: "Server region",
				Required:    true,
				Choices:     buildRegionChoices(),
			},
		},
	},
	{
		Name:        "unsubscribe",
		Description: "Unsubscribe from a summoner",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "Riot ID (e.g., name#tag)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "region",
				Description: "Server region",
				Required:    true,
				Choices:     buildRegionChoices(),
			},
		},
	},
	{
		Name:        "list",
		Description: "List all subscriptions in this channel",
	},
}

type handlerResult struct {
	Response string
	Err      error
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponent(s, i)
	case discordgo.InteractionModalSubmit:
		b.handleModalSubmit(s, i)
	}
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if i.Member != nil && i.Member.User != nil && !b.rateLimiter.Allow(i.Member.User.ID) {
		b.respond(s, i, "\u26a0\ufe0f You're sending commands too fast. Please wait a moment.")
		return
	}

	var result handlerResult
	cmd := i.ApplicationCommandData().Name

	switch cmd {
	case "subscribe":
		result = b.handleSubscribe(i)
	case "unsubscribe":
		result = b.handleUnsubscribe(i)
	case "list":
		result = b.handleListForChannel(i)
	}

	b.respond(s, i, result.Response)

	if result.Err == nil {
		return
	}

	if _, ok := errors.AsType[*userError](result.Err); ok {
		if b.config.GuildID != "" {
			b.log.WarnContext(ctx, "user error", "command", cmd, "error", result.Err, "channel_id", i.ChannelID)
		}
	} else {
		b.log.ErrorContext(ctx, "command failed", "command", cmd, "error", result.Err, "channel_id", i.ChannelID)
	}
}

func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	customID := i.MessageComponentData().CustomID
	messageID := i.Message.ID

	switch customID {
	case "feedback_good":
		_, err := b.repo.CreateFeedback(ctx, db.CreateFeedbackParams{
			DiscordMessageID: messageID,
			FeedbackText:     "👍",
		})
		if err != nil {
			b.log.ErrorContext(ctx, "failed to store positive feedback", "error", err, "message_id", messageID)
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Thanks for the feedback!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	case "feedback_fix":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "feedback_modal:" + messageID,
				Title:    "Suggest a Correction",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "correction_text",
								Label:       "What should the translation be?",
								Style:       discordgo.TextInputParagraph,
								Placeholder: "e.g., 托儿索 should be 'Torso' not 'Yasuo wannabe'",
								Required:    true,
								MaxLength:   500,
							},
						},
					},
				},
			},
		})
	}
}

func (b *Bot) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	data := i.ModalSubmitData()

	parts := strings.Split(data.CustomID, ":")
	if len(parts) != 2 || parts[0] != "feedback_modal" {
		return
	}
	messageID := parts[1]

	var correctionText string
	for _, row := range data.Components {
		if actionsRow, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range actionsRow.Components {
				if input, ok := comp.(*discordgo.TextInput); ok && input.CustomID == "correction_text" {
					correctionText = input.Value
				}
			}
		}
	}

	_, err := b.repo.CreateFeedback(ctx, db.CreateFeedbackParams{
		DiscordMessageID: messageID,
		FeedbackText:     correctionText,
	})
	if err != nil {
		b.log.ErrorContext(ctx, "failed to store correction feedback", "error", err, "message_id", messageID)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Thanks! Your correction has been recorded.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

type userError struct {
	Err error
}

func (e *userError) Error() string {
	return e.Err.Error()
}

func (e *userError) Unwrap() error {
	return e.Err
}

func newUserError(err error) *userError {
	return &userError{Err: err}
}

func getOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}
	return ""
}

func (b *Bot) handleSubscribe(i *discordgo.InteractionCreate) handlerResult {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID
	serverID := i.GuildID
	// TODO: Probably need to handle this better, it's a shame that discordgo doesn't have context built into interactions
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	count, err := b.repo.CountSubscriptionsByServer(ctx, serverID)
	if count > b.config.MaxSubscriptionsPerServer {
		return handlerResult{
			Response: "❌ Already at maxium subscription count per server, please /unsubscribe to some before subscribing to more.",
		}
	}

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		return handlerResult{
			Response: "❌ Invalid Riot ID format. Use `name#tag` (e.g., `Faker#KR1`)",
			Err:      newUserError(err),
		}
	}

	if !riot.IsValidRegion(region) {
		return handlerResult{
			Response: fmt.Sprintf("❌ Invalid region: %s", region),
			Err:      newUserError(fmt.Errorf("invalid region: %s", region)),
		}
	}

	account, err := b.riotClient.GetAccountByRiotID(ctx, gameName, tagLine, region)
	if errors.Is(err, riot.ErrNotFound) {
		return handlerResult{
			Response: fmt.Sprintf("❌ Summoner **%s#%s** not found in **%s**", gameName, tagLine, region),
			Err:      newUserError(err),
		}
	}
	if err != nil {
		return handlerResult{
			Response: "❌ Failed to verify summoner. Please try again later.",
			Err:      fmt.Errorf("verify summoner %s in %s: %w", username, region, err),
		}
	}

	canonicalName := fmt.Sprintf("%s#%s", account.GameName, account.TagLine)

	_, err = b.repo.CreateSubscription(ctx, db.CreateSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
		ServerID:         serverID,
	})
	if db.IsNoRows(err) {
		return handlerResult{
			Response: fmt.Sprintf("⚠️ Already subscribed to **%s** (%s)", canonicalName, region),
			Err:      newUserError(err),
		}
	}
	if err != nil {
		return handlerResult{
			Response: "❌ Failed to subscribe. Please try again later.",
			Err:      fmt.Errorf("create subscription for %s in %s: %w", canonicalName, region, err),
		}
	}

	b.log.InfoContext(ctx, "subscription created", "username", canonicalName, "region", region, "channel_id", channelID)
	return handlerResult{Response: fmt.Sprintf("✅ Subscribed to **%s** (%s)! Will autounsubscribe after 3 weeks of no gameplay.", canonicalName, region)}
}

func (b *Bot) handleUnsubscribe(i *discordgo.InteractionCreate) handlerResult {
	options := i.ApplicationCommandData().Options
	username := getOption(options, "username")
	region := getOption(options, "region")
	channelID := i.ChannelID

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	gameName, tagLine, err := riot.ParseRiotID(username)
	if err != nil {
		return handlerResult{
			Response: "❌ Invalid Riot ID format. Use `name#tag`",
			Err:      newUserError(err),
		}
	}

	canonicalName := fmt.Sprintf("%s#%s", gameName, tagLine)

	rowsAffected, err := b.repo.DeleteSubscription(ctx, db.DeleteSubscriptionParams{
		DiscordChannelID: channelID,
		LolUsername:      canonicalName,
		Region:           region,
	})
	if err != nil {
		return handlerResult{
			Response: "❌ Failed to unsubscribe. Please try again later.",
			Err:      fmt.Errorf("delete subscription for %s in %s: %w", canonicalName, region, err),
		}
	}
	if rowsAffected == 0 {
		return handlerResult{
			Response: fmt.Sprintf("⚠️ No subscription found for **%s** (%s)", canonicalName, region),
			Err:      newUserError(fmt.Errorf("subscription not found: %s in %s", canonicalName, region)),
		}
	}

	b.log.InfoContext(ctx, "subscription deleted", "username", canonicalName, "region", region, "channel_id", channelID)
	return handlerResult{Response: fmt.Sprintf("✅ Unsubscribed from **%s** (%s)!", canonicalName, region)}
}

func (b *Bot) handleListForChannel(i *discordgo.InteractionCreate) handlerResult {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	channelID := i.ChannelID

	subs, err := b.repo.GetSubscriptionsByChannel(ctx, channelID)
	if err != nil {
		return handlerResult{
			Response: "❌ Failed to list subscriptions. Please try again later.",
			Err:      fmt.Errorf("list subscriptions: %w", err),
		}
	}

	if len(subs) == 0 {
		return handlerResult{Response: "No subscriptions in this channel. Use `/subscribe name#tag region` to add one!"}
	}

	content := "**Subscriptions in this channel:**\n"
	for _, sub := range subs {
		content += fmt.Sprintf("• %s (%s)\n", sub.LolUsername, sub.Region)
	}
	return handlerResult{Response: content}
}

func (b *Bot) respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		b.log.ErrorContext(ctx, "failed to respond to interaction", "error", err)
	}
}

type sendMessageJob struct {
	username       string
	translations   []translation.Translation
	subscriptionID int64
	channelID      string
	gameID         int64
}

func (b *Bot) produceForServer(ctx context.Context, subs []db.Subscription) ([]sendMessageJob, error) {
	var jobs []sendMessageJob
	var eg errgroup.Group
	for _, sub := range subs {
		eg.Go(func() error {
			username, tag, err := riot.ParseRiotID(sub.LolUsername)
			if err != nil {
				return fmt.Errorf("parsing riot id: %w", err)
			}

			acc, err := b.riotClient.GetAccountByRiotID(ctx, username, tag, sub.Region)
			if err != nil {
				return fmt.Errorf("getting account by riot id: %w", err)
			}

			game, err := b.riotClient.GetActiveGame(ctx, acc.PUUID, sub.Region)
			if errors.Is(err, riot.ErrNotInGame) {
				b.log.InfoContext(ctx, "user not in game", "username", sub.LolUsername, "region", sub.Region)
				return nil
			}
			if err != nil {
				return fmt.Errorf("getting active game %w", err)
			}

			_, err = b.repo.GetEvalByGameAndSubscription(ctx,
				db.GetEvalByGameAndSubscriptionParams{
					GameID:         sql.NullInt64{Int64: game.GameID, Valid: true},
					SubscriptionID: sub.ID,
				})

			// TODO: Ignore games that were already seen for a specific channel. This can happen if there's
			// 2 subs, 1 for each player, for a 2-premade. Or if they just encounter each other on the rift.
			// This is a little annoying so I'll put this off until people complain about it.
			if !db.IsNoRows(err) {
				b.log.InfoContext(ctx, "game already evaluated for this subscription", "subscription_id", sub.ID, "game_id", game.GameID)
				return nil
			}

			var names []string
			for _, p := range game.Participants {
				if !containsForeignCharacters(p.GameName) {
					continue
				}

				// Ignore self
				if p.GameName == sub.LolUsername {
					continue
				}

				// Don't best effort within a game because missing some translations seems sloppy and is bad UX.
				name, _, err := riot.ParseRiotID(p.GameName)
				if err != nil {
					return fmt.Errorf("unable to parse riot id %s: %w", p.GameName, err)
				}
				names = append(names, name)
			}

			if len(names) == 0 {
				b.log.InfoContext(ctx, "no foreign character names in game", "subscription_id", sub.ID, "game_id", game.GameID, "names", game.Participants)
				return nil
			}

			translations, err := b.translator.TranslateUsernames(ctx, names)
			if err != nil {
				return nil
			}

			jobs = append(jobs, sendMessageJob{
				username:       sub.LolUsername,
				translations:   translations,
				channelID:      sub.DiscordChannelID,
				subscriptionID: sub.ID,
				gameID:         game.GameID,
			})
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		// Return jobs for best effort
		return jobs, fmt.Errorf("producing for all channels in server: %w", err)
	}

	return jobs, nil
}

func (b *Bot) consumeTranslationMessages(ctx context.Context, job sendMessageJob) error {
	msg, err := b.messageServer.SendMessage(ctx, job)
	if err != nil {
		return fmt.Errorf("sending discord message: %w", err)
	}

	// All or nothing because we don't want either the eval or the denormalized subscription field without the other
	// since it's an invariant violation.
	err = b.repo.WithTx(ctx, func(txRepo db.Repository) error {
		var txErr error
		_, txErr = txRepo.CreateEval(ctx, db.CreateEvalParams{
			SubscriptionID:   job.subscriptionID,
			EvalStatus:       "NEW_TRANSLATIONS",
			DiscordMessageID: sql.NullString{String: msg.ID, Valid: true},
			GameID:           sql.NullInt64{Int64: job.gameID, Valid: true},
		})
		if txErr != nil {
			return fmt.Errorf("creating eval record: %w", txErr)
		}

		txErr = txRepo.UpdateSubscriptionLastEvaluatedAt(ctx, job.subscriptionID)
		if txErr != nil {
			return fmt.Errorf("updating subscription last evaluated at: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return err
	}

	b.log.InfoContext(ctx, "sent and processed translation message",
		"subscription_id", job.subscriptionID,
		"channel_id", job.channelID,
		"game_id", job.gameID,
	)

	return nil
}

func (b *Bot) produceTranslationMessages(ctx context.Context) ([]sendMessageJob, error) {
	b.log.InfoContext(ctx, "starting eval loop...")
	subs, err := b.repo.GetAllSubscriptions(ctx, 1000)
	if err != nil {
		return nil, fmt.Errorf("getting all subscriptions %w", err)
	}
	b.log.InfoContext(ctx, "subscriptions", "subs", subs, "err", err)

	servers := lo.GroupBy(subs, func(s db.Subscription) string {
		return s.ServerID
	})

	var eg errgroup.Group
	var mu sync.Mutex
	var jobs []sendMessageJob

	for server, subs := range servers {
		eg.Go(func() error {
			serverJobs, err := b.produceForServer(ctx, subs)
			mu.Lock()
			jobs = append(jobs, serverJobs...)
			mu.Unlock()
			// Best effort
			if err != nil {
				return fmt.Errorf("producing for server %s: %w", server, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		err = fmt.Errorf("producing translation messages: %w", err)
	}

	return jobs, err
}

// TODO: Only supports korean and chinese so far, make the language a user-passed flag.
func containsForeignCharacters(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

func (b *Bot) cleanupOldData(ctx context.Context) error {
	log := b.log.With("subsystem", "cleanup_old_data")
	rows, err := b.repo.DeleteEvals(ctx, time.Now().Add(-b.config.EvalExpirationDuration))
	if err != nil {
		return fmt.Errorf("deleting old evals: %w", err)
	}
	log.InfoContext(ctx, "Deleted rows", slog.Int64("rows", rows))

	subs, err := b.repo.FindSubscriptionsWithExpiredNewestOnlineEval(ctx, time.Now().Add(-b.config.OfflineActivityThreshold))
	if len(subs) == 0 {
		log.InfoContext(ctx, "No expired subs")
		return nil
	}
	if err != nil {
		return fmt.Errorf("retrieving expired subscriptions: %w", err)
	}

	subIds := lo.Map(subs, func(s db.FindSubscriptionsWithExpiredNewestOnlineEvalRow, _ int) int64 {
		return s.SubscriptionID
	})
	count, err := b.repo.DeleteSubscriptions(ctx, subIds)
	if err != nil {
		return fmt.Errorf("deleting expired subs: %w", err)
	}
	log.InfoContext(ctx, "deleted expired subs", slog.Int64("deleted_subs_count", count))

	return nil
}
