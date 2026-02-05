# Bot Package Tests

This directory contains comprehensive unit tests for the bot package.

## Running Tests

```bash
# Run all tests
go test -v ./internal/bot

# Run with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./internal/bot

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

## Test Coverage

**Current Coverage: 40.8%**

### Tested Components

#### ✅ Database Operations
- `TestCleanupOldData` - Tests cleanup of old evaluations and subscriptions
  - Successful cleanup
  - Error handling

#### ✅ Message Handling
- `TestConsumeTranslationMessages` - Tests Discord message sending and database updates
  - Successful message send with transaction
  - Discord API error handling
  - Database transaction error handling

#### ✅ Game Evaluation
- `TestProduceForServer` - Tests game status checking and translation job creation
  - Player in game with translations needed
  - Player not in game
  - Invalid username format
  - Eval already exists (deduplication)

#### ✅ Command Handlers
- `TestHandleSubscribe` - Tests subscription creation
  - Successful subscription
  - Subscription limit reached
  - Invalid Riot account

- `TestHandleUnsubscribe` - Tests subscription deletion
  - Successful unsubscription
  - Subscription not found

- `TestHandleListForChannel` - Tests listing subscriptions
  - List multiple subscriptions
  - No subscriptions

## Test Architecture

### Mock Implementations

The tests use comprehensive mock implementations for all bot dependencies:
- `MockLogger` - Logging interface
- `MockDiscordSession` - Discord API operations
- `MockRepository` - Database operations
- `MockRiotClient` - Riot API operations
- `MockTranslator` - Translation service

### Key Testing Patterns

1. **Dependency Injection**: All tests use the interface-based Bot constructor
2. **Mocking**: testify/mock for predictable test behavior
3. **Assertions**: testify/assert for clear test failures
4. **Coverage**: Tests cover happy paths and error scenarios

## Coverage Details

High coverage areas:
- `handleSubscribe`: 82.1%
- `handleUnsubscribe`: 88.2%
- `handleListForChannel`: 91.7%
- `consumeTranslationMessages`: 87.5%
- `produceForServer`: 83.3%

Areas for improvement:
- Integration testing for full bot lifecycle
- Event handler testing (handleInteraction, handleCommand)
- Producer/consumer goroutine coordination
- Discord interaction response handling
