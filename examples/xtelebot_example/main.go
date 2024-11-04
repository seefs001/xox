package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xtelebot"
)

// Handler is a function that handles a specific update
type Handler func(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update)

// Router manages the routing of updates to handlers
type Router struct {
	commandHandlers map[string]Handler
	textHandlers    []Handler
	privateHandlers []Handler
	groupHandlers   []Handler
	// Add more handler types as needed
}

// NewRouter creates a new Router
func NewRouter() *Router {
	return &Router{
		commandHandlers: make(map[string]Handler),
		textHandlers:    []Handler{},
		privateHandlers: []Handler{},
		groupHandlers:   []Handler{},
	}
}

// Handle registers a handler for a specific command
func (r *Router) Handle(command string, handler Handler) {
	r.commandHandlers[command] = handler
}

// OnText registers a handler for text messages
func (r *Router) OnText(handler Handler) {
	r.textHandlers = append(r.textHandlers, handler)
}

// OnPrivate registers a handler for private messages
func (r *Router) OnPrivate(handler Handler) {
	r.privateHandlers = append(r.privateHandlers, handler)
}

// OnGroup registers a handler for group messages
func (r *Router) OnGroup(handler Handler) {
	r.groupHandlers = append(r.groupHandlers, handler)
}

// HandleUpdate routes an update to the appropriate handler
func (r *Router) HandleUpdate(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	if update.Message != nil {
		if update.Message.Text != "" {
			if strings.HasPrefix(update.Message.Text, "/") {
				command := strings.Split(update.Message.Text, " ")[0]
				if handler, ok := r.commandHandlers[command]; ok {
					handler(ctx, bot, update)
					return
				}
			}

			// Handle text messages
			for _, handler := range r.textHandlers {
				handler(ctx, bot, update)
			}
		}

		// Handle private messages
		if update.Message.Chat.Type == "private" {
			for _, handler := range r.privateHandlers {
				handler(ctx, bot, update)
			}
		}

		// Handle group messages
		if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
			for _, handler := range r.groupHandlers {
				handler(ctx, bot, update)
			}
		}
	}

	// Handle unrecognized commands or non-command messages
	defaultHandler(ctx, bot, update)
}

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
		xtelebot.WithDebug(true),
		xtelebot.WithHttpClientDebug(false),
		xtelebot.WithErrorHandler(customErrorHandler),
		// xtelebot.WithBaseURL("https://api.telegram.org/bot"),
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

	// Create a new router
	router := NewRouter()

	// Register command handlers
	router.Handle("/start", startHandler)
	router.Handle("/hello", helloHandler)
	router.Handle("/photo", photoHandler)
	router.Handle("/location", locationHandler)
	router.Handle("/keyboard", keyboardHandler)
	router.Handle("/echo", echoHandler)
	router.Handle("/deletecommands", deleteCommandsHandler)

	// Register other types of handlers
	router.OnText(textHandler)
	router.OnPrivate(privateHandler)
	router.OnGroup(groupHandler)

	// Create a channel to receive updates
	updatesChan, err := bot.GetUpdatesChan(ctx, offset, 100)
	if err != nil {
		xlog.Error("Failed to get updates channel", "error", err)
		return
	}

	// Listen for updates
	for update := range updatesChan {
		router.HandleUpdate(ctx, bot, update)
		offset = update.UpdateID + 1
	}
}

// Handlers

func startHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	sendWelcomeMessage(ctx, bot, update.Message.Chat.ID)
}

func helloHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	sendGreeting(ctx, bot, update.Message.Chat.ID, update.Message.From.FirstName)
}

func photoHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	sendPhoto(ctx, bot, update.Message.Chat.ID)
}

func locationHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	sendLocation(ctx, bot, update.Message.Chat.ID)
}

func keyboardHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	sendInlineKeyboard(ctx, bot, update.Message.Chat.ID)
}

func echoHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	echoMessage(ctx, bot, update.Message)
}

func deleteCommandsHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	deleteCommands(ctx, bot, update.Message.Chat.ID)
}

func defaultHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	if update.Message != nil {
		sendDefaultResponse(ctx, bot, update.Message.Chat.ID)
	} else if update.CallbackQuery != nil {
		handleCallbackQuery(ctx, bot, update.CallbackQuery)
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

	err := bot.SetMyCommands(ctx, commands)
	if err != nil {
		xlog.Error("Failed to set bot commands", "error", err)
	} else {
		xlog.Info("Bot commands set successfully")
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

func sendGreeting(ctx context.Context, bot *xtelebot.Bot, chatID int64, name string) {
	greeting := fmt.Sprintf("Hello, %s! How are you today?", name)
	xlog.Info("chatID", "data", chatID)
	sendMessage(ctx, bot, chatID, greeting)
	sendDiceParam := xtelebot.DiceConfig{
		Emoji: "🎲",
		BaseChat: xtelebot.BaseChat{
			ChatID: fmt.Sprintf("%d", chatID),
		},
	}

	bot.APIRequestWithObject(ctx, sendDiceParam.Method(), sendDiceParam)
}

func sendPhoto(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	photoURL := "https://www.google.com/images/branding/googlelogo/2x/googlelogo_light_color_272x92dp.png"
	resp, err := x.Must1(xhttpc.NewClient()).Get(ctx, photoURL)
	if err != nil {
		xlog.Error("Failed to download photo", "error", err)
		return
	}
	defer resp.Body.Close()

	_, err = bot.SendPhoto(ctx, chatID, resp.Body)
	if err != nil {
		xlog.Error("Failed to send photo", "error", err)
	}

	// Send photo by URL
	_, err = bot.SendPhoto(ctx, chatID, photoURL)
	if err != nil {
		xlog.Error("Failed to send photo", "error", err)
	}
}

func sendLocation(ctx context.Context, bot *xtelebot.Bot, chatID interface{}) {
	// Coordinates for New York Times Square
	latitude := 40.758896
	longitude := -73.985130
	_, err := bot.SendLocation(ctx, chatID, latitude, longitude)
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

	_, err := bot.SendMessage(ctx, chatID, "Here's an inline keyboard:", xtelebot.WithReplyMarkup(keyboard))
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
		latitude := 40.758896
		longitude := -73.985130
		_, err := bot.SendLocation(ctx, query.Message.Chat.ID, latitude, longitude)
		if err != nil {
			xlog.Error("Failed to send location", "error", err)
		}
		text = "I've sent you the location of New York Times Square."
	default:
		text = "Unknown button clicked."
	}

	// Answer the callback query
	err := bot.AnswerCallbackQuery(ctx, query.ID, xtelebot.WithCallbackQueryText(text))
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
	err := bot.DeleteMyCommands(ctx)
	if err != nil {
		xlog.Error("Failed to delete bot commands", "error", err)
		sendMessage(ctx, bot, chatID, "Failed to delete bot commands")
	} else {
		xlog.Info("Bot commands deleted successfully")
		sendMessage(ctx, bot, chatID, "Bot commands deleted successfully")
	}
}

func sendMessage(ctx context.Context, bot *xtelebot.Bot, chatID interface{}, text string) {
	_, err := bot.SendMessage(ctx, chatID, text)
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
		"username", botInfo.UserName,
		"can_join_groups", botInfo.CanJoinGroups,
		"can_read_all_group_messages", botInfo.CanReadAllGroupMessages,
		"supports_inline_queries", botInfo.SupportsInlineQueries,
	)
}

// New handlers for different types of messages

func textHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	xlog.Info("Received text message", "from", update.Message.From.UserName, "text", update.Message.Text)
	// Add your text message handling logic here
}

func privateHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	xlog.Info("Received private message", "from", update.Message.From.UserName)
	// Add your private message handling logic here
}

func groupHandler(ctx context.Context, bot *xtelebot.Bot, update xtelebot.Update) {
	xlog.Info("Received group message", "from", update.Message.From.UserName, "group", update.Message.Chat.Title)
	// Add your group message handling logic here
}
