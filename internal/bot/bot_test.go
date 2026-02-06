package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) With(args ...any) Logger {
	ret := m.Called(args)
	return ret.Get(0).(Logger)
}

type MockDiscordSession struct {
	mock.Mock
}

func (m *MockDiscordSession) AddHandler(handler interface{}) func() {
	ret := m.Called(handler)
	return ret.Get(0).(func())
}

func (m *MockDiscordSession) Open() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDiscordSession) Close() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDiscordSession) ApplicationCommandBulkOverwrite(appID, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	ret := m.Called(appID, guildID, commands, options)
	return ret.Get(0).([]*discordgo.ApplicationCommand), ret.Error(1)
}

func (m *MockDiscordSession) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	ret := m.Called(channelID, data, options)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*discordgo.Message), ret.Error(1)
}

func (m *MockDiscordSession) GetUserID() string {
	ret := m.Called()
	return ret.String(0)
}

func (m *MockDiscordSession) UserChannelPermissions(userID, channelID string, options ...discordgo.RequestOption) (int64, error) {
	ret := m.Called(userID, channelID, options)
	return ret.Get(0).(int64), ret.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateSubscription(ctx context.Context, arg db.CreateSubscriptionParams) (db.Subscription, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.Subscription), ret.Error(1)
}

func (m *MockRepository) GetAllSubscriptions(ctx context.Context, limit int32) ([]db.Subscription, error) {
	ret := m.Called(ctx, limit)
	return ret.Get(0).([]db.Subscription), ret.Error(1)
}

func (m *MockRepository) GetSubscriptionsByChannel(ctx context.Context, discordChannelID string) ([]db.Subscription, error) {
	ret := m.Called(ctx, discordChannelID)
	return ret.Get(0).([]db.Subscription), ret.Error(1)
}

func (m *MockRepository) GetSubscriptionByID(ctx context.Context, id int64) (db.Subscription, error) {
	ret := m.Called(ctx, id)
	return ret.Get(0).(db.Subscription), ret.Error(1)
}

func (m *MockRepository) CountSubscriptionsByServer(ctx context.Context, serverID string) (int64, error) {
	ret := m.Called(ctx, serverID)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRepository) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) (int64, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRepository) DeleteSubscriptions(ctx context.Context, ids []int64) (int64, error) {
	ret := m.Called(ctx, ids)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRepository) UpdateSubscriptionLastEvaluatedAt(ctx context.Context, id int64) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockRepository) CreateEval(ctx context.Context, arg db.CreateEvalParams) (db.Eval, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.Eval), ret.Error(1)
}

func (m *MockRepository) GetEvalByGameAndSubscription(ctx context.Context, arg db.GetEvalByGameAndSubscriptionParams) (db.Eval, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.Eval), ret.Error(1)
}

func (m *MockRepository) GetLatestEvalForSubscription(ctx context.Context, subscriptionID int64) (db.Eval, error) {
	ret := m.Called(ctx, subscriptionID)
	return ret.Get(0).(db.Eval), ret.Error(1)
}

func (m *MockRepository) DeleteEvals(ctx context.Context, before time.Time) (int64, error) {
	ret := m.Called(ctx, before)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRepository) FindSubscriptionsWithExpiredNewestOnlineEval(ctx context.Context, before time.Time) ([]db.FindSubscriptionsWithExpiredNewestOnlineEvalRow, error) {
	ret := m.Called(ctx, before)
	return ret.Get(0).([]db.FindSubscriptionsWithExpiredNewestOnlineEvalRow), ret.Error(1)
}

func (m *MockRepository) CreateTranslation(ctx context.Context, arg db.CreateTranslationParams) (db.Translation, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.Translation), ret.Error(1)
}

func (m *MockRepository) GetTranslation(ctx context.Context, username string) (db.Translation, error) {
	ret := m.Called(ctx, username)
	return ret.Get(0).(db.Translation), ret.Error(1)
}

func (m *MockRepository) GetTranslations(ctx context.Context, usernames []string) ([]db.Translation, error) {
	ret := m.Called(ctx, usernames)
	return ret.Get(0).([]db.Translation), ret.Error(1)
}

func (m *MockRepository) GetTranslationsForEval(ctx context.Context, evalID int64) ([]db.Translation, error) {
	ret := m.Called(ctx, evalID)
	return ret.Get(0).([]db.Translation), ret.Error(1)
}

func (m *MockRepository) CreateTranslationToEval(ctx context.Context, arg db.CreateTranslationToEvalParams) error {
	ret := m.Called(ctx, arg)
	return ret.Error(0)
}

