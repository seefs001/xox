package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xtelebot"
)

func main() {
	xenv.Load()
	// Get the Telegram Bot Token from environment variables
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	// Create a new Bot instance with custom options
	bot, err := xtelebot.NewBot(token,
		xtelebot.WithRequestTimeout(10*time.Second),
		xtelebot.WithErrorHandler(customErrorHandler),
		xtelebot.WithBaseURL("https://api.telegram.org"),
	)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create a background context
	ctx := context.Background()

	// Initialize update offset
	offset := 0

	xlog.Info("Bot started successfully")

	// Set bot commands
	setCommands(ctx, bot)

	// Add this new function call
	getBotInfo(ctx, bot)

	// Create a channel to receive updates
	updatesChan, err := bot.GetUpdatesChan(ctx, offset, 100)
	if err != nil {
		xlog.Error("Failed to get updates channel", "error", err)
		return
	}

	// Listen for updates
	for update := range updatesChan {
		if update.Message != nil {
			handleMessage(ctx, bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(ctx, bot, update.CallbackQuery)
		}
		offset = update.UpdateID + 1
	}
}

func setCommands(ctx context.Context, bot *xtelebot.Bot) {
	commands := []xtelebot.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "help", Description: "Get help"},
		{Command: "photo", Description: "Receive a sample photo"},
		{Command: "location", Description: "Get a sample location"},
		{Command: "keyboard", Description: "See an inline keyboard"},
		{Command: "echo", Description: "Echo back your message"},
		{Command: "deletecommands", Description: "Delete bot commands"},
	}

	params := url.Values{}
	jsonCommands, _ := json.Marshal(commands)
	params.Set(xtelebot.ParamCommands, string(jsonCommands))

	_, err := bot.APIRequest(ctx, xtelebot.MethodSetMyCommands, params)
	if err != nil {
		xlog.Error("Failed to set bot commands", "error", err)
	} else {
		xlog.Info("Bot commands set successfully")
	}
}

