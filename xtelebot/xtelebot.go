package xtelebot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
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
	MethodEditMessageText     = "editMessageText"
	MethodEditMessageCaption  = "editMessageCaption"
	MethodEditMessageMedia    = "editMessageMedia"

	// Parameter Keys
	ParamChatID                = "chat_id"
	ParamText                  = "text"
	ParamPhoto                 = "photo"
	ParamDocument              = "document"
	ParamLatitude              = "latitude"
	ParamLongitude             = "longitude"
	ParamReplyMarkup           = "reply_markup"
	ParamParseMode             = "parse_mode"
	ParamCallbackQueryID       = "callback_query_id"
	ParamURL                   = "url"
	ParamMaxConnections        = "max_connections"
	ParamAllowedUpdates        = "allowed_updates"
	ParamCertificate           = "certificate"
	ParamCommands              = "commands"
	ParamAudio                 = "audio"
	ParamVideo                 = "video"
	ParamVoice                 = "voice"
	ParamVideoNote             = "video_note"
	ParamMedia                 = "media"
	ParamOffset                = "offset"
	ParamLimit                 = "limit"
	ParamBusinessConnectionID  = "business_connection_id"
	ParamMessageID             = "message_id"
	ParamInlineMessageID       = "inline_message_id"
	ParamEntities              = "entities"
	ParamLinkPreviewOptions    = "link_preview_options"
	ParamCaption               = "caption"
	ParamCaptionEntities       = "caption_entities"
	ParamShowCaptionAboveMedia = "show_caption_above_media"

	DefaultBaseURL = "https://api.telegram.org/bot"
)