func (m *MockRepository) CreateFeedback(ctx context.Context, arg db.CreateFeedbackParams) (db.Feedback, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.Feedback), ret.Error(1)
}

func (m *MockRepository) GetCachedAccount(ctx context.Context, arg db.GetCachedAccountParams) (db.GetCachedAccountRow, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.GetCachedAccountRow), ret.Error(1)
}

func (m *MockRepository) CacheAccount(ctx context.Context, arg db.CacheAccountParams) error {
	ret := m.Called(ctx, arg)
	return ret.Error(0)
}

func (m *MockRepository) GetCachedGameStatus(ctx context.Context, arg db.GetCachedGameStatusParams) (db.GetCachedGameStatusRow, error) {
	ret := m.Called(ctx, arg)
	return ret.Get(0).(db.GetCachedGameStatusRow), ret.Error(1)
}

func (m *MockRepository) CacheGameStatus(ctx context.Context, arg db.CacheGameStatusParams) error {
	ret := m.Called(ctx, arg)
	return ret.Error(0)
}

func (m *MockRepository) DeleteExpiredAccountCache(ctx context.Context) error {
	ret := m.Called(ctx)
	return ret.Error(0)
}

func (m *MockRepository) DeleteExpiredGameCache(ctx context.Context) error {
	ret := m.Called(ctx)
	return ret.Error(0)
}

func (m *MockRepository) WithTx(ctx context.Context, fn func(repo db.Repository) error) error {
	ret := m.Called(ctx, fn)
	if ret.Error(0) == nil {
		return fn(m)
	}
	return ret.Error(0)
}

func (m *MockRepository) Close() error {
	ret := m.Called()
	return ret.Error(0)
}

type MockRiotClient struct {
	mock.Mock
}

func (m *MockRiotClient) GetAccountByRiotID(ctx context.Context, gameName, tagLine, region string) (riot.Account, error) {
	ret := m.Called(ctx, gameName, tagLine, region)
	return ret.Get(0).(riot.Account), ret.Error(1)
}

func (m *MockRiotClient) GetActiveGame(ctx context.Context, puuid, region string) (riot.ActiveGame, error) {
	ret := m.Called(ctx, puuid, region)
	return ret.Get(0).(riot.ActiveGame), ret.Error(1)
}

type MockTranslator struct {
	mock.Mock
}

func (m *MockTranslator) TranslateUsernames(ctx context.Context, usernames []string) ([]translation.Translation, error) {
	ret := m.Called(ctx, usernames)
	return ret.Get(0).([]translation.Translation), ret.Error(1)
}

type MockMessageServer struct {
	mock.Mock
}

func (m *MockMessageServer) SendMessage(ctx context.Context, job sendMessageJob) (*discordgo.Message, error) {
	ret := m.Called(ctx, job)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*discordgo.Message), ret.Error(1)
}

func (m *MockMessageServer) ReplyToMessage(channelID, messageID, content string) error {
	ret := m.Called(channelID, messageID, content)
	return ret.Error(0)
}

// Helper function to create a test bot
func newTestBot(
	log Logger,
	session DiscordSession,
	messageServer MessageServer,
	repo db.Repository,
	riotClient RiotClient,
	translator Translator,
) *Bot {
	return New(log, session, messageServer, repo, riotClient, translator, Config{
		MaxSubscriptionsPerServer:    10,
		EvaluateSubscriptionsTimeout: time.Minute,
		EvalExpirationDuration:       10 * time.Minute,
		OfflineActivityThreshold:     5 * time.Minute,
		NumConsumers:                 2,
		GuildID:                      "",
		JobBufferSize:                20,
	})
}