func getUpdates(ctx context.Context, bot *xtelebot.Bot, offset int) ([]xtelebot.Update, error) {
	params := url.Values{}
	params.Set(xtelebot.ParamOffset, strconv.Itoa(offset))
	params.Set(xtelebot.ParamLimit, "100")

	body, err := bot.APIRequest(ctx, xtelebot.MethodGetUpdates, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool              `json:"ok"`
		Result []xtelebot.Update `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !resp.Ok {
		return nil, fmt.Errorf("API response not OK")
	}

	return resp.Result, nil
}

func handleMessage(ctx context.Context, bot *xtelebot.Bot, message *xtelebot.Message) {
	xlog.Info("Received message", "from", message.From.Username, "text", message.Text)
	xlog.Info(x.MustToJSON(message))

	switch {
	case strings.HasPrefix(message.Text, "/start"):
		sendWelcomeMessage(ctx, bot, message.Chat.ID)
	case strings.HasPrefix(message.Text, "/hello"):
		sendGreeting(ctx, bot, message.Chat.ID, message.From.FirstName)
	case strings.HasPrefix(message.Text, "/photo"):
		sendPhoto(ctx, bot, message.Chat.ID)
	case strings.HasPrefix(message.Text, "/location"):
		sendLocation(ctx, bot, message.Chat.ID)
	case strings.HasPrefix(message.Text, "/keyboard"):
		sendInlineKeyboard(ctx, bot, message.Chat.ID)
	case strings.HasPrefix(message.Text, "/echo"):
		echoMessage(ctx, bot, message)
	case strings.HasPrefix(message.Text, "/deletecommands"):
		deleteCommands(ctx, bot, message.Chat.ID)
	default:
		sendDefaultResponse(ctx, bot, message.Chat.ID)
	}
}

func sendWelcomeMessage(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	welcomeMessage := `Welcome to the example bot! Here are available commands:
/hello - Get a personalized greeting
/photo - Receive a sample photo
/location - Get a sample location
/keyboard - See an inline keyboard
/echo <text> - Echo back your message`

	sendMessage(ctx, bot, chatID, welcomeMessage)
}

func sendGreeting(ctx context.Context, bot *xtelebot.Bot, chatID interface{}, name string) {
	greeting := fmt.Sprintf("Hello, %s! How are you today?", name)
	sendMessage(ctx, bot, chatID, greeting)
}

func sendPhoto(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	photoURL := "https://www.google.com/images/branding/googlelogo/2x/googlelogo_light_color_272x92dp.png"
	params := url.Values{}
	params.Set(xtelebot.ParamChatID, fmt.Sprintf("%v", chatID))
	params.Set(xtelebot.ParamPhoto, photoURL)
	params.Set(xtelebot.ParamParseMode, "HTML")

	_, err := bot.APIRequest(ctx, xtelebot.MethodSendPhoto, params)
	if err != nil {
		xlog.Error("Failed to send photo", "error", err)
	}
}

func sendLocation(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	// Coordinates for New York Times Square
	params := url.Values{}
	params.Set(xtelebot.ParamChatID, fmt.Sprintf("%v", chatID))
	params.Set(xtelebot.ParamLatitude, "40.758896")
	params.Set(xtelebot.ParamLongitude, "-73.985130")

	_, err := bot.APIRequest(ctx, xtelebot.MethodSendLocation, params)
	if err != nil {
		xlog.Error("Failed to send location", "error", err)
	}
}

func sendInlineKeyboard(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	keyboard := xtelebot.NewInlineKeyboardMarkup(
		[]xtelebot.InlineKeyboardButton{
			xtelebot.NewInlineKeyboardButton("Say Hello", "", "hello"),
			xtelebot.NewInlineKeyboardButton("Get Time", "", "time"),
		},
		[]xtelebot.InlineKeyboardButton{
			xtelebot.NewInlineKeyboardButton("Send Location", "", "location"),
		},
	)

	params := url.Values{}
	params.Set(xtelebot.ParamChatID, fmt.Sprintf("%v", chatID))
	params.Set(xtelebot.ParamText, "Here's an inline keyboard:")
	params.Set(xtelebot.ParamReplyMarkup, string(x.MustToJSON(keyboard)))

	_, err := bot.APIRequest(ctx, xtelebot.MethodSendMessage, params)
	if err != nil {
		xlog.Error("Failed to send message with keyboard", "error", err)
	}
}

func echoMessage(ctx context.Context, bot *xtelebot.Bot, message *xtelebot.Message) {
	echoText := strings.TrimPrefix(message.Text, "/echo ")
	if echoText == "" {
		echoText = "You didn't provide any text to echo!"
	}
	sendMessage(ctx, bot, message.Chat.ID, echoText)
}

func sendDefaultResponse(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	sendMessage(ctx, bot, chatID, "I don't understand that command. Try /start for a list of available commands.")
}

func handleCallbackQuery(ctx context.Context, bot *xtelebot.Bot, query *xtelebot.CallbackQuery) {
	var text string
	switch query.Data {
	case "hello":
		text = "Hello! You clicked the 'Say Hello' button."
	case "time":
		text = fmt.Sprintf("The current time is: %s", time.Now().Format(time.RFC3339))
	case "location":
		// Send a location as a new message
		params := url.Values{}
		params.Set(xtelebot.ParamChatID, fmt.Sprintf("%v", query.Message.Chat.ID))
		params.Set(xtelebot.ParamLatitude, "40.758896")
		params.Set(xtelebot.ParamLongitude, "-73.985130")

		_, err := bot.APIRequest(ctx, xtelebot.MethodSendLocation, params)
		if err != nil {
			xlog.Error("Failed to send location", "error", err)
		}
		text = "I've sent you the location of New York Times Square."
	default:
		text = "Unknown button clicked."
	}

	// Answer the callback query
	params := url.Values{}
	params.Set(xtelebot.ParamCallbackQueryID, query.ID)
	params.Set(xtelebot.ParamText, text)

	_, err := bot.APIRequest(ctx, xtelebot.MethodAnswerCallbackQuery, params)
	if err != nil {
		xlog.Error("Failed to answer callback query", "error", err)
	}

	// Update the original message to reflect the action taken
	sendMessage(ctx, bot, query.Message.Chat.ID, text)
}

func customErrorHandler(err error) {
	xlog.Error("Bot error occurred", "error", err)
}

func deleteCommands(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	_, err := bot.APIRequest(ctx, xtelebot.MethodDeleteMyCommands, nil)
	if err != nil {
		xlog.Error("Failed to delete bot commands", "error", err)
		sendMessage(ctx, bot, chatID, "Failed to delete bot commands")
	} else {
		xlog.Info("Bot commands deleted successfully")
		sendMessage(ctx, bot, chatID, "Bot commands deleted successfully")
	}
}

func sendMessage(ctx context.Context, bot *xtelebot.Bot, chatID interface{}, text string) {
	params := url.Values{}
	params.Set(xtelebot.ParamChatID, fmt.Sprintf("%v", chatID))
	params.Set(xtelebot.ParamText, text)

	_, err := bot.APIRequest(ctx, xtelebot.MethodSendMessage, params)
	if err != nil {
		xlog.Error("Failed to send message", "error", err)
	}
}

func getBotInfo(ctx context.Context, bot *xtelebot.Bot) {
	botInfo, err := bot.GetMe(ctx)
	if err != nil {
		xlog.Error("Failed to get bot information", "error", err)
		return
	}

	xlog.Info("Bot information retrieved successfully",
		"id", botInfo.ID,
		"name", botInfo.FirstName,
		"username", botInfo.Username,
		"can_join_groups", botInfo.CanJoinGroups,
		"can_read_all_group_messages", botInfo.CanReadAllGroupMessages,
		"supports_inline_queries", botInfo.SupportsInlineQueries,
	)
}