// Bot represents a Telegram Bot
type Bot struct {
	token          string
	baseURL        string
	client         *xhttpc.Client
	errorHandler   ErrorHandler
	requestTimeout time.Duration
	rateLimiter    *RateLimiter
	debug          bool
	isTestServer   bool
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
		baseURL:        DefaultBaseURL,
		client:         x.Must1(xhttpc.NewClient()),
		errorHandler:   defaultErrorHandler,
		requestTimeout: 30 * time.Second,
		rateLimiter:    NewRateLimiter(0, 0), // Default to no rate limiting
		debug:          false,                // Default debug mode is off
		isTestServer:   false,                // Default to production server
	}

	for _, option := range options {
		option(bot)
	}

	if bot.debug {
		xlog.Debug("Bot created in debug mode")
		bot.client.SetDebug(true)
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

// WithDebug enables or disables debug mode
func WithDebug(debug bool) BotOption {
	return func(b *Bot) {
		b.debug = debug
	}
}

// WithTestServer sets the bot to use the Telegram test server
func WithTestServer() BotOption {
	return func(b *Bot) {
		b.isTestServer = true
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
var urlValuesMethods = []string{
	MethodGetMe,
	MethodGetUpdates,
	MethodGetWebhookInfo,
	MethodGetMyCommands,
}

func (b *Bot) APIRequestWithObject(ctx context.Context, method string, params interface{}) ([]byte, error) {
	jsonData, err := x.ToJSON(params)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to convert params to JSON")
	}
	values, err := x.JSONToURLValues(jsonData)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to convert JSON to URL values")
	}
	return b.APIRequest(ctx, method, values)
}

func (b *Bot) APIRequest(ctx context.Context, method string, params url.Values) ([]byte, error) {
	if b.debug {
		xlog.Debug("Sending API request", "method", method, "params", params)
	}

	b.rateLimiter.Wait() // Wait for rate limiting

	var url string
	if b.isTestServer {
		url = fmt.Sprintf("%s%s/test/%s", b.baseURL, b.token, method)
	} else {
		url = fmt.Sprintf("%s%s/%s", b.baseURL, b.token, method)
	}

	ctx, cancel := context.WithTimeout(ctx, b.requestTimeout)
	defer cancel()

	var resp *http.Response
	var err error

	if x.Contains(urlValuesMethods, method) {
		resp, err = b.client.PostURLEncoded(ctx, url, xhttpc.URLEncodedForm(params))
	} else {
		formData := make(xhttpc.FormData)
		for key, values := range params {
			if len(values) > 0 {
				formData[key] = values[0]
			}
		}
		resp, err = b.client.PostFormData(ctx, url, formData)
	}

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

	if b.debug {
		xlog.Debug("API response received", "status", resp.StatusCode, "body", string(body))
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
	body, err := b.APIRequest(ctx, MethodGetMe, url.Values{})
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

	if b.debug {
		xlog.Debug("Sending message", "chatID", chatIDOrUsername, "text", text)
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

	if b.debug {
		xlog.Debug("Message sent", "result", resp.Result)
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
func (b *Bot) SendPhoto(ctx context.Context, chatIDOrUsername interface{}, photo interface{}, options ...MessageOption) (*Message, error) {
	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}

	for _, option := range options {
		option(params)
	}

	var resp *http.Response
	var err error

	switch v := photo.(type) {
	case string:
		// If photo is a string, assume it's a file_id or URL
		params.Set(ParamPhoto, v)
		body, err := b.APIRequest(ctx, MethodSendPhoto, params)
		if err != nil {
			return nil, err
		}
		resp = &http.Response{
			Body: io.NopCloser(bytes.NewReader(body)),
		}
	case io.Reader:
		// If photo is an io.Reader, upload it as a file
		resp, err = b.uploadFile(ctx, MethodSendPhoto, params, ParamPhoto, "photo.jpg", v)
	default:
		return nil, fmt.Errorf("photo must be either a string (file_id or URL) or io.Reader")
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("API response not OK")
	}

	return &result.Result, nil
}

// SendDocument sends a document to a chat
func (b *Bot) SendDocument(ctx context.Context, chatIDOrUsername interface{}, document interface{}, options ...MessageOption) (*Message, error) {
	params := url.Values{}
	switch v := chatIDOrUsername.(type) {
	case int64:
		params.Set(ParamChatID, strconv.FormatInt(v, 10))
	case string:
		params.Set(ParamChatID, v)
	default:
		return nil, fmt.Errorf("chat_id must be either int64 or string")
	}

	for _, option := range options {
		option(params)
	}

	var resp *http.Response
	var err error
	var body []byte

	switch v := document.(type) {
	case string:
		// If document is a string, assume it's a file_id or URL
		params.Set(ParamDocument, v)
		body, err = b.APIRequest(ctx, MethodSendDocument, params)
		if err != nil {
			return nil, err
		}
	case io.Reader:
		// If document is an io.Reader, upload it as a file
		resp, err = b.uploadFile(ctx, MethodSendDocument, params, ParamDocument, "document", v)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	default:
		return nil, fmt.Errorf("document must be either a string (file_id or URL) or io.Reader")
	}

	var result struct {
		Ok     bool    `json:"ok"`
		Result Message `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("API response not OK")
	}

	return &result.Result, nil
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

// NewInlineKeyboardMarkup creates a new inline keyboard markup
func NewInlineKeyboardMarkup(buttons ...[]InlineKeyboardButton) InlineKeyboardMarkup {
	return InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// NewInlineKeyboardButton creates a new inline keyboard button
func NewInlineKeyboardButton(text, url, callbackData string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		URL:          &url,
		CallbackData: &callbackData,
	}
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

// EditMessageText edits text and game messages
func (b *Bot) EditMessageText(ctx context.Context, options ...MessageOption) (*Message, error) {
	params := url.Values{}

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodEditMessageText, params)
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

// EditMessageCaption edits captions of messages
func (b *Bot) EditMessageCaption(ctx context.Context, options ...MessageOption) (*Message, error) {
	params := url.Values{}

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodEditMessageCaption, params)
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

// EditMessageMedia edits animation, audio, document, photo, or video messages
func (b *Bot) EditMessageMedia(ctx context.Context, options ...MessageOption) (*Message, error) {
	params := url.Values{}

	for _, option := range options {
		option(params)
	}

	body, err := b.APIRequest(ctx, MethodEditMessageMedia, params)
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

// WithBusinessConnectionID sets the business_connection_id parameter
func WithBusinessConnectionID(id string) MessageOption {
	return func(v url.Values) {
		v.Set("business_connection_id", id)
	}
}

// WithChatID sets the chat_id parameter
func WithChatID(chatID interface{}) MessageOption {
	return func(v url.Values) {
		switch id := chatID.(type) {
		case int64:
			v.Set(ParamChatID, strconv.FormatInt(id, 10))
		case string:
			v.Set(ParamChatID, id)
		}
	}
}

// WithMessageID sets the message_id parameter
func WithMessageID(messageID int) MessageOption {
	return func(v url.Values) {
		v.Set(ParamMessageID, strconv.Itoa(messageID))
	}
}

// WithInlineMessageID sets the inline_message_id parameter
func WithInlineMessageID(inlineMessageID string) MessageOption {
	return func(v url.Values) {
		v.Set("inline_message_id", inlineMessageID)
	}
}

// WithText sets the text parameter
func WithText(text string) MessageOption {
	return func(v url.Values) {
		v.Set(ParamText, text)
	}
}

// WithCaption sets the caption parameter
func WithCaption(caption string) MessageOption {
	return func(v url.Values) {
		v.Set(ParamCaption, caption)
	}
}

// WithEntities sets the entities parameter
func WithEntities(entities []MessageEntity) MessageOption {
	return func(v url.Values) {
		entitiesJSON, _ := json.Marshal(entities)
		v.Set(ParamEntities, string(entitiesJSON))
	}
}

// WithShowCaptionAboveMedia sets the show_caption_above_media parameter
func WithShowCaptionAboveMedia(show bool) MessageOption {
	return func(v url.Values) {
		v.Set(ParamShowCaptionAboveMedia, strconv.FormatBool(show))
	}
}

// WithMedia sets the media parameter
// InputMedia represents the content of a media message to be sent
type InputMedia interface {
	GetType() string
}

// InputMediaPhoto represents a photo to be sent
type InputMediaPhoto struct {
	Type                  string          `json:"type"`
	Media                 string          `json:"media"`
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	HasSpoiler            bool            `json:"has_spoiler,omitempty"`
}

func (imp InputMediaPhoto) GetType() string {
	return "photo"
}

// WithMedia sets the media parameter
func WithMedia(media InputMedia) MessageOption {
	return func(v url.Values) {
		mediaJSON, _ := json.Marshal(media)
		v.Set(ParamMedia, string(mediaJSON))
	}
}

// LinkPreviewOptions represents the options for link preview generation
type LinkPreviewOptions struct {
	IsDisabled       bool   `json:"is_disabled,omitempty"`
	URL              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefer_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

// WithLinkPreviewOptions sets the link_preview_options parameter
func WithLinkPreviewOptions(options LinkPreviewOptions) MessageOption {
	return func(v url.Values) {
		optionsJSON, _ := json.Marshal(options)
		v.Set(ParamLinkPreviewOptions, string(optionsJSON))
	}
}

// GetUpdatesChan returns a channel that receives updates
func (b *Bot) GetUpdatesChan(ctx context.Context, offset int, limit int) (<-chan Update, error) {
	if b.debug {
		xlog.Debug("Starting update channel", "offset", offset, "limit", limit)
	}

	updates := make(chan Update, 100)

	go func() {
		defer close(updates)

		for {
			select {
			case <-ctx.Done():
				if b.debug {
					xlog.Debug("Update channel closed due to context cancellation")
				}
				return
			default:
				newUpdates, err := b.GetUpdates(ctx, offset, limit)
				if err != nil {
					b.errorHandler(fmt.Errorf("failed to get updates: %w", err))
					time.Sleep(5 * time.Second)
					continue
				}

				if b.debug {
					xlog.Debug("Received updates", "count", len(newUpdates))
				}

				for _, update := range newUpdates {
					select {
					case <-ctx.Done():
						if b.debug {
							xlog.Debug("Update channel closed due to context cancellation")
						}
						return
					case updates <- update:
						offset = update.UpdateID + 1
						if b.debug {
							xlog.Debug("Sent update to channel", "updateID", update.UpdateID)
						}
					}
				}

				if len(newUpdates) == 0 {
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()

	return updates, nil
}

// uploadFile uploads a file using multipart/form-data
func (b *Bot) uploadFile(ctx context.Context, method string, params url.Values, fieldName string, fileName string, fileData io.Reader) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add other fields
	for key, values := range params {
		for _, value := range values {
			err := writer.WriteField(key, value)
			if err != nil {
				return nil, fmt.Errorf("failed to write field: %w", err)
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	var url string
	if b.isTestServer {
		url = fmt.Sprintf("%s%s/test/%s", b.baseURL, b.token, method)
	} else {
		url = fmt.Sprintf("%s%s/%s", b.baseURL, b.token, method)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return b.client.Do(req)
}