// Test cleanupOldData
func TestCleanupOldData(t *testing.T) {
	ctx := context.Background()

	t.Run("successful cleanup", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSubLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		mockLogger.On("With", mock.Anything).Return(mockSubLogger)
		mockSubLogger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()

		mockRepo.On("DeleteEvals", mock.Anything, mock.Anything).Return(int64(5), nil)
		mockRepo.On("FindSubscriptionsWithExpiredNewestOnlineEval", mock.Anything, mock.Anything).
			Return([]db.FindSubscriptionsWithExpiredNewestOnlineEvalRow{}, nil)

		err := bot.cleanupOldData(ctx)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("delete evals error", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSubLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		mockLogger.On("With", mock.Anything).Return(mockSubLogger)
		mockRepo.On("DeleteEvals", mock.Anything, mock.Anything).Return(int64(0), errors.New("db error"))

		err := bot.cleanupOldData(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting old evals")
		mockRepo.AssertExpectations(t)
	})
}

// Test consumeTranslationMessages
func TestConsumeTranslationMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("successful message send and db operations", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		job := sendMessageJob{
			username:       "TestUser",
			translations:   []translation.Translation{{Original: "테스트", Translated: "Test"}},
			channelID:      "channel-123",
			subscriptionID: 1,
			gameID:         999,
		}

		mockMessageServer.On("SendMessage", mock.Anything, mock.MatchedBy(func(j sendMessageJob) bool {
			return j.channelID == "channel-123" && j.subscriptionID == 1
		})).Return(&discordgo.Message{ID: "msg-456"}, nil)

		mockRepo.On("WithTx", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("CreateEval", mock.Anything, mock.MatchedBy(func(params db.CreateEvalParams) bool {
			return params.SubscriptionID == 1 &&
				params.EvalStatus == "NEW_TRANSLATIONS" &&
				params.DiscordMessageID.String == "msg-456" &&
				params.GameID.Int64 == 999
		})).Return(db.Eval{ID: 10}, nil)
		mockRepo.On("UpdateSubscriptionLastEvaluatedAt", mock.Anything, int64(1)).Return(nil)

		mockLogger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()

		err := bot.consumeTranslationMessages(ctx, job)
		require.NoError(t, err)
		mockMessageServer.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("discord send error", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		job := sendMessageJob{
			username:       "TestUser",
			translations:   []translation.Translation{{Original: "테스트", Translated: "Test"}},
			channelID:      "channel-123",
			subscriptionID: 1,
			gameID:         999,
		}

		mockMessageServer.On("SendMessage", mock.Anything, mock.MatchedBy(func(j sendMessageJob) bool {
			return j.channelID == "channel-123" && j.subscriptionID == 1
		})).Return(nil, errors.New("discord error"))

		err := bot.consumeTranslationMessages(ctx, job)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sending discord message")
		mockMessageServer.AssertExpectations(t)
	})

	t.Run("transaction error replies to orphaned message", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		job := sendMessageJob{
			username:       "TestUser",
			translations:   []translation.Translation{{Original: "테스트", Translated: "Test"}},
			channelID:      "channel-123",
			subscriptionID: 1,
			gameID:         999,
		}

		mockMessageServer.On("SendMessage", mock.Anything, mock.MatchedBy(func(j sendMessageJob) bool {
			return j.channelID == "channel-123" && j.subscriptionID == 1
		})).Return(&discordgo.Message{ID: "msg-456"}, nil)

		mockRepo.On("WithTx", mock.Anything, mock.Anything).Return(errors.New("tx error"))
		mockMessageServer.On("ReplyToMessage", "channel-123", "msg-456", mock.Anything).Return(nil)

		err := bot.consumeTranslationMessages(ctx, job)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tx error")
		mockMessageServer.AssertCalled(t, "ReplyToMessage", "channel-123", "msg-456", mock.Anything)
	})

	t.Run("transaction error with failed reply logs warning", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		job := sendMessageJob{
			username:       "TestUser",
			translations:   []translation.Translation{{Original: "테스트", Translated: "Test"}},
			channelID:      "channel-123",
			subscriptionID: 1,
			gameID:         999,
		}

		mockMessageServer.On("SendMessage", mock.Anything, mock.MatchedBy(func(j sendMessageJob) bool {
			return j.channelID == "channel-123" && j.subscriptionID == 1
		})).Return(&discordgo.Message{ID: "msg-456"}, nil)

		mockRepo.On("WithTx", mock.Anything, mock.Anything).Return(errors.New("tx error"))
		mockMessageServer.On("ReplyToMessage", "channel-123", "msg-456", mock.Anything).Return(errors.New("reply failed"))
		mockLogger.On("WarnContext", mock.Anything, "failed to reply to orphaned message after transaction failure", mock.Anything).Return()

		err := bot.consumeTranslationMessages(ctx, job)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tx error")
		mockLogger.AssertCalled(t, "WarnContext", mock.Anything, "failed to reply to orphaned message after transaction failure", mock.Anything)
	})
}

