package xtelebot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
)

const (
	// API Methods
	MethodSendMessage         = "sendMessage"
	MethodGetUpdates          = "getUpdates"
	MethodSendPhoto           = "sendPhoto"
	MethodSendDocument        = "sendDocument"
	MethodSendLocation        = "sendLocation"
	MethodAnswerCallbackQuery = "answerCallbackQuery"
	MethodSetWebhook          = "setWebhook"
	MethodDeleteWebhook       = "deleteWebhook"
	MethodGetWebhookInfo      = "getWebhookInfo"
	MethodSetMyCommands       = "setMyCommands"
	MethodDeleteMyCommands    = "deleteMyCommands"
	MethodGetMyCommands       = "getMyCommands"
	MethodSendAudio           = "sendAudio"
	MethodSendVideo           = "sendVideo"
	MethodSendVoice           = "sendVoice"
	MethodSendVideoNote       = "sendVideoNote"
	MethodSendMediaGroup      = "sendMediaGroup"
	MethodGetMe               = "getMe"

	// Parameter Keys
	ParamChatID          = "chat_id"
	ParamText            = "text"
	ParamPhoto           = "photo"
	ParamDocument        = "document"
	ParamLatitude        = "latitude"
	ParamLongitude       = "longitude"
	ParamReplyMarkup     = "reply_markup"
	ParamParseMode       = "parse_mode"
	ParamCallbackQueryID = "callback_query_id"
	ParamURL             = "url"
	ParamMaxConnections  = "max_connections"
	ParamAllowedUpdates  = "allowed_updates"
	ParamCertificate     = "certificate"
	ParamCommands        = "commands"
	ParamAudio           = "audio"
	ParamVideo           = "video"
	ParamVoice           = "voice"
	ParamVideoNote       = "video_note"
	ParamMedia           = "media"
	ParamOffset          = "offset"
	ParamLimit           = "limit"
)

// Bot represents a Telegram Bot
type Bot struct {
	token          string
	baseURL        string
	client         *xhttpc.Client
	errorHandler   ErrorHandler
	requestTimeout time.Duration
	rateLimiter    *RateLimiter
}

// BotOption allows customizing the Bot
type BotOption func(*Bot)

// ErrorHandler is a function that handles errors
type ErrorHandler func(error)

// NewBot creates a new Telegram Bot instance
func NewBot(token string, options ...BotOption) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	bot := &Bot{
		token:          token,
		baseURL:        "https://api.telegram.org",
		client:         xhttpc.NewClient(xhttpc.WithDebug(true)),
		errorHandler:   defaultErrorHandler,
		requestTimeout: 30 * time.Second,
		rateLimiter:    NewRateLimiter(0, 0), // Default to no rate limiting
	}

	for _, option := range options {
		option(bot)
	}

	return bot, nil
}

