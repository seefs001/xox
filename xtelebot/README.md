# xtelebot

xtelebot is a Go package for building Telegram bots. It provides a simple and flexible API to interact with the Telegram Bot API.

## Features

- Easy-to-use API for common Telegram bot operations
- Support for long polling and webhooks
- Rate limiting to comply with Telegram's API limits
- Customizable HTTP client
- Debug mode for easier development

## Installation

```bash
go get github.com/seefs001/xox/xtelebot
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/seefs001/xox/xtelebot"
)

func main() {
    bot, err := xtelebot.NewBot("YOUR_BOT_TOKEN")
    if err != nil {
        log.Fatalf("Failed to create bot: %v", err)
    }

    updates, err := bot.GetUpdatesChan(context.Background(), 0, 100)
    if err != nil {
        log.Fatalf("Failed to get updates channel: %v", err)
    }

    for update := range updates {
        if update.Message != nil {
            _, err := bot.SendMessage(context.Background(), update.Message.Chat.ID, "Hello, I received your message!")
            if err != nil {
                log.Printf("Failed to send message: %v", err)
            }
        }
    }
}
```

## API Reference

### Creating a Bot

```go
bot, err := xtelebot.NewBot(token string, options ...BotOption)
```

Options:
- `WithBaseURL(baseURL string)`: Set a custom base URL for the Telegram API
- `WithHTTPClient(client *xhttpc.Client)`: Set a custom HTTP client
- `WithErrorHandler(handler ErrorHandler)`: Set a custom error handler
- `WithRequestTimeout(timeout time.Duration)`: Set a custom timeout for API requests
- `WithRateLimit(requestsPerSecond float64, burstSize int)`: Set a rate limit for API requests
- `WithDebug(debug bool)`: Enable or disable debug mode

### Sending Messages

```go
message, err := bot.SendMessage(ctx context.Context, chatID interface{}, text string, options ...MessageOption)
```

Options:
- `WithParseMode(mode string)`: Set the parse mode for the message
- `WithReplyMarkup(markup interface{})`: Add a reply markup to the message

Example:
```go
keyboard := xtelebot.NewInlineKeyboardMarkup(
    []xtelebot.InlineKeyboardButton{
        xtelebot.NewInlineKeyboardButton("Button 1", "", "callback_data_1"),
        xtelebot.NewInlineKeyboardButton("Button 2", "https://example.com", ""),
    },
)

message, err := bot.SendMessage(ctx, chatID, "Hello, World!",
    xtelebot.WithParseMode("HTML"),
    xtelebot.WithReplyMarkup(keyboard),
)
```

### Sending Photos

```go
message, err := bot.SendPhoto(ctx context.Context, chatID interface{}, photo interface{}, options ...MessageOption)
```

The `photo` parameter can be either a file ID string or an `io.Reader`.

### Sending Documents

```go
message, err := bot.SendDocument(ctx context.Context, chatID interface{}, document interface{}, options ...MessageOption)
```

The `document` parameter can be either a file ID string or an `io.Reader`.

### Sending Location

```go
message, err := bot.SendLocation(ctx context.Context, chatID interface{}, latitude, longitude float64, options ...MessageOption)
```

### Answering Callback Queries

```go
err := bot.AnswerCallbackQuery(ctx context.Context, callbackQueryID string, options ...CallbackQueryOption)
```

Options:
- `WithCallbackQueryText(text string)`: Set the text to be shown to the user
- `WithCallbackQueryShowAlert(showAlert bool)`: Set whether to show an alert instead of a notification

### Getting Updates

```go
updates, err := bot.GetUpdates(ctx context.Context, offset int, limit int)
```

### Using Update Channel

```go
updatesChan, err := bot.GetUpdatesChan(ctx context.Context, offset int, limit int)
```

### Working with Webhooks

```go
err := bot.SetWebhook(ctx context.Context, webhookURL string, options ...WebhookOption)
err := bot.DeleteWebhook(ctx context.Context)
webhookInfo, err := bot.GetWebhookInfo(ctx context.Context)
```

Webhook options:
- `WithMaxConnections(max int)`: Set the maximum allowed number of simultaneous HTTPS connections to the webhook
- `WithAllowedUpdates(types []string)`: Set a list of the update types you want your bot to receive
- `WithCertificate(cert string)`: Set a self-signed certificate for webhook

### Managing Bot Commands

```go
err := bot.SetMyCommands(ctx context.Context, commands []BotCommand)
err := bot.DeleteMyCommands(ctx context.Context)
commands, err := bot.GetMyCommands(ctx context.Context)
```

### Editing Messages

```go
message, err := bot.EditMessageText(ctx context.Context, options ...MessageOption)
message, err := bot.EditMessageCaption(ctx context.Context, options ...MessageOption)
message, err := bot.EditMessageMedia(ctx context.Context, options ...MessageOption)
```

## Error Handling

The package uses a default error handler that logs errors using `xlog.Error`. You can set a custom error handler using the `WithErrorHandler` option when creating a new bot.

## Rate Limiting

The package includes a built-in rate limiter to help comply with Telegram's API limits. You can configure the rate limit using the `WithRateLimit` option when creating a new bot.

## Debug Mode

You can enable debug mode using the `WithDebug` option when creating a new bot. This will log additional information about API requests and responses.

## Best Practices

1. Always use a context when making API calls to handle timeouts and cancellations.
2. Use rate limiting to avoid hitting Telegram's API limits.
3. Handle errors gracefully and consider implementing a custom error handler for better logging or recovery strategies.
4. Use webhook mode in production for better performance and scalability.
5. Implement proper security measures when using webhooks, such as HTTPS and secret tokens.