// Test produceForServer
func TestProduceForServer(t *testing.T) {
	ctx := context.Background()

	t.Run("player in game with translations", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		subs := []db.Subscription{
			{
				ID:               1,
				DiscordChannelID: "channel-123",
				ServerID:         "server-456",
				LolUsername:      "Player#NA1",
				Region:           "NA",
			},
		}

		mockRiot.On("GetAccountByRiotID", ctx, "Player", "NA1", "NA").
			Return(riot.Account{PUUID: "puuid-123", GameName: "Player", TagLine: "NA1"}, nil)

		mockRiot.On("GetActiveGame", ctx, "puuid-123", "NA").
			Return(riot.ActiveGame{
				GameID: 999,
				Participants: []riot.Participant{
					{GameName: "Player1#NA1"},
					{GameName: "玩家2#NA1"},
				},
			}, nil)

		mockRepo.On("GetEvalByGameAndSubscription", ctx, mock.MatchedBy(func(params db.GetEvalByGameAndSubscriptionParams) bool {
			return params.GameID.Int64 == 999 && params.SubscriptionID == 1
		})).Return(db.Eval{}, db.ErrNoRows)

		mockTranslator.On("TranslateUsernames", ctx, []string{"玩家2"}).
			Return([]translation.Translation{
				{Original: "玩家2", Translated: "Player 2"},
			}, nil)

		jobs, err := bot.produceForServer(ctx, subs)
		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, "channel-123", jobs[0].channelID)
		assert.Equal(t, int64(1), jobs[0].subscriptionID)
		assert.Equal(t, int64(999), jobs[0].gameID)

		mockRiot.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockTranslator.AssertExpectations(t)
	})

	t.Run("player not in game", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		subs := []db.Subscription{
			{
				ID:               1,
				DiscordChannelID: "channel-123",
				ServerID:         "server-456",
				LolUsername:      "Player#NA1",
				Region:           "NA",
			},
		}

		mockLogger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()

		mockRiot.On("GetAccountByRiotID", ctx, "Player", "NA1", "NA").
			Return(riot.Account{PUUID: "puuid-123", GameName: "Player", TagLine: "NA1"}, nil)

		mockRiot.On("GetActiveGame", ctx, "puuid-123", "NA").
			Return(riot.ActiveGame{}, riot.ErrNotInGame)

		jobs, err := bot.produceForServer(ctx, subs)
		require.NoError(t, err)
		assert.Len(t, jobs, 0)

		mockRiot.AssertExpectations(t)
	})

	t.Run("invalid username format", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		subs := []db.Subscription{
			{
				ID:               1,
				DiscordChannelID: "channel-123",
				ServerID:         "server-456",
				LolUsername:      "InvalidFormat",
				Region:           "NA",
			},
		}

		jobs, err := bot.produceForServer(ctx, subs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
		assert.Len(t, jobs, 0)
	})

	t.Run("eval already exists", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		subs := []db.Subscription{
			{
				ID:               1,
				DiscordChannelID: "channel-123",
				ServerID:         "server-456",
				LolUsername:      "Player#NA1",
				Region:           "NA",
			},
		}

		mockLogger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()

		mockRiot.On("GetAccountByRiotID", ctx, "Player", "NA1", "NA").
			Return(riot.Account{PUUID: "puuid-123", GameName: "Player", TagLine: "NA1"}, nil)

		mockRiot.On("GetActiveGame", ctx, "puuid-123", "NA").
			Return(riot.ActiveGame{
				GameID:       999,
				Participants: []riot.Participant{{GameName: "Player1#NA1"}},
			}, nil)

		mockRepo.On("GetEvalByGameAndSubscription", ctx, mock.Anything).
			Return(db.Eval{ID: 5}, nil)

		jobs, err := bot.produceForServer(ctx, subs)
		require.NoError(t, err)
		assert.Len(t, jobs, 0)

		mockRiot.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

// Test handleSubscribe
func TestHandleSubscribe(t *testing.T) {
	t.Run("successful subscription", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Name: "username", Type: discordgo.ApplicationCommandOptionString, Value: "Player#NA1"},
						{Name: "region", Type: discordgo.ApplicationCommandOptionString, Value: "NA"},
					},
				},
				GuildID:   "guild-123",
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("CountSubscriptionsByServer", mock.Anything, "guild-123").
			Return(int64(5), nil)

		mockRiot.On("GetAccountByRiotID", mock.Anything, "Player", "NA1", "NA").
			Return(riot.Account{PUUID: "puuid-123", GameName: "Player", TagLine: "NA1"}, nil)

		mockRepo.On("CreateSubscription", mock.Anything, mock.MatchedBy(func(params db.CreateSubscriptionParams) bool {
			return params.DiscordChannelID == "channel-456" &&
				params.LolUsername == "Player#NA1" &&
				params.Region == "NA" &&
				params.ServerID == "guild-123"
		})).Return(db.Subscription{ID: 1}, nil)

		mockLogger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()

		result := bot.handleSubscribe(interaction)
		assert.NoError(t, result.Err)
		assert.Contains(t, result.Response, "Subscribed")

		mockRiot.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("subscription limit reached", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Name: "username", Type: discordgo.ApplicationCommandOptionString, Value: "Player#NA1"},
						{Name: "region", Type: discordgo.ApplicationCommandOptionString, Value: "NA"},
					},
				},
				GuildID:   "guild-123",
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("CountSubscriptionsByServer", mock.Anything, "guild-123").
			Return(int64(11), nil)

		result := bot.handleSubscribe(interaction)
		assert.NoError(t, result.Err)
		assert.Contains(t, result.Response, "maxium subscription")
	})

	t.Run("invalid riot account", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Name: "username", Type: discordgo.ApplicationCommandOptionString, Value: "Invalid#NA1"},
						{Name: "region", Type: discordgo.ApplicationCommandOptionString, Value: "NA"},
					},
				},
				GuildID:   "guild-123",
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("CountSubscriptionsByServer", mock.Anything, "guild-123").
			Return(int64(5), nil)

		mockRiot.On("GetAccountByRiotID", mock.Anything, "Invalid", "NA1", "NA").
			Return(riot.Account{}, riot.ErrNotFound)

		result := bot.handleSubscribe(interaction)
		assert.Error(t, result.Err)
		var ue *userError
		assert.ErrorAs(t, result.Err, &ue)
		assert.Contains(t, result.Response, "not found")

		mockRiot.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