// WithBaseURL sets a custom base URL for the Telegram API
func WithBaseURL(baseURL string) BotOption {
	return func(b *Bot) {
		if baseURL != "" {
			b.baseURL = baseURL
		}
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *xhttpc.Client) BotOption {
	return func(b *Bot) {
		if client != nil {
			b.client = client
		}
	}
}

// WithErrorHandler sets a custom error handler
func WithErrorHandler(handler ErrorHandler) BotOption {
	return func(b *Bot) {
		if handler != nil {
			b.errorHandler = handler
		}
	}
}

// WithRequestTimeout sets a custom timeout for API requests
func WithRequestTimeout(timeout time.Duration) BotOption {
	return func(b *Bot) {
		if timeout > 0 {
			b.requestTimeout = timeout
		}
	}
}

// WithRateLimit sets a rate limit for API requests
func WithRateLimit(requestsPerSecond float64, burstSize int) BotOption {
	return func(b *Bot) {
		if requestsPerSecond > 0 && burstSize > 0 {
			b.rateLimiter = NewRateLimiter(requestsPerSecond, burstSize)
		}
	}
}

// defaultErrorHandler is the default error handler that logs errors using xlog
func defaultErrorHandler(err error) {
	xlog.Error("Telegram Bot error", "error", err)
}

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	rate       float64
	bucketSize int
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(rate float64, bucketSize int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     float64(bucketSize),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.tokens = min(float64(r.bucketSize), r.tokens+elapsed*r.rate)
	r.lastRefill = now

	if r.tokens < 1 {
		time.Sleep(time.Duration((1-r.tokens)/r.rate) * time.Second)
		r.tokens = 0
	} else {
		r.tokens--
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// APIRequest sends a request to the Telegram API
func (b *Bot) APIRequest(ctx context.Context, method string, params url.Values) ([]byte, error) {
	b.rateLimiter.Wait() // Wait for rate limiting

	url := fmt.Sprintf("%s/bot%s/%s", b.baseURL, b.token, method)

	ctx, cancel := context.WithTimeout(ctx, b.requestTimeout)
	defer cancel()

	resp, err := b.client.Post(ctx, url, params)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to send request: %w", err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to read response body: %w", err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
		b.errorHandler(err)
		return nil, err
	}

	return body, nil
}

// GetMe gets information about the bot
func (b *Bot) GetMe(ctx context.Context) (*User, error) {
	body, err := b.APIRequest(ctx, MethodGetMe, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot information: %w", err)
	}

	var resp struct {
		Ok     bool `json:"ok"`
		Result User `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendMessage sends a text message to a chat
func (b *Bot) SendMessage(ctx context.Context, chatIDOrUsername interface{}, text string, options ...MessageOption) (*Message, error) {
	if text == "" {
		return nil, fmt.Errorf("message text cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamText, text)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendMessage, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// GetUpdates gets updates from the Telegram server
func (b *Bot) GetUpdates(ctx context.Context, offset int, limit int) ([]Update, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}

	params := url.Values{}
	if offset > 0 {
		params.Set(ParamOffset, strconv.Itoa(offset))
	}
	params.Set(ParamLimit, strconv.Itoa(limit))

	body, err := b.APIRequest(ctx, MethodGetUpdates, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return resp.Result, nil
}

// SendPhoto sends a photo to a chat
func (b *Bot) SendPhoto(ctx context.Context, chatIDOrUsername interface{}, photo string, options ...MessageOption) (*Message, error) {
	if photo == "" {
		return nil, fmt.Errorf("photo cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamPhoto, photo)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendPhoto, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendDocument sends a document to a chat
func (b *Bot) SendDocument(ctx context.Context, chatIDOrUsername interface{}, document string, options ...MessageOption) (*Message, error) {
	if document == "" {
		return nil, fmt.Errorf("document cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamDocument, document)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendDocument, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendLocation sends a location to a chat
func (b *Bot) SendLocation(ctx context.Context, chatIDOrUsername interface{}, latitude, longitude float64, options ...MessageOption) (*Message, error) {
	if latitude < -90 || latitude > 90 {
		return nil, fmt.Errorf("latitude must be between -90 and 90")
	}
	if longitude < -180 || longitude > 180 {
		return nil, fmt.Errorf("longitude must be between -180 and 180")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamLatitude, strconv.FormatFloat(latitude, 'f', 6, 64))
	params.Set(ParamLongitude, strconv.FormatFloat(longitude, 'f', 6, 64))

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendLocation, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// MessageOption allows customizing message parameters
type MessageOption func(v url.Values)

// WithReplyMarkup adds a reply markup to the message
func WithReplyMarkup(markup interface{}) MessageOption {
	return func(v url.Values) {
		jsonMarkup, err := json.Marshal(markup)
		if err != nil {
			xlog.Error("Failed to marshal reply markup", "error", err)
			return
		}
		v.Set(ParamReplyMarkup, string(jsonMarkup))
	}
}

// WithParseMode sets the parse mode for the message
func WithParseMode(mode string) MessageOption {
	return func(v url.Values) {
		v.Set(ParamParseMode, mode)
	}
}

// Update represents a Telegram update
type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

// Message represents a Telegram message
type Message struct {
	MessageID int         `json:"message_id"`
	Text      string      `json:"text"`
	Chat      Chat        `json:"chat"`
	From      User        `json:"from"`
	Date      int         `json:"date"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Document  Document    `json:"document,omitempty"`
	Location  Location    `json:"location,omitempty"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// User represents a Telegram user
type User struct {
	ID                      int    `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name,omitempty"`
	Username                string `json:"username,omitempty"`
	LanguageCode            string `json:"language_code,omitempty"`
	CanJoinGroups           bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages,omitempty"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries,omitempty"`
}

// PhotoSize represents one size of a photo or a file / sticker thumbnail
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size,omitempty"`
}

// Document represents a general file
type Document struct {
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	Thumb        PhotoSize `json:"thumb,omitempty"`
	FileName     string    `json:"file_name,omitempty"`
	MimeType     string    `json:"mime_type,omitempty"`
	FileSize     int       `json:"file_size,omitempty"`
}

// Location represents a point on the map
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// InlineKeyboardMarkup represents an inline keyboard
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents one button of an inline keyboard
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// NewInlineKeyboardMarkup creates a new inline keyboard markup
func NewInlineKeyboardMarkup(buttons ...[]InlineKeyboardButton) InlineKeyboardMarkup {
	return InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// NewInlineKeyboardButton creates a new inline keyboard button
func NewInlineKeyboardButton(text, url, callbackData string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		URL:          url,
		CallbackData: callbackData,
	}
}

// CallbackQuery represents an incoming callback query from a callback button in an inline keyboard
type CallbackQuery struct {
	ID              string   `json:"id"`
	From            User     `json:"from"`
	Message         *Message `json:"message,omitempty"`
	InlineMessageID string   `json:"inline_message_id,omitempty"`
	ChatInstance    string   `json:"chat_instance"`
	Data            string   `json:"data,omitempty"`
	GameShortName   string   `json:"game_short_name,omitempty"`
}

// AnswerCallbackQuery sends a response to a callback query
func (b *Bot) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, options ...CallbackQueryOption) error {
	if callbackQueryID == "" {
		return fmt.Errorf("callback query ID cannot be empty")
	}

	params := url.Values{}
	params.Set(ParamCallbackQueryID, callbackQueryID)

	for _, option := range options {
		option(params)
	}

	_, err := b.APIRequest(ctx, MethodAnswerCallbackQuery, params)
	return err
}

// CallbackQueryOption allows customizing callback query response parameters
type CallbackQueryOption func(v url.Values)

// WithCallbackQueryText sets the text to be shown to the user
func WithCallbackQueryText(text string) CallbackQueryOption {
	return func(v url.Values) {
		v.Set(ParamText, text)
	}
}

// WithCallbackQueryShowAlert sets whether to show an alert instead of a notification
func WithCallbackQueryShowAlert(showAlert bool) CallbackQueryOption {
	return func(v url.Values) {
		v.Set("show_alert", strconv.FormatBool(showAlert))
	}
}

// SetWebhook sets a webhook for the bot
func (b *Bot) SetWebhook(ctx context.Context, webhookURL string, options ...WebhookOption) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	params := url.Values{}
	params.Set(ParamURL, webhookURL)

	for _, option := range options {
		option(params)
	}

	_, err := b.APIRequest(ctx, MethodSetWebhook, params)
	return err
}

// DeleteWebhook deletes the webhook for the bot
func (b *Bot) DeleteWebhook(ctx context.Context) error {
	_, err := b.APIRequest(ctx, MethodDeleteWebhook, nil)
	return err
}

// GetWebhookInfo gets the current webhook status
func (b *Bot) GetWebhookInfo(ctx context.Context) (*WebhookInfo, error) {
	body, err := b.APIRequest(ctx, MethodGetWebhookInfo, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool        `json:"ok"`
		Result WebhookInfo `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// WebhookOption allows customizing webhook parameters
type WebhookOption func(v url.Values)

// WithMaxConnections sets the maximum allowed number of simultaneous HTTPS connections to the webhook
func WithMaxConnections(max int) WebhookOption {
	return func(v url.Values) {
		if max < 1 || max > 100 {
			xlog.Warn("Invalid max connections value, must be between 1 and 100")
			return
		}
		v.Set(ParamMaxConnections, strconv.Itoa(max))
	}
}

// WithAllowedUpdates sets a list of the update types you want your bot to receive
func WithAllowedUpdates(types []string) WebhookOption {
	return func(v url.Values) {
		jsonTypes, err := json.Marshal(types)
		if err != nil {
			xlog.Error("Failed to marshal allowed updates", "error", err)
			return
		}
		v.Set(ParamAllowedUpdates, string(jsonTypes))
	}
}

// WithCertificate sets a self-signed certificate for webhook
func WithCertificate(cert string) WebhookOption {
	return func(v url.Values) {
		v.Set(ParamCertificate, cert)
	}
}

// WebhookInfo contains information about the current status of a webhook
type WebhookInfo struct {
	URL                  string   `json:"url"`
	HasCustomCertificate bool     `json:"has_custom_certificate"`
	PendingUpdateCount   int      `json:"pending_update_count"`
	LastErrorDate        int      `json:"last_error_date"`
	LastErrorMessage     string   `json:"last_error_message"`
	MaxConnections       int      `json:"max_connections"`
	AllowedUpdates       []string `json:"allowed_updates"`
}

// BotCommand represents a bot command
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// SetMyCommands sets the list of the bot's commands
func (b *Bot) SetMyCommands(ctx context.Context, commands []BotCommand) error {
	if len(commands) == 0 {
		return fmt.Errorf("commands list cannot be empty")
	}

	params := url.Values{}

	jsonCommands, err := json.Marshal(commands)
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	params.Set(ParamCommands, string(jsonCommands))

	_, err = b.APIRequest(ctx, MethodSetMyCommands, params)
	if err != nil {
		return fmt.Errorf("failed to set commands: %w", err)
	}

	return nil
}

// DeleteMyCommands deletes the list of the bot's commands
func (b *Bot) DeleteMyCommands(ctx context.Context) error {
	_, err := b.APIRequest(ctx, MethodDeleteMyCommands, nil)
	if err != nil {
		return fmt.Errorf("failed to delete commands: %w", err)
	}

	return nil
}

// GetMyCommands gets the current list of the bot's commands
func (b *Bot) GetMyCommands(ctx context.Context) ([]BotCommand, error) {
	body, err := b.APIRequest(ctx, MethodGetMyCommands, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}

	var resp struct {
		Ok     bool         `json:"ok"`
		Result []BotCommand `json:"result"`
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

// SendAudio sends an audio file to a chat
func (b *Bot) SendAudio(ctx context.Context, chatIDOrUsername interface{}, audio string, options ...MessageOption) (*Message, error) {
	if audio == "" {
		return nil, fmt.Errorf("audio cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamAudio, audio)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendAudio, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendVideo sends a video file to a chat
func (b *Bot) SendVideo(ctx context.Context, chatIDOrUsername interface{}, video string, options ...MessageOption) (*Message, error) {
	if video == "" {
		return nil, fmt.Errorf("video cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamVideo, video)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendVideo, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendVoice sends a voice message to a chat
func (b *Bot) SendVoice(ctx context.Context, chatIDOrUsername interface{}, voice string, options ...MessageOption) (*Message, error) {
	if voice == "" {
		return nil, fmt.Errorf("voice cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamVoice, voice)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendVoice, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendVideoNote sends a video note to a chat
func (b *Bot) SendVideoNote(ctx context.Context, chatIDOrUsername interface{}, videoNote string, options ...MessageOption) (*Message, error) {
	if videoNote == "" {
		return nil, fmt.Errorf("video note cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}
	params.Set(ParamVideoNote, videoNote)

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendVideoNote, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return &resp.Result, nil
}

// SendMediaGroup sends a group of photos, videos, documents or audios as an album
func (b *Bot) SendMediaGroup(ctx context.Context, chatIDOrUsername interface{}, media []interface{}, options ...MessageOption) ([]Message, error) {
	if len(media) == 0 {
		return nil, fmt.Errorf("media group cannot be empty")
	}

	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal media: %w", err)
	}
	params.Set(ParamMedia, string(mediaJSON))

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodSendMediaGroup, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok     bool      `json:"ok"`
		Result []Message `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		b.errorHandler(fmt.Errorf("failed to unmarshal response: %w", err))
		return nil, err
	}

	if !resp.Ok {
		err := fmt.Errorf("API response not OK")
		b.errorHandler(err)
		return nil, err
	}

	return resp.Result, nil
}