// Test handleUnsubscribe
func TestHandleUnsubscribe(t *testing.T) {
	t.Run("successful unsubscription", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Name: "username", Type: discordgo.ApplicationCommandOptionString, Value: "Player#NA1"},
						{Name: "region", Type: discordgo.ApplicationCommandOptionString, Value: "NA"},
					},
				},
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("DeleteSubscription", mock.Anything, mock.MatchedBy(func(params db.DeleteSubscriptionParams) bool {
			return params.DiscordChannelID == "channel-456" &&
				params.LolUsername == "Player#NA1" &&
				params.Region == "NA"
		})).Return(int64(1), nil)

		mockLogger.On("InfoContext", mock.Anything, mock.Anything, mock.Anything).Return()

		result := bot.handleUnsubscribe(interaction)
		assert.NoError(t, result.Err)
		assert.Contains(t, result.Response, "Unsubscribed")

		mockRepo.AssertExpectations(t)
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Name: "username", Type: discordgo.ApplicationCommandOptionString, Value: "Player#NA1"},
						{Name: "region", Type: discordgo.ApplicationCommandOptionString, Value: "NA"},
					},
				},
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("DeleteSubscription", mock.Anything, mock.Anything).
			Return(int64(0), nil)

		result := bot.handleUnsubscribe(interaction)
		assert.Error(t, result.Err)
		var ue *userError
		assert.ErrorAs(t, result.Err, &ue)
		assert.Contains(t, result.Response, "No subscription found")

		mockRepo.AssertExpectations(t)
	})
}

// Test handleListForChannel
func TestHandleListForChannel(t *testing.T) {
	t.Run("list subscriptions", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type:      discordgo.InteractionApplicationCommand,
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("GetSubscriptionsByChannel", mock.Anything, "channel-456").
			Return([]db.Subscription{
				{ID: 1, LolUsername: "Player1#NA1", Region: "NA"},
				{ID: 2, LolUsername: "Player2#EUW", Region: "euw1"},
			}, nil)

		result := bot.handleListForChannel(interaction)
		assert.NoError(t, result.Err)
		assert.Contains(t, result.Response, "Player1#NA1")
		assert.Contains(t, result.Response, "Player2#EUW")

		mockRepo.AssertExpectations(t)
	})

	t.Run("no subscriptions", func(t *testing.T) {
		mockLogger := new(MockLogger)
		mockSession := new(MockDiscordSession)
		mockMessageServer := new(MockMessageServer)
		mockRepo := new(MockRepository)
		mockRiot := new(MockRiotClient)
		mockTranslator := new(MockTranslator)

		bot := newTestBot(mockLogger, mockSession, mockMessageServer, mockRepo, mockRiot, mockTranslator)

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type:      discordgo.InteractionApplicationCommand,
				ChannelID: "channel-456",
			},
		}

		mockRepo.On("GetSubscriptionsByChannel", mock.Anything, "channel-456").
			Return([]db.Subscription{}, nil)

		result := bot.handleListForChannel(interaction)
		assert.NoError(t, result.Err)
		assert.Contains(t, result.Response, "No subscriptions")

		mockRepo.AssertExpectations(t)
	})
}
