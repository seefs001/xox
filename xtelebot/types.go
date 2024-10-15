package xtelebot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
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

// Params represents a set of parameters that gets passed to a request.
type Params map[string]string

// APIResponse is a response from the Telegram API with the result
// stored raw.
type APIResponse struct {
	Ok          bool                `json:"ok"`
	Result      json.RawMessage     `json:"result,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	Description string              `json:"description,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
}

// Error is an error containing extra information returned by the Telegram API.
type Error struct {
	Code    int
	Message string
	ResponseParameters
}

// Error message string.
func (e Error) Error() string {
	return e.Message
}

// Update is an update response, from GetUpdates.
type Update struct {
	// UpdateID is the update's unique identifier.
	// Update identifiers start from a certain positive number and increase
	// sequentially.
	// This ID becomes especially handy if you're using Webhooks,
	// since it allows you to ignore repeated updates or to restore
	// the correct update sequence, should they get out of order.
	// If there are no new updates for at least a week, then identifier
	// of the next update will be chosen randomly instead of sequentially.
	UpdateID int `json:"update_id"`
	// Message new incoming message of any kind — text, photo, sticker, etc.
	//
	// optional
	Message *Message `json:"message,omitempty"`
	// EditedMessage new version of a message that is known to the bot and was
	// edited
	//
	// optional
	EditedMessage *Message `json:"edited_message,omitempty"`
	// ChannelPost new version of a message that is known to the bot and was
	// edited
	//
	// optional
	ChannelPost *Message `json:"channel_post,omitempty"`
	// EditedChannelPost new incoming channel post of any kind — text, photo,
	// sticker, etc.
	//
	// optional
	EditedChannelPost *Message `json:"edited_channel_post,omitempty"`
	// InlineQuery new incoming inline query
	//
	// optional
	InlineQuery *InlineQuery `json:"inline_query,omitempty"`
	// ChosenInlineResult is the result of an inline query
	// that was chosen by a user and sent to their chat partner.
	// Please see our documentation on the feedback collecting
	// for details on how to enable these updates for your bot.
	//
	// optional
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
	// CallbackQuery new incoming callback query
	//
	// optional
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
	// ShippingQuery new incoming shipping query. Only for invoices with
	// flexible price
	//
	// optional
	ShippingQuery *ShippingQuery `json:"shipping_query,omitempty"`
	// PreCheckoutQuery new incoming pre-checkout query. Contains full
	// information about checkout
	//
	// optional
	PreCheckoutQuery *PreCheckoutQuery `json:"pre_checkout_query,omitempty"`
	// Pool new poll state. Bots receive only updates about stopped polls and
	// polls, which are sent by the bot
	//
	// optional
	Poll *Poll `json:"poll,omitempty"`
	// PollAnswer user changed their answer in a non-anonymous poll. Bots
	// receive new votes only in polls that were sent by the bot itself.
	//
	// optional
	PollAnswer *PollAnswer `json:"poll_answer,omitempty"`
	// MyChatMember is the bot's chat member status was updated in a chat. For
	// private chats, this update is received only when the bot is blocked or
	// unblocked by the user.
	//
	// optional
	MyChatMember *ChatMemberUpdated `json:"my_chat_member,omitempty"`
	// ChatMember is a chat member's status was updated in a chat. The bot must
	// be an administrator in the chat and must explicitly specify "chat_member"
	// in the list of allowed_updates to receive these updates.
	//
	// optional
	ChatMember *ChatMemberUpdated `json:"chat_member,omitempty"`
	// ChatJoinRequest is a request to join the chat has been sent. The bot must
	// have the can_invite_users administrator right in the chat to receive
	// these updates.
	//
	// optional
	ChatJoinRequest *ChatJoinRequest `json:"chat_join_request,omitempty"`
}

// SentFrom returns the user who sent an update. Can be nil, if Telegram did not provide information
// about the user in the update object.
func (u *Update) SentFrom() *User {
	switch {
	case u.Message != nil:
		return u.Message.From
	case u.EditedMessage != nil:
		return u.EditedMessage.From
	case u.InlineQuery != nil:
		return u.InlineQuery.From
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From
	case u.CallbackQuery != nil:
		return u.CallbackQuery.From
	case u.ShippingQuery != nil:
		return u.ShippingQuery.From
	case u.PreCheckoutQuery != nil:
		return u.PreCheckoutQuery.From
	default:
		return nil
	}
}

// CallbackData returns the callback query data, if it exists.
func (u *Update) CallbackData() string {
	if u.CallbackQuery != nil {
		return u.CallbackQuery.Data
	}
	return ""
}

// FromChat returns the chat where an update occurred.
func (u *Update) FromChat() *Chat {
	switch {
	case u.Message != nil:
		return u.Message.Chat
	case u.EditedMessage != nil:
		return u.EditedMessage.Chat
	case u.ChannelPost != nil:
		return u.ChannelPost.Chat
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost.Chat
	case u.CallbackQuery != nil:
		return u.CallbackQuery.Message.Chat
	default:
		return nil
	}
}

// UpdatesChannel is the channel for getting updates.
type UpdatesChannel <-chan Update

// Clear discards all unprocessed incoming updates.
func (ch UpdatesChannel) Clear() {
	for len(ch) != 0 {
		<-ch
	}
}

// User represents a Telegram user or bot.
type User struct {
	// ID is a unique identifier for this user or bot
	ID int64 `json:"id"`
	// IsBot true, if this user is a bot
	//
	// optional
	IsBot bool `json:"is_bot,omitempty"`
	// IsPremium true, if user has Telegram Premium
	//
	// optional
	IsPremium bool `json:"is_premium,omitempty"`
	// FirstName user's or bot's first name
	FirstName string `json:"first_name"`
	// LastName user's or bot's last name
	//
	// optional
	LastName string `json:"last_name,omitempty"`
	// UserName user's or bot's username
	//
	// optional
	UserName string `json:"username,omitempty"`
	// LanguageCode IETF language tag of the user's language
	// more info: https://en.wikipedia.org/wiki/IETF_language_tag
	//
	// optional
	LanguageCode string `json:"language_code,omitempty"`
	// CanJoinGroups is true, if the bot can be invited to groups.
	// Returned only in getMe.
	//
	// optional
	CanJoinGroups bool `json:"can_join_groups,omitempty"`
	// CanReadAllGroupMessages is true, if privacy mode is disabled for the bot.
	// Returned only in getMe.
	//
	// optional
	CanReadAllGroupMessages bool `json:"can_read_all_group_messages,omitempty"`
	// SupportsInlineQueries is true, if the bot supports inline queries.
	// Returned only in getMe.
	//
	// optional
	SupportsInlineQueries bool `json:"supports_inline_queries,omitempty"`
}

// String displays a simple text version of a user.
//
// It is normally a user's username, but falls back to a first/last
// name as available.
func (u *User) String() string {
	if u == nil {
		return ""
	}
	if u.UserName != "" {
		return u.UserName
	}

	name := u.FirstName
	if u.LastName != "" {
		name += " " + u.LastName
	}

	return name
}

// Chat represents a chat.
type Chat struct {
	// ID is a unique identifier for this chat
	ID int64 `json:"id"`
	// Type of chat, can be either "private", "group", "supergroup" or "channel"
	Type ChatType `json:"type"`
	// Title for supergroups, channels and group chats
	//
	// optional
	Title string `json:"title,omitempty"`
	// UserName for private chats, supergroups and channels if available
	//
	// optional
	UserName string `json:"username,omitempty"`
	// FirstName of the other party in a private chat
	//
	// optional
	FirstName string `json:"first_name,omitempty"`
	// LastName of the other party in a private chat
	//
	// optional
	LastName string `json:"last_name,omitempty"`
	// Photo is a chat photo
	Photo *ChatPhoto `json:"photo"`
	// Bio is the bio of the other party in a private chat. Returned only in
	// getChat
	//
	// optional
	Bio string `json:"bio,omitempty"`
	// HasPrivateForwards is true if privacy settings of the other party in the
	// private chat allows to use tg://user?id=<user_id> links only in chats
	// with the user. Returned only in getChat.
	//
	// optional
	HasPrivateForwards bool `json:"has_private_forwards,omitempty"`
	// Description for groups, supergroups and channel chats
	//
	// optional
	Description string `json:"description,omitempty"`
	// InviteLink is a chat invite link, for groups, supergroups and channel chats.
	// Each administrator in a chat generates their own invite links,
	// so the bot must first generate the link using exportChatInviteLink
	//
	// optional
	InviteLink string `json:"invite_link,omitempty"`
	// PinnedMessage is the pinned message, for groups, supergroups and channels
	//
	// optional
	PinnedMessage *Message `json:"pinned_message,omitempty"`
	// Permissions are default chat member permissions, for groups and
	// supergroups. Returned only in getChat.
	//
	// optional
	Permissions *ChatPermissions `json:"permissions,omitempty"`
	// SlowModeDelay is for supergroups, the minimum allowed delay between
	// consecutive messages sent by each unprivileged user. Returned only in
	// getChat.
	//
	// optional
	SlowModeDelay int `json:"slow_mode_delay,omitempty"`
	// MessageAutoDeleteTime is the time after which all messages sent to the
	// chat will be automatically deleted; in seconds. Returned only in getChat.
	//
	// optional
	MessageAutoDeleteTime int `json:"message_auto_delete_time,omitempty"`
	// HasProtectedContent is true if messages from the chat can't be forwarded
	// to other chats. Returned only in getChat.
	//
	// optional
	HasProtectedContent bool `json:"has_protected_content,omitempty"`
	// StickerSetName is for supergroups, name of group sticker set.Returned
	// only in getChat.
	//
	// optional
	StickerSetName string `json:"sticker_set_name,omitempty"`
	// CanSetStickerSet is true, if the bot can change the group sticker set.
	// Returned only in getChat.
	//
	// optional
	CanSetStickerSet bool `json:"can_set_sticker_set,omitempty"`
	// LinkedChatID is a unique identifier for the linked chat, i.e. the
	// discussion group identifier for a channel and vice versa; for supergroups
	// and channel chats.
	//
	// optional
	LinkedChatID int64 `json:"linked_chat_id,omitempty"`
	// Location is for supergroups, the location to which the supergroup is
	// connected. Returned only in getChat.
	//
	// optional
	Location *ChatLocation `json:"location,omitempty"`
}

// Constants for chat types
const (
	ChatTypePrivate    ChatType = "private"
	ChatTypeGroup      ChatType = "group"
	ChatTypeSupergroup ChatType = "supergroup"
	ChatTypeChannel    ChatType = "channel"
)

// IsPrivate returns true if the chat is a private chat
func (c *Chat) IsPrivate() bool {
	return c.Type == ChatTypePrivate
}

// IsGroup returns true if the chat is a group chat
func (c *Chat) IsGroup() bool {
	return c.Type == ChatTypeGroup
}

// IsSuperGroup returns true if the chat is a supergroup chat
func (c *Chat) IsSuperGroup() bool {
	return c.Type == ChatTypeSupergroup
}

// IsChannel returns true if the chat is a channel
func (c *Chat) IsChannel() bool {
	return c.Type == ChatTypeChannel
}

// Message represents a message.
type Message struct {
	// MessageID is a unique message identifier inside this chat
	MessageID int `json:"message_id"`
	// From is a sender, empty for messages sent to channels;
	//
	// optional
	From *User `json:"from,omitempty"`
	// SenderChat is the sender of the message, sent on behalf of a chat. The
	// channel itself for channel messages. The supergroup itself for messages
	// from anonymous group administrators. The linked channel for messages
	// automatically forwarded to the discussion group
	//
	// optional
	SenderChat *Chat `json:"sender_chat,omitempty"`
	// Date of the message was sent in Unix time
	Date int `json:"date"`
	// Chat is the conversation the message belongs to
	Chat *Chat `json:"chat"`
	// ForwardFrom for forwarded messages, sender of the original message;
	//
	// optional
	ForwardFrom *User `json:"forward_from,omitempty"`
	// ForwardFromChat for messages forwarded from channels,
	// information about the original channel;
	//
	// optional
	ForwardFromChat *Chat `json:"forward_from_chat,omitempty"`
	// ForwardFromMessageID for messages forwarded from channels,
	// identifier of the original message in the channel;
	//
	// optional
	ForwardFromMessageID int `json:"forward_from_message_id,omitempty"`
	// ForwardSignature for messages forwarded from channels, signature of the
	// post author if present
	//
	// optional
	ForwardSignature string `json:"forward_signature,omitempty"`
	// ForwardSenderName is the sender's name for messages forwarded from users
	// who disallow adding a link to their account in forwarded messages
	//
	// optional
	ForwardSenderName string `json:"forward_sender_name,omitempty"`
	// ForwardDate for forwarded messages, date the original message was sent in Unix time;
	//
	// optional
	ForwardDate int `json:"forward_date,omitempty"`
	// IsAutomaticForward is true if the message is a channel post that was
	// automatically forwarded to the connected discussion group.
	//
	// optional
	IsAutomaticForward bool `json:"is_automatic_forward,omitempty"`
	// ReplyToMessage for replies, the original message.
	// Note that the Message object in this field will not contain further ReplyToMessage fields
	// even if it itself is a reply;
	//
	// optional
	ReplyToMessage *Message `json:"reply_to_message,omitempty"`
	// ViaBot through which the message was sent;
	//
	// optional
	ViaBot *User `json:"via_bot,omitempty"`
	// EditDate of the message was last edited in Unix time;
	//
	// optional
	EditDate int `json:"edit_date,omitempty"`
	// HasProtectedContent is true if the message can't be forwarded.
	//
	// optional
	HasProtectedContent bool `json:"has_protected_content,omitempty"`
	// MediaGroupID is the unique identifier of a media message group this message belongs to;
	//
	// optional
	MediaGroupID string `json:"media_group_id,omitempty"`
	// AuthorSignature is the signature of the post author for messages in channels;
	//
	// optional
	AuthorSignature string `json:"author_signature,omitempty"`
	// Text is for text messages, the actual UTF-8 text of the message, 0-4096 characters;
	//
	// optional
	Text string `json:"text,omitempty"`
	// Entities are for text messages, special entities like usernames,
	// URLs, bot commands, etc. that appear in the text;
	//
	// optional
	Entities []MessageEntity `json:"entities,omitempty"`
	// Animation message is an animation, information about the animation.
	// For backward compatibility, when this field is set, the document field will also be set;
	//
	// optional
	Animation *Animation `json:"animation,omitempty"`
	// PremiumAnimation message is an animation, information about the animation.
	// For backward compatibility, when this field is set, the document field will also be set;
	//
	// optional
	PremiumAnimation *Animation `json:"premium_animation,omitempty"`
	// Audio message is an audio file, information about the file;
	//
	// optional
	Audio *Audio `json:"audio,omitempty"`
	// Document message is a general file, information about the file;
	//
	// optional
	Document *Document `json:"document,omitempty"`
	// Photo message is a photo, available sizes of the photo;
	//
	// optional
	Photo []PhotoSize `json:"photo,omitempty"`
	// Sticker message is a sticker, information about the sticker;
	//
	// optional
	Sticker *Sticker `json:"sticker,omitempty"`
	// Video message is a video, information about the video;
	//
	// optional
	Video *Video `json:"video,omitempty"`
	// VideoNote message is a video note, information about the video message;
	//
	// optional
	VideoNote *VideoNote `json:"video_note,omitempty"`
	// Voice message is a voice message, information about the file;
	//
	// optional
	Voice *Voice `json:"voice,omitempty"`
	// Caption for the animation, audio, document, photo, video or voice, 0-1024 characters;
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// CaptionEntities;
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// Contact message is a shared contact, information about the contact;
	//
	// optional
	Contact *Contact `json:"contact,omitempty"`
	// Dice is a dice with random value;
	//
	// optional
	Dice *Dice `json:"dice,omitempty"`
	// Game message is a game, information about the game;
	//
	// optional
	Game *Game `json:"game,omitempty"`
	// Poll is a native poll, information about the poll;
	//
	// optional
	Poll *Poll `json:"poll,omitempty"`
	// Venue message is a venue, information about the venue.
	// For backward compatibility, when this field is set, the location field
	// will also be set;
	//
	// optional
	Venue *Venue `json:"venue,omitempty"`
	// Location message is a shared location, information about the location;
	//
	// optional
	Location *Location `json:"location,omitempty"`
	// NewChatMembers that were added to the group or supergroup
	// and information about them (the bot itself may be one of these members);
	//
	// optional
	NewChatMembers []User `json:"new_chat_members,omitempty"`
	// LeftChatMember is a member was removed from the group,
	// information about them (this member may be the bot itself);
	//
	// optional
	LeftChatMember *User `json:"left_chat_member,omitempty"`
	// NewChatTitle is a chat title was changed to this value;
	//
	// optional
	NewChatTitle string `json:"new_chat_title,omitempty"`
	// NewChatPhoto is a chat photo was change to this value;
	//
	// optional
	NewChatPhoto []PhotoSize `json:"new_chat_photo,omitempty"`
	// DeleteChatPhoto is a service message: the chat photo was deleted;
	//
	// optional
	DeleteChatPhoto bool `json:"delete_chat_photo,omitempty"`
	// GroupChatCreated is a service message: the group has been created;
	//
	// optional
	GroupChatCreated bool `json:"group_chat_created,omitempty"`
	// SuperGroupChatCreated is a service message: the supergroup has been created.
	// This field can't be received in a message coming through updates,
	// because bot can't be a member of a supergroup when it is created.
	// It can only be found in ReplyToMessage if someone replies to a very first message
	// in a directly created supergroup;
	//
	// optional
	SuperGroupChatCreated bool `json:"supergroup_chat_created,omitempty"`
	// ChannelChatCreated is a service message: the channel has been created.
	// This field can't be received in a message coming through updates,
	// because bot can't be a member of a channel when it is created.
	// It can only be found in ReplyToMessage
	// if someone replies to a very first message in a channel;
	//
	// optional
	ChannelChatCreated bool `json:"channel_chat_created,omitempty"`
	// MessageAutoDeleteTimerChanged is a service message: auto-delete timer
	// settings changed in the chat.
	//
	// optional
	MessageAutoDeleteTimerChanged *MessageAutoDeleteTimerChanged `json:"message_auto_delete_timer_changed,omitempty"`
	// MigrateToChatID is the group has been migrated to a supergroup with the specified identifier.
	// This number may be greater than 32 bits and some programming languages
	// may have difficulty/silent defects in interpreting it.
	// But it is smaller than 52 bits, so a signed 64-bit integer
	// or double-precision float type are safe for storing this identifier;
	//
	// optional
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	// MigrateFromChatID is the supergroup has been migrated from a group with the specified identifier.
	// This number may be greater than 32 bits and some programming languages
	// may have difficulty/silent defects in interpreting it.
	// But it is smaller than 52 bits, so a signed 64-bit integer
	// or double-precision float type are safe for storing this identifier;
	//
	// optional
	MigrateFromChatID int64 `json:"migrate_from_chat_id,omitempty"`
	// PinnedMessage is a specified message was pinned.
	// Note that the Message object in this field will not contain further ReplyToMessage
	// fields even if it is itself a reply;
	//
	// optional
	PinnedMessage *Message `json:"pinned_message,omitempty"`
	// Invoice message is an invoice for a payment;
	//
	// optional
	Invoice *Invoice `json:"invoice,omitempty"`
	// SuccessfulPayment message is a service message about a successful payment,
	// information about the payment;
	//
	// optional
	SuccessfulPayment *SuccessfulPayment `json:"successful_payment,omitempty"`
	// ConnectedWebsite is the domain name of the website on which the user has
	// logged in;
	//
	// optional
	ConnectedWebsite string `json:"connected_website,omitempty"`
	// PassportData is a Telegram Passport data;
	//
	// optional
	PassportData *PassportData `json:"passport_data,omitempty"`
	// ProximityAlertTriggered is a service message. A user in the chat
	// triggered another user's proximity alert while sharing Live Location
	//
	// optional
	ProximityAlertTriggered *ProximityAlertTriggered `json:"proximity_alert_triggered,omitempty"`
	// VideoChatScheduled is a service message: video chat scheduled.
	//
	// optional
	VideoChatScheduled *VideoChatScheduled `json:"video_chat_scheduled,omitempty"`
	// VideoChatStarted is a service message: video chat started.
	//
	// optional
	VideoChatStarted *VideoChatStarted `json:"video_chat_started,omitempty"`
	// VideoChatEnded is a service message: video chat ended.
	//
	// optional
	VideoChatEnded *VideoChatEnded `json:"video_chat_ended,omitempty"`
	// VideoChatParticipantsInvited is a service message: new participants
	// invited to a video chat.
	//
	// optional
	VideoChatParticipantsInvited *VideoChatParticipantsInvited `json:"video_chat_participants_invited,omitempty"`
	// WebAppData is a service message: data sent by a Web App.
	//
	// optional
	WebAppData *WebAppData `json:"web_app_data,omitempty"`
	// ReplyMarkup is the Inline keyboard attached to the message.
	// login_url buttons are represented as ordinary url buttons.
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// Time converts the message timestamp into a Time.
func (m *Message) Time() time.Time {
	return time.Unix(int64(m.Date), 0)
}

// IsCommand returns true if message starts with a "bot_command" entity.
func (m *Message) IsCommand() bool {
	if m.Entities == nil || len(m.Entities) == 0 {
		return false
	}

	entity := m.Entities[0]
	return entity.Offset == 0 && entity.IsCommand()
}

// Command checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at name syntax, it is removed. Use
// CommandWithAt() if you do not want that.
func (m *Message) Command() string {
	command := m.CommandWithAt()

	if i := strings.Index(command, "@"); i != -1 {
		command = command[:i]
	}

	return command
}

// CommandWithAt checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at name syntax, it is not removed. Use Command()
// if you want that.
func (m *Message) CommandWithAt() string {
	if !m.IsCommand() {
		return ""
	}

	// IsCommand() checks that the message begins with a bot_command entity
	entity := m.Entities[0]
	return m.Text[1:entity.Length]
}

// CommandArguments checks if the message was a command and if it was,
// returns all text after the command name. If the Message was not a
// command, it returns an empty string.
//
// Note: The first character after the command name is omitted:
// - "/foo bar baz" yields "bar baz", not " bar baz"
// - "/foo-bar baz" yields "bar baz", too
// Even though the latter is not a command conforming to the spec, the API
// marks "/foo" as command entity.
func (m *Message) CommandArguments() string {
	if !m.IsCommand() {
		return ""
	}

	// IsCommand() checks that the message begins with a bot_command entity
	entity := m.Entities[0]

	if len(m.Text) == entity.Length {
		return "" // The command makes up the whole message
	}

	return m.Text[entity.Length+1:]
}

// MessageID represents a unique message identifier.
type MessageID struct {
	MessageID int `json:"message_id"`
}

// MessageEntity represents one special entity in a text message.
type MessageEntity struct {
	// Type of the entity.
	// Can be:
	//  "mention" (@username),
	//  "hashtag" (#hashtag),
	//  "cashtag" ($USD),
	//  "bot_command" (/start@jobs_bot),
	//  "url" (https://telegram.org),
	//  "email" (do-not-reply@telegram.org),
	//  "phone_number" (+1-212-555-0123),
	//  "bold" (bold text),
	//  "italic" (italic text),
	//  "underline" (underlined text),
	//  "strikethrough" (strikethrough text),
	//  "spoiler" (spoiler message),
	//  "code" (monowidth string),
	//  "pre" (monowidth block),
	//  "text_link" (for clickable text URLs),
	//  "text_mention" (for users without usernames)
	Type string `json:"type"`
	// Offset in UTF-16 code units to the start of the entity
	Offset int `json:"offset"`
	// Length
	Length int `json:"length"`
	// URL for "text_link" only, url that will be opened after user taps on the text
	//
	// optional
	URL string `json:"url,omitempty"`
	// User for "text_mention" only, the mentioned user
	//
	// optional
	User *User `json:"user,omitempty"`
	// Language for "pre" only, the programming language of the entity text
	//
	// optional
	Language string `json:"language,omitempty"`
}

// Constants for message entities
const (
	EntityMention       = "mention"
	EntityHashtag       = "hashtag"
	EntityCashtag       = "cashtag"
	EntityBotCommand    = "bot_command"
	EntityURL           = "url"
	EntityEmail         = "email"
	EntityPhoneNumber   = "phone_number"
	EntityBold          = "bold"
	EntityItalic        = "italic"
	EntityUnderline     = "underline"
	EntityStrikethrough = "strikethrough"
	EntityCode          = "code"
	EntityPre           = "pre"
	EntityTextLink      = "text_link"
	EntityTextMention   = "text_mention"
)

// HasEntities returns true if the message has entities
func (m *Message) HasEntities() bool {
	return len(m.Entities) > 0
}

// GetEntityText returns the text of a specific entity
func (m *Message) GetEntityText(e *MessageEntity) string {
	return m.Text[e.Offset : e.Offset+e.Length]
}

// ParseURL attempts to parse a URL contained within a MessageEntity.
func (e MessageEntity) ParseURL() (*url.URL, error) {
	if e.URL == "" {
		return nil, errors.New(ErrBadURL)
	}

	return url.Parse(e.URL)
}

// IsMention returns true if the type of the message entity is "mention" (@username).
func (e MessageEntity) IsMention() bool {
	return e.Type == EntityMention
}

// IsTextMention returns true if the type of the message entity is "text_mention"
// (At this time, the user field exists, and occurs when tagging a member without a username)
func (e MessageEntity) IsTextMention() bool {
	return e.Type == EntityTextMention
}

// IsHashtag returns true if the type of the message entity is "hashtag".
func (e MessageEntity) IsHashtag() bool {
	return e.Type == EntityHashtag
}

// IsCommand returns true if the type of the message entity is "bot_command".
func (e MessageEntity) IsCommand() bool {
	return e.Type == EntityBotCommand
}

// IsURL returns true if the type of the message entity is "url".
func (e MessageEntity) IsURL() bool {
	return e.Type == EntityURL
}

// IsEmail returns true if the type of the message entity is "email".
func (e MessageEntity) IsEmail() bool {
	return e.Type == EntityEmail
}

// IsBold returns true if the type of the message entity is "bold" (bold text).
func (e MessageEntity) IsBold() bool {
	return e.Type == EntityBold
}

// IsItalic returns true if the type of the message entity is "italic" (italic text).
func (e MessageEntity) IsItalic() bool {
	return e.Type == EntityItalic
}

// IsCode returns true if the type of the message entity is "code" (monowidth string).
func (e MessageEntity) IsCode() bool {
	return e.Type == EntityCode
}

// IsPre returns true if the type of the message entity is "pre" (monowidth block).
func (e MessageEntity) IsPre() bool {
	return e.Type == EntityPre
}

// IsTextLink returns true if the type of the message entity is "text_link" (clickable text URL).
func (e MessageEntity) IsTextLink() bool {
	return e.Type == EntityTextLink
}

// PhotoSize represents one size of a photo or a file / sticker thumbnail.
type PhotoSize struct {
	// FileID identifier for this file, which can be used to download or reuse
	// the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Width photo width
	Width int `json:"width"`
	// Height photo height
	Height int `json:"height"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// Animation represents an animation file.
type Animation struct {
	// FileID is the identifier for this file, which can be used to download or reuse
	// the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Width video width as defined by sender
	Width int `json:"width"`
	// Height video height as defined by sender
	Height int `json:"height"`
	// Duration of the video in seconds as defined by sender
	Duration int `json:"duration"`
	// Thumbnail animation thumbnail as defined by sender
	//
	// optional
	Thumbnail *PhotoSize `json:"thumb,omitempty"`
	// FileName original animation filename as defined by sender
	//
	// optional
	FileName string `json:"file_name,omitempty"`
	// MimeType of the file as defined by sender
	//
	// optional
	MimeType string `json:"mime_type,omitempty"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// Audio represents an audio file to be treated as music by the Telegram clients.
type Audio struct {
	// FileID is an identifier for this file, which can be used to download or
	// reuse the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Duration of the audio in seconds as defined by sender
	Duration int `json:"duration"`
	// Performer of the audio as defined by sender or by audio tags
	//
	// optional
	Performer string `json:"performer,omitempty"`
	// Title of the audio as defined by sender or by audio tags
	//
	// optional
	Title string `json:"title,omitempty"`
	// FileName is the original filename as defined by sender
	//
	// optional
	FileName string `json:"file_name,omitempty"`
	// MimeType of the file as defined by sender
	//
	// optional
	MimeType string `json:"mime_type,omitempty"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
	// Thumbnail is the album cover to which the music file belongs
	//
	// optional
	Thumbnail *PhotoSize `json:"thumb,omitempty"`
}

// Document represents a general file.
type Document struct {
	// FileID is an identifier for this file, which can be used to download or
	// reuse the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Thumbnail document thumbnail as defined by sender
	//
	// optional
	Thumbnail *PhotoSize `json:"thumb,omitempty"`
	// FileName original filename as defined by sender
	//
	// optional
	FileName string `json:"file_name,omitempty"`
	// MimeType  of the file as defined by sender
	//
	// optional
	MimeType string `json:"mime_type,omitempty"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// Video represents a video file.
type Video struct {
	// FileID identifier for this file, which can be used to download or reuse
	// the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Width video width as defined by sender
	Width int `json:"width"`
	// Height video height as defined by sender
	Height int `json:"height"`
	// Duration of the video in seconds as defined by sender
	Duration int `json:"duration"`
	// Thumbnail video thumbnail
	//
	// optional
	Thumbnail *PhotoSize `json:"thumb,omitempty"`
	// FileName is the original filename as defined by sender
	//
	// optional
	FileName string `json:"file_name,omitempty"`
	// MimeType of a file as defined by sender
	//
	// optional
	MimeType string `json:"mime_type,omitempty"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// VideoNote object represents a video message.
type VideoNote struct {
	// FileID identifier for this file, which can be used to download or reuse the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Length video width and height (diameter of the video message) as defined by sender
	Length int `json:"length"`
	// Duration of the video in seconds as defined by sender
	Duration int `json:"duration"`
	// Thumbnail video thumbnail
	//
	// optional
	Thumbnail *PhotoSize `json:"thumb,omitempty"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// Voice represents a voice note.
type Voice struct {
	// FileID identifier for this file, which can be used to download or reuse the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Duration of the audio in seconds as defined by sender
	Duration int `json:"duration"`
	// MimeType of the file as defined by sender
	//
	// optional
	MimeType string `json:"mime_type,omitempty"`
	// FileSize file size
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// Contact represents a phone contact.
//
// Note that LastName and UserID may be empty.
type Contact struct {
	// PhoneNumber contact's phone number
	PhoneNumber string `json:"phone_number"`
	// FirstName contact's first name
	FirstName string `json:"first_name"`
	// LastName contact's last name
	//
	// optional
	LastName string `json:"last_name,omitempty"`
	// UserID contact's user identifier in Telegram
	//
	// optional
	UserID int64 `json:"user_id,omitempty"`
	// VCard is additional data about the contact in the form of a vCard.
	//
	// optional
	VCard string `json:"vcard,omitempty"`
}

// Dice represents an animated emoji that displays a random value.
type Dice struct {
	// Emoji on which the dice throw animation is based
	Emoji string `json:"emoji"`
	// Value of the dice
	Value int `json:"value"`
}

// PollOption contains information about one answer option in a poll.
type PollOption struct {
	// Text is the option text, 1-100 characters
	Text string `json:"text"`
	// VoterCount is the number of users that voted for this option
	VoterCount int `json:"voter_count"`
}

// PollAnswer represents an answer of a user in a non-anonymous poll.
type PollAnswer struct {
	// PollID is the unique poll identifier
	PollID string `json:"poll_id"`
	// User who changed the answer to the poll
	User User `json:"user"`
	// OptionIDs is the 0-based identifiers of poll options chosen by the user.
	// May be empty if user retracted vote.
	OptionIDs []int `json:"option_ids"`
}

// Poll contains information about a poll.
type Poll struct {
	// ID is the unique poll identifier
	ID string `json:"id"`
	// Question is the poll question, 1-255 characters
	Question string `json:"question"`
	// Options is the list of poll options
	Options []PollOption `json:"options"`
	// TotalVoterCount is the total numbers of users who voted in the poll
	TotalVoterCount int `json:"total_voter_count"`
	// IsClosed is if the poll is closed
	IsClosed bool `json:"is_closed"`
	// IsAnonymous is if the poll is anonymous
	IsAnonymous bool `json:"is_anonymous"`
	// Type is the poll type, currently can be "regular" or "quiz"
	Type string `json:"type"`
	// AllowsMultipleAnswers is true, if the poll allows multiple answers
	AllowsMultipleAnswers bool `json:"allows_multiple_answers"`
	// CorrectOptionID is the 0-based identifier of the correct answer option.
	// Available only for polls in quiz mode, which are closed, or was sent (not
	// forwarded) by the bot or to the private chat with the bot.
	//
	// optional
	CorrectOptionID int `json:"correct_option_id,omitempty"`
	// Explanation is text that is shown when a user chooses an incorrect answer
	// or taps on the lamp icon in a quiz-style poll, 0-200 characters
	//
	// optional
	Explanation string `json:"explanation,omitempty"`
	// ExplanationEntities are special entities like usernames, URLs, bot
	// commands, etc. that appear in the explanation
	//
	// optional
	ExplanationEntities []MessageEntity `json:"explanation_entities,omitempty"`
	// OpenPeriod is the amount of time in seconds the poll will be active
	// after creation
	//
	// optional
	OpenPeriod int `json:"open_period,omitempty"`
	// CloseDate is the point in time (unix timestamp) when the poll will be
	// automatically closed
	//
	// optional
	CloseDate int `json:"close_date,omitempty"`
}

// Constants for poll types
const (
	PollTypeRegular = "regular"
	PollTypeQuiz    = "quiz"
)

// IsRegularPoll returns true if the poll is a regular poll
func (p *Poll) IsRegularPoll() bool {
	return p.Type == PollTypeRegular
}

// IsQuizPoll returns true if the poll is a quiz poll
func (p *Poll) IsQuizPoll() bool {
	return p.Type == PollTypeQuiz
}

// Location represents a point on the map.
type Location struct {
	// Longitude as defined by sender
	Longitude float64 `json:"longitude"`
	// Latitude as defined by sender
	Latitude float64 `json:"latitude"`
	// HorizontalAccuracy is the radius of uncertainty for the location,
	// measured in meters; 0-1500
	//
	// optional
	HorizontalAccuracy float64 `json:"horizontal_accuracy,omitempty"`
	// LivePeriod is time relative to the message sending date, during which the
	// location can be updated, in seconds. For active live locations only.
	//
	// optional
	LivePeriod int `json:"live_period,omitempty"`
	// Heading is the direction in which user is moving, in degrees; 1-360. For
	// active live locations only.
	//
	// optional
	Heading int `json:"heading,omitempty"`
	// ProximityAlertRadius is the maximum distance for proximity alerts about
	// approaching another chat member, in meters. For sent live locations only.
	//
	// optional
	ProximityAlertRadius int `json:"proximity_alert_radius,omitempty"`
}

// Venue represents a venue.
type Venue struct {
	// Location is the venue location
	Location Location `json:"location"`
	// Title is the name of the venue
	Title string `json:"title"`
	// Address of the venue
	Address string `json:"address"`
	// FoursquareID is the foursquare identifier of the venue
	//
	// optional
	FoursquareID string `json:"foursquare_id,omitempty"`
	// FoursquareType is the foursquare type of the venue
	//
	// optional
	FoursquareType string `json:"foursquare_type,omitempty"`
	// GooglePlaceID is the Google Places identifier of the venue
	//
	// optional
	GooglePlaceID string `json:"google_place_id,omitempty"`
	// GooglePlaceType is the Google Places type of the venue
	//
	// optional
	GooglePlaceType string `json:"google_place_type,omitempty"`
}

// WebAppData Contains data sent from a Web App to the bot.
type WebAppData struct {
	// Data is the data. Be aware that a bad client can send arbitrary data in this field.
	Data string `json:"data"`
	// ButtonText is the text of the web_app keyboard button, from which the Web App
	// was opened. Be aware that a bad client can send arbitrary data in this field.
	ButtonText string `json:"button_text"`
}

// ProximityAlertTriggered represents a service message sent when a user in the
// chat triggers a proximity alert sent by another user.
type ProximityAlertTriggered struct {
	// Traveler is the user that triggered the alert
	Traveler User `json:"traveler"`
	// Watcher is the user that set the alert
	Watcher User `json:"watcher"`
	// Distance is the distance between the users
	Distance int `json:"distance"`
}

// MessageAutoDeleteTimerChanged represents a service message about a change in
// auto-delete timer settings.
type MessageAutoDeleteTimerChanged struct {
	// New auto-delete time for messages in the chat.
	MessageAutoDeleteTime int `json:"message_auto_delete_time"`
}

// VideoChatScheduled represents a service message about a voice chat scheduled
// in the chat.
type VideoChatScheduled struct {
	// Point in time (Unix timestamp) when the voice chat is supposed to be
	// started by a chat administrator
	StartDate int `json:"start_date"`
}

// Time converts the scheduled start date into a Time.
func (m *VideoChatScheduled) Time() time.Time {
	return time.Unix(int64(m.StartDate), 0)
}

// VideoChatStarted represents a service message about a voice chat started in
// the chat.
type VideoChatStarted struct{}

// VideoChatEnded represents a service message about a voice chat ended in the
// chat.
type VideoChatEnded struct {
	// Voice chat duration; in seconds.
	Duration int `json:"duration"`
}

// VideoChatParticipantsInvited represents a service message about new members
// invited to a voice chat.
type VideoChatParticipantsInvited struct {
	// New members that were invited to the voice chat.
	//
	// optional
	Users []User `json:"users,omitempty"`
}

// UserProfilePhotos contains a set of user profile photos.
type UserProfilePhotos struct {
	// TotalCount total number of profile pictures the target user has
	TotalCount int `json:"total_count"`
	// Photos requested profile pictures (in up to 4 sizes each)
	Photos [][]PhotoSize `json:"photos"`
}

// File contains information about a file to download from Telegram.
type File struct {
	// FileID identifier for this file, which can be used to download or reuse
	// the file
	FileID string `json:"file_id"`
	// FileUniqueID is the unique identifier for this file, which is supposed to
	// be the same over time and for different bots. Can't be used to download
	// or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// FileSize file size, if known
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
	// FilePath file path
	//
	// optional
	FilePath string `json:"file_path,omitempty"`
}

// Link returns a full path to the download URL for a File.
//
// It requires the Bot token to create the link.
func (f *File) Link(token string) string {
	return fmt.Sprintf(FileEndpoint, token, f.FilePath)
}

// WebAppInfo contains information about a Web App.
type WebAppInfo struct {
	// URL is the HTTPS URL of a Web App to be opened with additional data as
	// specified in Initializing Web Apps.
	URL string `json:"url"`
}

// ReplyKeyboardMarkup represents a custom keyboard with reply options.
type ReplyKeyboardMarkup struct {
	// Keyboard is an array of button rows, each represented by an Array of KeyboardButton objects
	Keyboard [][]KeyboardButton `json:"keyboard"`
	// ResizeKeyboard requests clients to resize the keyboard vertically for optimal fit
	// (e.g., make the keyboard smaller if there are just two rows of buttons).
	// Defaults to false, in which case the custom keyboard
	// is always of the same height as the app's standard keyboard.
	//
	// optional
	ResizeKeyboard bool `json:"resize_keyboard,omitempty"`
	// OneTimeKeyboard requests clients to hide the keyboard as soon as it's been used.
	// The keyboard will still be available, but clients will automatically display
	// the usual letter-keyboard in the chat – the user can press a special button
	// in the input field to see the custom keyboard again.
	// Defaults to false.
	//
	// optional
	OneTimeKeyboard bool `json:"one_time_keyboard,omitempty"`
	// InputFieldPlaceholder is the placeholder to be shown in the input field when
	// the keyboard is active; 1-64 characters.
	//
	// optional
	InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
	// Selective use this parameter if you want to show the keyboard to specific users only.
	// Targets:
	//  1) users that are @mentioned in the text of the Message object;
	//  2) if the bot's message is a reply (has Message.ReplyToMessage not nil), sender of the original message.
	//
	// Example: A user requests to change the bot's language,
	// bot replies to the request with a keyboard to select the new language.
	// Other users in the group don't see the keyboard.
	//
	// optional
	Selective bool `json:"selective,omitempty"`
}

// KeyboardButton represents one button of the reply keyboard. For simple text
// buttons String can be used instead of this object to specify text of the
// button. Optional fields request_contact, request_location, and request_poll
// are mutually exclusive.
type KeyboardButton struct {
	// Text of the button. If none of the optional fields are used,
	// it will be sent as a message when the button is pressed.
	Text string `json:"text"`
	// RequestContact if True, the user's phone number will be sent
	// as a contact when the button is pressed.
	// Available in private chats only.
	//
	// optional
	RequestContact bool `json:"request_contact,omitempty"`
	// RequestLocation if True, the user's current location will be sent when
	// the button is pressed.
	// Available in private chats only.
	//
	// optional
	RequestLocation bool `json:"request_location,omitempty"`
	// RequestPoll if specified, the user will be asked to create a poll and send it
	// to the bot when the button is pressed. Available in private chats only
	//
	// optional
	RequestPoll *KeyboardButtonPollType `json:"request_poll,omitempty"`
	// WebApp if specified, the described Web App will be launched when the button
	// is pressed. The Web App will be able to send a "web_app_data" service
	// message. Available in private chats only.
	//
	// optional
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

// KeyboardButtonPollType represents type of poll, which is allowed to
// be created and sent when the corresponding button is pressed.
type KeyboardButtonPollType struct {
	// Type is if quiz is passed, the user will be allowed to create only polls
	// in the quiz mode. If regular is passed, only regular polls will be
	// allowed. Otherwise, the user will be allowed to create a poll of any type.
	Type string `json:"type"`
}

// ReplyKeyboardRemove Upon receiving a message with this object, Telegram
// clients will remove the current custom keyboard and display the default
// letter-keyboard. By default, custom keyboards are displayed until a new
// keyboard is sent by a bot. An exception is made for one-time keyboards
// that are hidden immediately after the user presses a button.
type ReplyKeyboardRemove struct {
	// RemoveKeyboard requests clients to remove the custom keyboard
	// (user will not be able to summon this keyboard;
	// if you want to hide the keyboard from sight but keep it accessible,
	// use one_time_keyboard in ReplyKeyboardMarkup).
	RemoveKeyboard bool `json:"remove_keyboard"`
	// Selective use this parameter if you want to remove the keyboard for specific users only.
	// Targets:
	//  1) users that are @mentioned in the text of the Message object;
	//  2) if the bot's message is a reply (has Message.ReplyToMessage not nil), sender of the original message.
	//
	// Example: A user votes in a poll, bot returns confirmation message
	// in reply to the vote and removes the keyboard for that user,
	// while still showing the keyboard with poll options to users who haven't voted yet.
	//
	// optional
	Selective bool `json:"selective,omitempty"`
}

// InlineKeyboardMarkup represents an inline keyboard that appears right next to
// the message it belongs to.
type InlineKeyboardMarkup struct {
	// InlineKeyboard array of button rows, each represented by an Array of
	// InlineKeyboardButton objects
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents one button of an inline keyboard. You must
// use exactly one of the optional fields.
//
// Note that some values are references as even an empty string
// will change behavior.
//
// CallbackGame, if set, MUST be first button in first row.
type InlineKeyboardButton struct {
	// Text label text on the button
	Text string `json:"text"`
	// URL HTTP or tg:// url to be opened when button is pressed.
	//
	// optional
	URL *string `json:"url,omitempty"`
	// LoginURL is an HTTP URL used to automatically authorize the user. Can be
	// used as a replacement for the Telegram Login Widget
	//
	// optional
	LoginURL *LoginURL `json:"login_url,omitempty"`
	// CallbackData data to be sent in a callback query to the bot when button is pressed, 1-64 bytes.
	//
	// optional
	CallbackData *string `json:"callback_data,omitempty"`
	// WebApp is the Description of the Web App that will be launched when the user presses the button.
	// The Web App will be able to send an arbitrary message on behalf of the user using the method
	// answerWebAppQuery. Available only in private chats between a user and the bot.
	//
	// optional
	WebApp *WebAppInfo `json:"web_app,omitempty"`
	// SwitchInlineQuery if set, pressing the button will prompt the user to select one of their chats,
	// open that chat and insert the bot's username and the specified inline query in the input field.
	// Can be empty, in which case just the bot's username will be inserted.
	//
	// This offers an easy way for users to start using your bot
	// in inline mode when they are currently in a private chat with it.
	// Especially useful when combined with switch_pm… actions – in this case
	// the user will be automatically returned to the chat they switched from,
	// skipping the chat selection screen.
	//
	// optional
	SwitchInlineQuery *string `json:"switch_inline_query,omitempty"`
	// SwitchInlineQueryCurrentChat if set, pressing the button will insert the bot's username
	// and the specified inline query in the current chat's input field.
	// Can be empty, in which case only the bot's username will be inserted.
	//
	// This offers a quick way for the user to open your bot in inline mode
	// in the same chat – good for selecting something from multiple options.
	//
	// optional
	SwitchInlineQueryCurrentChat *string `json:"switch_inline_query_current_chat,omitempty"`
	// CallbackGame description of the game that will be launched when the user presses the button.
	//
	// optional
	CallbackGame *CallbackGame `json:"callback_game,omitempty"`
	// Pay specify True, to send a Pay button.
	//
	// NOTE: This type of button must always be the first button in the first row.
	//
	// optional
	Pay bool `json:"pay,omitempty"`
}

// LoginURL represents a parameter of the inline keyboard button used to
// automatically authorize a user. Serves as a great replacement for the
// Telegram Login Widget when the user is coming from Telegram. All the user
// needs to do is tap/click a button and confirm that they want to log in.
type LoginURL struct {
	// URL is an HTTP URL to be opened with user authorization data added to the
	// query string when the button is pressed. If the user refuses to provide
	// authorization data, the original URL without information about the user
	// will be opened. The data added is the same as described in Receiving
	// authorization data.
	//
	// NOTE: You must always check the hash of the received data to verify the
	// authentication and the integrity of the data as described in Checking
	// authorization.
	URL string `json:"url"`
	// ForwardText is the new text of the button in forwarded messages
	//
	// optional
	ForwardText string `json:"forward_text,omitempty"`
	// BotUsername is the username of a bot, which will be used for user
	// authorization. See Setting up a bot for more details. If not specified,
	// the current bot's username will be assumed. The url's domain must be the
	// same as the domain linked with the bot. See Linking your domain to the
	// bot for more details.
	//
	// optional
	BotUsername string `json:"bot_username,omitempty"`
	// RequestWriteAccess if true requests permission for your bot to send
	// messages to the user
	//
	// optional
	RequestWriteAccess bool `json:"request_write_access,omitempty"`
}

// CallbackQuery represents an incoming callback query from a callback button in
// an inline keyboard. If the button that originated the query was attached to a
// message sent by the bot, the field message will be present. If the button was
// attached to a message sent via the bot (in inline mode), the field
// inline_message_id will be present. Exactly one of the fields data or
// game_short_name will be present.
type CallbackQuery struct {
	// ID unique identifier for this query
	ID string `json:"id"`
	// From sender
	From *User `json:"from"`
	// Message with the callback button that originated the query.
	// Note that message content and message date will not be available if the
	// message is too old.
	//
	// optional
	Message *Message `json:"message,omitempty"`
	// InlineMessageID identifier of the message sent via the bot in inline
	// mode, that originated the query.
	//
	// optional
	InlineMessageID string `json:"inline_message_id,omitempty"`
	// ChatInstance global identifier, uniquely corresponding to the chat to
	// which the message with the callback button was sent. Useful for high
	// scores in games.
	ChatInstance string `json:"chat_instance"`
	// Data associated with the callback button. Be aware that
	// a bad client can send arbitrary data in this field.
	//
	// optional
	Data string `json:"data,omitempty"`
	// GameShortName short name of a Game to be returned, serves as the unique identifier for the game.
	//
	// optional
	GameShortName string `json:"game_short_name,omitempty"`
}

// ForceReply when receiving a message with this object, Telegram clients will
// display a reply interface to the user (act as if the user has selected the
// bot's message and tapped 'Reply'). This can be extremely useful if you  want
// to create user-friendly step-by-step interfaces without having to sacrifice
// privacy mode.
type ForceReply struct {
	// ForceReply shows reply interface to the user,
	// as if they manually selected the bot's message and tapped 'Reply'.
	ForceReply bool `json:"force_reply"`
	// InputFieldPlaceholder is the placeholder to be shown in the input field when
	// the reply is active; 1-64 characters.
	//
	// optional
	InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
	// Selective use this parameter if you want to force reply from specific users only.
	// Targets:
	//  1) users that are @mentioned in the text of the Message object;
	//  2) if the bot's message is a reply (has Message.ReplyToMessage not nil), sender of the original message.
	//
	// optional
	Selective bool `json:"selective,omitempty"`
}

// ChatPhoto represents a chat photo.
type ChatPhoto struct {
	// SmallFileID is a file identifier of small (160x160) chat photo.
	// This file_id can be used only for photo download and
	// only for as long as the photo is not changed.
	SmallFileID string `json:"small_file_id"`
	// SmallFileUniqueID is a unique file identifier of small (160x160) chat
	// photo, which is supposed to be the same over time and for different bots.
	// Can't be used to download or reuse the file.
	SmallFileUniqueID string `json:"small_file_unique_id"`
	// BigFileID is a file identifier of big (640x640) chat photo.
	// This file_id can be used only for photo download and
	// only for as long as the photo is not changed.
	BigFileID string `json:"big_file_id"`
	// BigFileUniqueID is a file identifier of big (640x640) chat photo, which
	// is supposed to be the same over time and for different bots. Can't be
	// used to download or reuse the file.
	BigFileUniqueID string `json:"big_file_unique_id"`
}

// ChatInviteLink represents an invite link for a chat.
type ChatInviteLink struct {
	// InviteLink is the invite link. If the link was created by another chat
	// administrator, then the second part of the link will be replaced with "…".
	InviteLink string `json:"invite_link"`
	// Creator of the link.
	Creator User `json:"creator"`
	// CreatesJoinRequest is true if users joining the chat via the link need to
	// be approved by chat administrators.
	//
	// optional
	CreatesJoinRequest bool `json:"creates_join_request,omitempty"`
	// IsPrimary is true, if the link is primary.
	IsPrimary bool `json:"is_primary"`
	// IsRevoked is true, if the link is revoked.
	IsRevoked bool `json:"is_revoked"`
	// Name is the name of the invite link.
	//
	// optional
	Name string `json:"name,omitempty"`
	// ExpireDate is the point in time (Unix timestamp) when the link will
	// expire or has been expired.
	//
	// optional
	ExpireDate int `json:"expire_date,omitempty"`
	// MemberLimit is the maximum number of users that can be members of the
	// chat simultaneously after joining the chat via this invite link; 1-99999.
	//
	// optional
	MemberLimit int `json:"member_limit,omitempty"`
	// PendingJoinRequestCount is the number of pending join requests created
	// using this link.
	//
	// optional
	PendingJoinRequestCount int `json:"pending_join_request_count,omitempty"`
}

type ChatAdministratorRights struct {
	IsAnonymous         bool `json:"is_anonymous"`
	CanManageChat       bool `json:"can_manage_chat"`
	CanDeleteMessages   bool `json:"can_delete_messages"`
	CanManageVideoChats bool `json:"can_manage_video_chats"`
	CanRestrictMembers  bool `json:"can_restrict_members"`
	CanPromoteMembers   bool `json:"can_promote_members"`
	CanChangeInfo       bool `json:"can_change_info"`
	CanInviteUsers      bool `json:"can_invite_users"`
	CanPostMessages     bool `json:"can_post_messages"`
	CanEditMessages     bool `json:"can_edit_messages"`
	CanPinMessages      bool `json:"can_pin_messages"`
}

// ChatMember contains information about one member of a chat.
type ChatMember struct {
	// User information about the user
	User *User `json:"user"`
	// Status the member's status in the chat.
	// Can be
	//  "creator",
	//  "administrator",
	//  "member",
	//  "restricted",
	//  "left" or
	//  "kicked"
	Status string `json:"status"`
	// CustomTitle owner and administrators only. Custom title for this user
	//
	// optional
	CustomTitle string `json:"custom_title,omitempty"`
	// IsAnonymous owner and administrators only. True, if the user's presence
	// in the chat is hidden
	//
	// optional
	IsAnonymous bool `json:"is_anonymous,omitempty"`
	// UntilDate restricted and kicked only.
	// Date when restrictions will be lifted for this user;
	// unix time.
	//
	// optional
	UntilDate int64 `json:"until_date,omitempty"`
	// CanBeEdited administrators only.
	// True, if the bot is allowed to edit administrator privileges of that user.
	//
	// optional
	CanBeEdited bool `json:"can_be_edited,omitempty"`
	// CanManageChat administrators only.
	// True, if the administrator can access the chat event log, chat
	// statistics, message statistics in channels, see channel members, see
	// anonymous administrators in supergroups and ignore slow mode. Implied by
	// any other administrator privilege.
	//
	// optional
	CanManageChat bool `json:"can_manage_chat,omitempty"`
	// CanPostMessages administrators only.
	// True, if the administrator can post in the channel;
	// channels only.
	//
	// optional
	CanPostMessages bool `json:"can_post_messages,omitempty"`
	// CanEditMessages administrators only.
	// True, if the administrator can edit messages of other users and can pin messages;
	// channels only.
	//
	// optional
	CanEditMessages bool `json:"can_edit_messages,omitempty"`
	// CanDeleteMessages administrators only.
	// True, if the administrator can delete messages of other users.
	//
	// optional
	CanDeleteMessages bool `json:"can_delete_messages,omitempty"`
	// CanManageVideoChats administrators only.
	// True, if the administrator can manage video chats.
	//
	// optional
	CanManageVideoChats bool `json:"can_manage_video_chats,omitempty"`
	// CanRestrictMembers administrators only.
	// True, if the administrator can restrict, ban or unban chat members.
	//
	// optional
	CanRestrictMembers bool `json:"can_restrict_members,omitempty"`
	// CanPromoteMembers administrators only.
	// True, if the administrator can add new administrators
	// with a subset of their own privileges or demote administrators that he has promoted,
	// directly or indirectly (promoted by administrators that were appointed by the user).
	//
	// optional
	CanPromoteMembers bool `json:"can_promote_members,omitempty"`
	// CanChangeInfo administrators and restricted only.
	// True, if the user is allowed to change the chat title, photo and other settings.
	//
	// optional
	CanChangeInfo bool `json:"can_change_info,omitempty"`
	// CanInviteUsers administrators and restricted only.
	// True, if the user is allowed to invite new users to the chat.
	//
	// optional
	CanInviteUsers bool `json:"can_invite_users,omitempty"`
	// CanPinMessages administrators and restricted only.
	// True, if the user is allowed to pin messages; groups and supergroups only
	//
	// optional
	CanPinMessages bool `json:"can_pin_messages,omitempty"`
	// IsMember is true, if the user is a member of the chat at the moment of
	// the request
	IsMember bool `json:"is_member"`
	// CanSendMessages
	//
	// optional
	CanSendMessages bool `json:"can_send_messages,omitempty"`
	// CanSendMediaMessages restricted only.
	// True, if the user is allowed to send text messages, contacts, locations and venues
	//
	// optional
	CanSendMediaMessages bool `json:"can_send_media_messages,omitempty"`
	// CanSendPolls restricted only.
	// True, if the user is allowed to send polls
	//
	// optional
	CanSendPolls bool `json:"can_send_polls,omitempty"`
	// CanSendOtherMessages restricted only.
	// True, if the user is allowed to send audios, documents,
	// photos, videos, video notes and voice notes.
	//
	// optional
	CanSendOtherMessages bool `json:"can_send_other_messages,omitempty"`
	// CanAddWebPagePreviews restricted only.
	// True, if the user is allowed to add web page previews to their messages.
	//
	// optional
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews,omitempty"`
}

// Constants for chat member status
const (
	MemberStatusCreator       = "creator"
	MemberStatusAdministrator = "administrator"
	MemberStatusMember        = "member"
	MemberStatusRestricted    = "restricted"
	MemberStatusLeft          = "left"
	MemberStatusKicked        = "kicked"
)

// IsCreator returns true if the chat member is the creator
func (cm *ChatMember) IsCreator() bool {
	return cm.Status == MemberStatusCreator
}

// IsAdministrator returns true if the chat member is an administrator
func (cm *ChatMember) IsAdministrator() bool {
	return cm.Status == MemberStatusAdministrator
}

// IsRestricted returns true if the chat member is restricted
func (cm *ChatMember) IsRestricted() bool {
	return cm.Status == MemberStatusRestricted
}

// HasLeft returns true if the chat member has left
func (cm *ChatMember) HasLeft() bool {
	return cm.Status == MemberStatusLeft
}

// WasKicked returns true if the chat member was kicked
func (cm *ChatMember) WasKicked() bool {
	return cm.Status == MemberStatusKicked
}

// ChatMemberUpdated represents changes in the status of a chat member.
type ChatMemberUpdated struct {
	// Chat the user belongs to.
	Chat Chat `json:"chat"`
	// From is the performer of the action, which resulted in the change.
	From User `json:"from"`
	// Date the change was done in Unix time.
	Date int `json:"date"`
	// Previous information about the chat member.
	OldChatMember ChatMember `json:"old_chat_member"`
	// New information about the chat member.
	NewChatMember ChatMember `json:"new_chat_member"`
	// InviteLink is the link which was used by the user to join the chat;
	// for joining by invite link events only.
	//
	// optional
	InviteLink *ChatInviteLink `json:"invite_link,omitempty"`
}

// ChatJoinRequest represents a join request sent to a chat.
type ChatJoinRequest struct {
	// Chat to which the request was sent.
	Chat Chat `json:"chat"`
	// User that sent the join request.
	From User `json:"from"`
	// Date the request was sent in Unix time.
	Date int `json:"date"`
	// Bio of the user.
	//
	// optional
	Bio string `json:"bio,omitempty"`
	// InviteLink is the link that was used by the user to send the join request.
	//
	// optional
	InviteLink *ChatInviteLink `json:"invite_link,omitempty"`
}

// ChatPermissions describes actions that a non-administrator user is
// allowed to take in a chat. All fields are optional.
type ChatPermissions struct {
	// CanSendMessages is true, if the user is allowed to send text messages,
	// contacts, locations and venues
	//
	// optional
	CanSendMessages bool `json:"can_send_messages,omitempty"`
	// CanSendMediaMessages is true, if the user is allowed to send audios,
	// documents, photos, videos, video notes and voice notes, implies
	// can_send_messages
	//
	// optional
	CanSendMediaMessages bool `json:"can_send_media_messages,omitempty"`
	// CanSendPolls is true, if the user is allowed to send polls, implies
	// can_send_messages
	//
	// optional
	CanSendPolls bool `json:"can_send_polls,omitempty"`
	// CanSendOtherMessages is true, if the user is allowed to send animations,
	// games, stickers and use inline bots, implies can_send_media_messages
	//
	// optional
	CanSendOtherMessages bool `json:"can_send_other_messages,omitempty"`
	// CanAddWebPagePreviews is true, if the user is allowed to add web page
	// previews to their messages, implies can_send_media_messages
	//
	// optional
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews,omitempty"`
	// CanChangeInfo is true, if the user is allowed to change the chat title,
	// photo and other settings. Ignored in public supergroups
	//
	// optional
	CanChangeInfo bool `json:"can_change_info,omitempty"`
	// CanInviteUsers is true, if the user is allowed to invite new users to the
	// chat
	//
	// optional
	CanInviteUsers bool `json:"can_invite_users,omitempty"`
	// CanPinMessages is true, if the user is allowed to pin messages. Ignored
	// in public supergroups
	//
	// optional
	CanPinMessages bool `json:"can_pin_messages,omitempty"`
}

// ChatLocation represents a location to which a chat is connected.
type ChatLocation struct {
	// Location is the location to which the supergroup is connected. Can't be a
	// live location.
	Location Location `json:"location"`
	// Address is the location address; 1-64 characters, as defined by the chat
	// owner
	Address string `json:"address"`
}

// BotCommand represents a bot command.
type BotCommand struct {
	// Command text of the command, 1-32 characters.
	// Can contain only lowercase English letters, digits and underscores.
	Command string `json:"command"`
	// Description of the command, 3-256 characters.
	Description string `json:"description"`
}

// BotCommandScope represents the scope to which bot commands are applied.
//
// It contains the fields for all types of scopes, different types only support
// specific (or no) fields.
type BotCommandScope struct {
	Type   string `json:"type"`
	ChatID int64  `json:"chat_id,omitempty"`
	UserID int64  `json:"user_id,omitempty"`
}

// MenuButton describes the bot's menu button in a private chat.
type MenuButton struct {
	// Type is the type of menu button, must be one of:
	// - `commands`
	// - `web_app`
	// - `default`
	Type string `json:"type"`
	// Text is the text on the button, for `web_app` type.
	Text string `json:"text,omitempty"`
	// WebApp is the description of the Web App that will be launched when the
	// user presses the button for the `web_app` type.
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

// ResponseParameters are various errors that can be returned in APIResponse.
type ResponseParameters struct {
	// The group has been migrated to a supergroup with the specified identifier.
	//
	// optional
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	// In case of exceeding flood control, the number of seconds left to wait
	// before the request can be repeated.
	//
	// optional
	RetryAfter int `json:"retry_after,omitempty"`
}

// BaseInputMedia is a base type for the InputMedia types.
type BaseInputMedia struct {
	// Type of the result.
	Type string `json:"type"`
	// Media file to send. Pass a file_id to send a file
	// that exists on the Telegram servers (recommended),
	// pass an HTTP URL for Telegram to get a file from the Internet,
	// or pass "attach://<file_attach_name>" to upload a new one
	// using multipart/form-data under <file_attach_name> name.
	Media RequestFileData `json:"media"`
	// thumb intentionally missing as it is not currently compatible

	// Caption of the video to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
}

// InputMediaVideo is a video to send as part of a media group.
type InputMediaVideo struct {
	BaseInputMedia
	// Thumbnail of the file sent; can be ignored if thumbnail generation for
	// the file is supported server-side.
	//
	// optional
	Thumb RequestFileData `json:"thumb,omitempty"`
	// Width video width
	//
	// optional
	Width int `json:"width,omitempty"`
	// Height video height
	//
	// optional
	Height int `json:"height,omitempty"`
	// Duration video duration
	//
	// optional
	Duration int `json:"duration,omitempty"`
	// SupportsStreaming pass True, if the uploaded video is suitable for streaming.
	//
	// optional
	SupportsStreaming bool `json:"supports_streaming,omitempty"`
}

// InputMediaAnimation is an animation to send as part of a media group.
type InputMediaAnimation struct {
	BaseInputMedia
	// Thumbnail of the file sent; can be ignored if thumbnail generation for
	// the file is supported server-side.
	//
	// optional
	Thumb RequestFileData `json:"thumb,omitempty"`
	// Width video width
	//
	// optional
	Width int `json:"width,omitempty"`
	// Height video height
	//
	// optional
	Height int `json:"height,omitempty"`
	// Duration video duration
	//
	// optional
	Duration int `json:"duration,omitempty"`
}

// InputMediaAudio is an audio to send as part of a media group.
type InputMediaAudio struct {
	BaseInputMedia
	// Thumbnail of the file sent; can be ignored if thumbnail generation for
	// the file is supported server-side.
	//
	// optional
	Thumb RequestFileData `json:"thumb,omitempty"`
	// Duration of the audio in seconds
	//
	// optional
	Duration int `json:"duration,omitempty"`
	// Performer of the audio
	//
	// optional
	Performer string `json:"performer,omitempty"`
	// Title of the audio
	//
	// optional
	Title string `json:"title,omitempty"`
}

// InputMediaDocument is a general file to send as part of a media group.
type InputMediaDocument struct {
	BaseInputMedia
	// Thumbnail of the file sent; can be ignored if thumbnail generation for
	// the file is supported server-side.
	//
	// optional
	Thumb RequestFileData `json:"thumb,omitempty"`
	// DisableContentTypeDetection disables automatic server-side content type
	// detection for files uploaded using multipart/form-data. Always true, if
	// the document is sent as part of an album
	//
	// optional
	DisableContentTypeDetection bool `json:"disable_content_type_detection,omitempty"`
}

// Sticker represents a sticker.
type Sticker struct {
	// FileID is an identifier for this file, which can be used to download or
	// reuse the file
	FileID string `json:"file_id"`
	// FileUniqueID is a unique identifier for this file,
	// which is supposed to be the same over time and for different bots.
	// Can't be used to download or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Width sticker width
	Width int `json:"width"`
	// Height sticker height
	Height int `json:"height"`
	// IsAnimated true, if the sticker is animated
	//
	// optional
	IsAnimated bool `json:"is_animated,omitempty"`
	// IsVideo true, if the sticker is a video sticker
	//
	// optional
	IsVideo bool `json:"is_video,omitempty"`
	// Thumbnail sticker thumbnail in the .WEBP or .JPG format
	//
	// optional
	Thumbnail *PhotoSize `json:"thumb,omitempty"`
	// Emoji associated with the sticker
	//
	// optional
	Emoji string `json:"emoji,omitempty"`
	// SetName of the sticker set to which the sticker belongs
	//
	// optional
	SetName string `json:"set_name,omitempty"`
	// PremiumAnimation for premium regular stickers, premium animation for the sticker
	//
	// optional
	PremiumAnimation *File `json:"premium_animation,omitempty"`
	// MaskPosition is for mask stickers, the position where the mask should be
	// placed
	//
	// optional
	MaskPosition *MaskPosition `json:"mask_position,omitempty"`
	// CustomEmojiID for custom emoji stickers, unique identifier of the custom emoji
	//
	// optional
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
	// FileSize
	//
	// optional
	FileSize int `json:"file_size,omitempty"`
}

// StickerSet represents a sticker set.
type StickerSet struct {
	// Name sticker set name
	Name string `json:"name"`
	// Title sticker set title
	Title string `json:"title"`
	// StickerType of stickers in the set, currently one of "regular", "mask", "custom_emoji"
	StickerType string `json:"sticker_type"`
	// IsAnimated true, if the sticker set contains animated stickers
	IsAnimated bool `json:"is_animated"`
	// IsVideo true, if the sticker set contains video stickers
	IsVideo bool `json:"is_video"`
	// ContainsMasks true, if the sticker set contains masks
	ContainsMasks bool `json:"contains_masks"`
	// Stickers list of all set stickers
	Stickers []Sticker `json:"stickers"`
	// Thumb is the sticker set thumbnail in the .WEBP or .TGS format
	Thumbnail *PhotoSize `json:"thumb"`
}

// MaskPosition describes the position on faces where a mask should be placed
// by default.
type MaskPosition struct {
	// The part of the face relative to which the mask should be placed.
	// One of "forehead", "eyes", "mouth", or "chin".
	Point string `json:"point"`
	// Shift by X-axis measured in widths of the mask scaled to the face size,
	// from left to right. For example, choosing -1.0 will place mask just to
	// the left of the default mask position.
	XShift float64 `json:"x_shift"`
	// Shift by Y-axis measured in heights of the mask scaled to the face size,
	// from top to bottom. For example, 1.0 will place the mask just below the
	// default mask position.
	YShift float64 `json:"y_shift"`
	// Mask scaling coefficient. For example, 2.0 means double size.
	Scale float64 `json:"scale"`
}

// Game represents a game. Use BotFather to create and edit games, their short
// names will act as unique identifiers.
type Game struct {
	// Title of the game
	Title string `json:"title"`
	// Description of the game
	Description string `json:"description"`
	// Photo that will be displayed in the game message in chats.
	Photo []PhotoSize `json:"photo"`
	// Text a brief description of the game or high scores included in the game message.
	// Can be automatically edited to include current high scores for the game
	// when the bot calls setGameScore, or manually edited using editMessageText. 0-4096 characters.
	//
	// optional
	Text string `json:"text,omitempty"`
	// TextEntities special entities that appear in text, such as usernames, URLs, bot commands, etc.
	//
	// optional
	TextEntities []MessageEntity `json:"text_entities,omitempty"`
	// Animation is an animation that will be displayed in the game message in chats.
	// Upload via BotFather (https://t.me/botfather).
	//
	// optional
	Animation Animation `json:"animation,omitempty"`
}

// GameHighScore is a user's score and position on the leaderboard.
type GameHighScore struct {
	// Position in high score table for the game
	Position int `json:"position"`
	// User user
	User User `json:"user"`
	// Score score
	Score int `json:"score"`
}

// CallbackGame is for starting a game in an inline keyboard button.
type CallbackGame struct{}

// WebhookInfo is information about a currently set webhook.
type WebhookInfo struct {
	// URL webhook URL, may be empty if webhook is not set up.
	URL string `json:"url"`
	// HasCustomCertificate true, if a custom certificate was provided for webhook certificate checks.
	HasCustomCertificate bool `json:"has_custom_certificate"`
	// PendingUpdateCount number of updates awaiting delivery.
	PendingUpdateCount int `json:"pending_update_count"`
	// IPAddress is the currently used webhook IP address
	//
	// optional
	IPAddress string `json:"ip_address,omitempty"`
	// LastErrorDate unix time for the most recent error
	// that happened when trying to deliver an update via webhook.
	//
	// optional
	LastErrorDate int `json:"last_error_date,omitempty"`
	// LastErrorMessage error message in human-readable format for the most recent error
	// that happened when trying to deliver an update via webhook.
	//
	// optional
	LastErrorMessage string `json:"last_error_message,omitempty"`
	// LastSynchronizationErrorDate is the unix time of the most recent error that
	// happened when trying to synchronize available updates with Telegram datacenters.
	LastSynchronizationErrorDate int `json:"last_synchronization_error_date,omitempty"`
	// MaxConnections maximum allowed number of simultaneous
	// HTTPS connections to the webhook for update delivery.
	//
	// optional
	MaxConnections int `json:"max_connections,omitempty"`
	// AllowedUpdates is a list of update types the bot is subscribed to.
	// Defaults to all update types
	//
	// optional
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

// IsSet returns true if a webhook is currently set.
func (info WebhookInfo) IsSet() bool {
	return info.URL != ""
}

// InlineQuery is a Query from Telegram for an inline request.
type InlineQuery struct {
	// ID unique identifier for this query
	ID string `json:"id"`
	// From sender
	From *User `json:"from"`
	// Query text of the query (up to 256 characters).
	Query string `json:"query"`
	// Offset of the results to be returned, can be controlled by the bot.
	Offset string `json:"offset"`
	// Type of the chat, from which the inline query was sent. Can be either
	// "sender" for a private chat with the inline query sender, "private",
	// "group", "supergroup", or "channel". The chat type should be always known
	// for requests sent from official clients and most third-party clients,
	// unless the request was sent from a secret chat
	//
	// optional
	ChatType ChatType `json:"chat_type,omitempty"`
	// Location sender location, only for bots that request user location.
	//
	// optional
	Location *Location `json:"location,omitempty"`
}

// ChatType represents the type of chat.
type ChatType string

const (
	ChatTypeSender  ChatType = "sender"
	ChatTypeUnknown ChatType = "unknown"
)

// InlineQueryResultCachedAudio is an inline query response with cached audio.
type InlineQueryResultCachedAudio struct {
	// Type of the result, must be audio
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// AudioID a valid file identifier for the audio file
	AudioID string `json:"audio_file_id"`
	// Caption 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the audio
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedDocument is an inline query response with cached document.
type InlineQueryResultCachedDocument struct {
	// Type of the result, must be a document
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// DocumentID a valid file identifier for the file
	DocumentID string `json:"document_file_id"`
	// Title for the result
	//
	// optional
	Title string `json:"title,omitempty"`
	// Caption of the document to be sent, 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// Description short description of the result
	//
	// optional
	Description string `json:"description,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	//	// See formatting options for more details
	//	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the file
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedGIF is an inline query response with cached gif.
type InlineQueryResultCachedGIF struct {
	// Type of the result, must be gif.
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes.
	ID string `json:"id"`
	// GifID a valid file identifier for the GIF file.
	GIFID string `json:"gif_file_id"`
	// Title for the result
	//
	// optional
	Title string `json:"title,omitempty"`
	// Caption of the GIF file to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message.
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the GIF animation.
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedMPEG4GIF is an inline query response with cached
// H.264/MPEG-4 AVC video without sound gif.
type InlineQueryResultCachedMPEG4GIF struct {
	// Type of the result, must be mpeg4_gif
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// MPEG4FileID a valid file identifier for the MP4 file
	MPEG4FileID string `json:"mpeg4_file_id"`
	// Title for the result
	//
	// optional
	Title string `json:"title,omitempty"`
	// Caption of the MPEG-4 file to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message.
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the video animation.
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedPhoto is an inline query response with cached photo.
type InlineQueryResultCachedPhoto struct {
	// Type of the result, must be a photo.
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes.
	ID string `json:"id"`
	// PhotoID a valid file identifier of the photo.
	PhotoID string `json:"photo_file_id"`
	// Title for the result.
	//
	// optional
	Title string `json:"title,omitempty"`
	// Description short description of the result.
	//
	// optional
	Description string `json:"description,omitempty"`
	// Caption of the photo to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the photo caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message.
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the photo.
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedSticker is an inline query response with cached sticker.
type InlineQueryResultCachedSticker struct {
	// Type of the result, must be a sticker
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// StickerID a valid file identifier of the sticker
	StickerID string `json:"sticker_file_id"`
	// Title is a title
	Title string `json:"title"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the sticker
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedVideo is an inline query response with cached video.
type InlineQueryResultCachedVideo struct {
	// Type of the result, must be video
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// VideoID a valid file identifier for the video file
	VideoID string `json:"video_file_id"`
	// Title for the result
	Title string `json:"title"`
	// Description short description of the result
	//
	// optional
	Description string `json:"description,omitempty"`
	// Caption of the video to be sent, 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the video
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultCachedVoice is an inline query response with cached voice.
type InlineQueryResultCachedVoice struct {
	// Type of the result, must be voice
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// VoiceID a valid file identifier for the voice message
	VoiceID string `json:"voice_file_id"`
	// Title voice message title
	Title string `json:"title"`
	// Caption 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the voice message
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultArticle represents a link to an article or web page.
type InlineQueryResultArticle struct {
	// Type of the result, must be article.
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 Bytes.
	ID string `json:"id"`
	// Title of the result
	Title string `json:"title"`
	// InputMessageContent content of the message to be sent.
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
	// ReplyMarkup Inline keyboard attached to the message.
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// URL of the result.
	//
	// optional
	URL string `json:"url,omitempty"`
	// HideURL pass True, if you don't want the URL to be shown in the message.
	//
	// optional
	HideURL bool `json:"hide_url,omitempty"`
	// Description short description of the result.
	//
	// optional
	Description string `json:"description,omitempty"`
	// ThumbURL url of the thumbnail for the result
	//
	// optional
	ThumbURL string `json:"thumb_url,omitempty"`
	// ThumbWidth thumbnail width
	//
	// optional
	ThumbWidth int `json:"thumb_width,omitempty"`
	// ThumbHeight thumbnail height
	//
	// optional
	ThumbHeight int `json:"thumb_height,omitempty"`
}

// InlineQueryResultAudio is an inline query response audio.
type InlineQueryResultAudio struct {
	// Type of the result, must be audio
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// URL a valid url for the audio file
	URL string `json:"audio_url"`
	// Title is a title
	Title string `json:"title"`
	// Caption 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// Performer is a performer
	//
	// optional
	Performer string `json:"performer,omitempty"`
	// Duration audio duration in seconds
	//
	// optional
	Duration int `json:"audio_duration,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the audio
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultContact is an inline query response contact.
type InlineQueryResultContact struct {
	Type                string                `json:"type"`         // required
	ID                  string                `json:"id"`           // required
	PhoneNumber         string                `json:"phone_number"` // required
	FirstName           string                `json:"first_name"`   // required
	LastName            string                `json:"last_name"`
	VCard               string                `json:"vcard"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	InputMessageContent interface{}           `json:"input_message_content,omitempty"`
	ThumbURL            string                `json:"thumb_url"`
	ThumbWidth          int                   `json:"thumb_width"`
	ThumbHeight         int                   `json:"thumb_height"`
}

// InlineQueryResultGame is an inline query response game.
type InlineQueryResultGame struct {
	// Type of the result, must be game
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// GameShortName short name of the game
	GameShortName string `json:"game_short_name"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// InlineQueryResultDocument is an inline query response document.
type InlineQueryResultDocument struct {
	// Type of the result, must be a document
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// Title for the result
	Title string `json:"title"`
	// Caption of the document to be sent, 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// URL a valid url for the file
	URL string `json:"document_url"`
	// MimeType of the content of the file, either "application/pdf" or "application/zip"
	MimeType string `json:"mime_type"`
	// Description short description of the result
	//
	// optional
	Description string `json:"description,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the file
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
	// ThumbURL url of the thumbnail (jpeg only) for the file
	//
	// optional
	ThumbURL string `json:"thumb_url,omitempty"`
	// ThumbWidth thumbnail width
	//
	// optional
	ThumbWidth int `json:"thumb_width,omitempty"`
	// ThumbHeight thumbnail height
	//
	// optional
	ThumbHeight int `json:"thumb_height,omitempty"`
}

// InlineQueryResultGIF is an inline query response GIF.
type InlineQueryResultGIF struct {
	// Type of the result, must be gif.
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes.
	ID string `json:"id"`
	// URL a valid URL for the GIF file. File size must not exceed 1MB.
	URL string `json:"gif_url"`
	// ThumbURL url of the static (JPEG or GIF) or animated (MPEG4) thumbnail for the result.
	ThumbURL string `json:"thumb_url"`
	// Width of the GIF
	//
	// optional
	Width int `json:"gif_width,omitempty"`
	// Height of the GIF
	//
	// optional
	Height int `json:"gif_height,omitempty"`
	// Duration of the GIF
	//
	// optional
	Duration int `json:"gif_duration,omitempty"`
	// Title for the result
	//
	// optional
	Title string `json:"title,omitempty"`
	// Caption of the GIF file to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the GIF animation.
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultLocation is an inline query response location.
type InlineQueryResultLocation struct {
	// Type of the result, must be location
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 Bytes
	ID string `json:"id"`
	// Latitude  of the location in degrees
	Latitude float64 `json:"latitude"`
	// Longitude of the location in degrees
	Longitude float64 `json:"longitude"`
	// Title of the location
	Title string `json:"title"`
	// HorizontalAccuracy is the radius of uncertainty for the location,
	// measured in meters; 0-1500
	//
	// optional
	HorizontalAccuracy float64 `json:"horizontal_accuracy,omitempty"`
	// LivePeriod is the period in seconds for which the location can be
	// updated, should be between 60 and 86400.
	//
	// optional
	LivePeriod int `json:"live_period,omitempty"`
	// Heading is for live locations, a direction in which the user is moving,
	// in degrees. Must be between 1 and 360 if specified.
	//
	// optional
	Heading int `json:"heading,omitempty"`
	// ProximityAlertRadius is for live locations, a maximum distance for
	// proximity alerts about approaching another chat member, in meters. Must
	// be between 1 and 100000 if specified.
	//
	// optional
	ProximityAlertRadius int `json:"proximity_alert_radius,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the location
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
	// ThumbURL url of the thumbnail for the result
	//
	// optional
	ThumbURL string `json:"thumb_url,omitempty"`
	// ThumbWidth thumbnail width
	//
	// optional
	ThumbWidth int `json:"thumb_width,omitempty"`
	// ThumbHeight thumbnail height
	//
	// optional
	ThumbHeight int `json:"thumb_height,omitempty"`
}

// InlineQueryResultMPEG4GIF is an inline query response MPEG4 GIF.
type InlineQueryResultMPEG4GIF struct {
	// Type of the result, must be mpeg4_gif
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// URL a valid URL for the MP4 file. File size must not exceed 1MB
	URL string `json:"mpeg4_url"`
	// Width video width
	//
	// optional
	Width int `json:"mpeg4_width,omitempty"`
	// Height vVideo height
	//
	// optional
	Height int `json:"mpeg4_height,omitempty"`
	// Duration video duration
	//
	// optional
	Duration int `json:"mpeg4_duration,omitempty"`
	// ThumbURL url of the static (JPEG or GIF) or animated (MPEG4) thumbnail for the result.
	ThumbURL string `json:"thumb_url"`
	// Title for the result
	//
	// optional
	Title string `json:"title,omitempty"`
	// Caption of the MPEG-4 file to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the video animation
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultPhoto is an inline query response photo.
type InlineQueryResultPhoto struct {
	// Type of the result, must be article.
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 Bytes.
	ID string `json:"id"`
	// URL a valid URL of the photo. Photo must be in jpeg format.
	// Photo size must not exceed 5MB.
	URL string `json:"photo_url"`
	// MimeType
	MimeType string `json:"mime_type"`
	// Width of the photo
	//
	// optional
	Width int `json:"photo_width,omitempty"`
	// Height of the photo
	//
	// optional
	Height int `json:"photo_height,omitempty"`
	// ThumbURL url of the thumbnail for the photo.
	//
	// optional
	ThumbURL string `json:"thumb_url,omitempty"`
	// Title for the result
	//
	// optional
	Title string `json:"title,omitempty"`
	// Description short description of the result
	//
	// optional
	Description string `json:"description,omitempty"`
	// Caption of the photo to be sent, 0-1024 characters after entities parsing.
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the photo caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// ReplyMarkup inline keyboard attached to the message.
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// InputMessageContent content of the message to be sent instead of the photo.
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultVenue is an inline query response venue.
type InlineQueryResultVenue struct {
	// Type of the result, must be venue
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 Bytes
	ID string `json:"id"`
	// Latitude of the venue location in degrees
	Latitude float64 `json:"latitude"`
	// Longitude of the venue location in degrees
	Longitude float64 `json:"longitude"`
	// Title of the venue
	Title string `json:"title"`
	// Address of the venue
	Address string `json:"address"`
	// FoursquareID foursquare identifier of the venue if known
	//
	// optional
	FoursquareID string `json:"foursquare_id,omitempty"`
	// FoursquareType foursquare type of the venue, if known.
	// (For example, "arts_entertainment/default", "arts_entertainment/aquarium" or "food/icecream.".)
	//
	// optional
	FoursquareType string `json:"foursquare_type,omitempty"`
	// GooglePlaceID is the Google Places identifier of the venue
	//
	// optional
	GooglePlaceID string `json:"google_place_id,omitempty"`
	// GooglePlaceType is the Google Places type of the venue
	//
	// optional
	GooglePlaceType string `json:"google_place_type,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the venue
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
	// ThumbURL url of the thumbnail for the result
	//
	// optional
	ThumbURL string `json:"thumb_url,omitempty"`
	// ThumbWidth thumbnail width
	//
	// optional
	ThumbWidth int `json:"thumb_width,omitempty"`
	// ThumbHeight thumbnail height
	//
	// optional
	ThumbHeight int `json:"thumb_height,omitempty"`
}

// InlineQueryResultVideo is an inline query response video.
type InlineQueryResultVideo struct {
	// Type of the result, must be video
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// URL a valid url for the embedded video player or video file
	URL string `json:"video_url"`
	// MimeType of the content of video url, "text/html" or "video/mp4"
	MimeType string `json:"mime_type"`
	//
	// ThumbURL url of the thumbnail (jpeg only) for the video
	// optional
	ThumbURL string `json:"thumb_url,omitempty"`
	// Title for the result
	Title string `json:"title"`
	// Caption of the video to be sent, 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// Width video width
	//
	// optional
	Width int `json:"video_width,omitempty"`
	// Height video height
	//
	// optional
	Height int `json:"video_height,omitempty"`
	// Duration video duration in seconds
	//
	// optional
	Duration int `json:"video_duration,omitempty"`
	// Description short description of the result
	//
	// optional
	Description string `json:"description,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the video.
	// This field is required if InlineQueryResultVideo is used to send
	// an HTML-page as a result (e.g., a YouTube video).
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// InlineQueryResultVoice is an inline query response voice.
type InlineQueryResultVoice struct {
	// Type of the result, must be voice
	Type string `json:"type"`
	// ID unique identifier for this result, 1-64 bytes
	ID string `json:"id"`
	// URL a valid URL for the voice recording
	URL string `json:"voice_url"`
	// Title recording title
	Title string `json:"title"`
	// Caption 0-1024 characters after entities parsing
	//
	// optional
	Caption string `json:"caption,omitempty"`
	// ParseMode mode for parsing entities in the video caption.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// CaptionEntities is a list of special entities that appear in the caption,
	// which can be specified instead of parse_mode
	//
	// optional
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// Duration recording duration in seconds
	//
	// optional
	Duration int `json:"voice_duration,omitempty"`
	// ReplyMarkup inline keyboard attached to the message
	//
	// optional
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	// InputMessageContent content of the message to be sent instead of the voice recording
	//
	// optional
	InputMessageContent interface{} `json:"input_message_content,omitempty"`
}

// ChosenInlineResult is an inline query result chosen by a User
type ChosenInlineResult struct {
	// ResultID the unique identifier for the result that was chosen
	ResultID string `json:"result_id"`
	// From the user that chose the result
	From *User `json:"from"`
	// Location sender location, only for bots that require user location
	//
	// optional
	Location *Location `json:"location,omitempty"`
	// InlineMessageID identifier of the sent inline message.
	// Available only if there is an inline keyboard attached to the message.
	// Will be also received in callback queries and can be used to edit the message.
	//
	// optional
	InlineMessageID string `json:"inline_message_id,omitempty"`
	// Query the query that was used to obtain the result
	Query string `json:"query"`
}

// SentWebAppMessage contains information about an inline message sent by a Web App
// on behalf of a user.
type SentWebAppMessage struct {
	// Identifier of the sent inline message. Available only if there is an inline
	// keyboard attached to the message.
	//
	// optional
	InlineMessageID string `json:"inline_message_id,omitempty"`
}

// InputTextMessageContent contains text for displaying
// as an inline query result.
type InputTextMessageContent struct {
	// Text of the message to be sent, 1-4096 characters
	Text string `json:"message_text"`
	// ParseMode mode for parsing entities in the message text.
	// See formatting options for more details
	// (https://core.telegram.org/bots/api#formatting-options).
	//
	// optional
	ParseMode string `json:"parse_mode,omitempty"`
	// Entities is a list of special entities that appear in message text, which
	// can be specified instead of parse_mode
	//
	// optional
	Entities []MessageEntity `json:"entities,omitempty"`
	// DisableWebPagePreview disables link previews for links in the sent message
	//
	// optional
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`
}

// InputLocationMessageContent contains a location for displaying
// as an inline query result.
type InputLocationMessageContent struct {
	// Latitude of the location in degrees
	Latitude float64 `json:"latitude"`
	// Longitude of the location in degrees
	Longitude float64 `json:"longitude"`
	// HorizontalAccuracy is the radius of uncertainty for the location,
	// measured in meters; 0-1500
	//
	// optional
	HorizontalAccuracy float64 `json:"horizontal_accuracy,omitempty"`
	// LivePeriod is the period in seconds for which the location can be
	// updated, should be between 60 and 86400
	//
	// optional
	LivePeriod int `json:"live_period,omitempty"`
	// Heading is for live locations, a direction in which the user is moving,
	// in degrees. Must be between 1 and 360 if specified.
	//
	// optional
	Heading int `json:"heading,omitempty"`
	// ProximityAlertRadius is for live locations, a maximum distance for
	// proximity alerts about approaching another chat member, in meters. Must
	// be between 1 and 100000 if specified.
	//
	// optional
	ProximityAlertRadius int `json:"proximity_alert_radius,omitempty"`
}

// InputVenueMessageContent contains a venue for displaying
// as an inline query result.
type InputVenueMessageContent struct {
	// Latitude of the venue in degrees
	Latitude float64 `json:"latitude"`
	// Longitude of the venue in degrees
	Longitude float64 `json:"longitude"`
	// Title name of the venue
	Title string `json:"title"`
	// Address of the venue
	Address string `json:"address"`
	// FoursquareID foursquare identifier of the venue, if known
	//
	// optional
	FoursquareID string `json:"foursquare_id,omitempty"`
	// FoursquareType Foursquare type of the venue, if known
	//
	// optional
	FoursquareType string `json:"foursquare_type,omitempty"`
	// GooglePlaceID is the Google Places identifier of the venue
	//
	// optional
	GooglePlaceID string `json:"google_place_id,omitempty"`
	// GooglePlaceType is the Google Places type of the venue
	//
	// optional
	GooglePlaceType string `json:"google_place_type,omitempty"`
}

// InputContactMessageContent contains a contact for displaying
// as an inline query result.
type InputContactMessageContent struct {
	// 	PhoneNumber contact's phone number
	PhoneNumber string `json:"phone_number"`
	// FirstName contact's first name
	FirstName string `json:"first_name"`
	// LastName contact's last name
	//
	// optional
	LastName string `json:"last_name,omitempty"`
	// Additional data about the contact in the form of a vCard
	//
	// optional
	VCard string `json:"vcard,omitempty"`
}

// InputInvoiceMessageContent represents the content of an invoice message to be
// sent as the result of an inline query.
type InputInvoiceMessageContent struct {
	// Product name, 1-32 characters
	Title string `json:"title"`
	// Product description, 1-255 characters
	Description string `json:"description"`
	// Bot-defined invoice payload, 1-128 bytes. This will not be displayed to
	// the user, use for your internal processes.
	Payload string `json:"payload"`
	// Payment provider token, obtained via Botfather
	ProviderToken string `json:"provider_token"`
	// Three-letter ISO 4217 currency code
	Currency string `json:"currency"`
	// Price breakdown, a JSON-serialized list of components (e.g. product
	// price, tax, discount, delivery cost, delivery tax, bonus, etc.)
	Prices []LabeledPrice `json:"prices"`
	// The maximum accepted amount for tips in the smallest units of the
	// currency (integer, not float/double).
	//
	// optional
	MaxTipAmount int `json:"max_tip_amount,omitempty"`
	// An array of suggested amounts of tip in the smallest units of the
	// currency (integer, not float/double). At most 4 suggested tip amounts can
	// be specified. The suggested tip amounts must be positive, passed in a
	// strictly increased order and must not exceed max_tip_amount.
	//
	// optional
	SuggestedTipAmounts []int `json:"suggested_tip_amounts,omitempty"`
	// A JSON-serialized object for data about the invoice, which will be shared
	// with the payment provider. A detailed description of the required fields
	// should be provided by the payment provider.
	//
	// optional
	ProviderData string `json:"provider_data,omitempty"`
	// URL of the product photo for the invoice. Can be a photo of the goods or
	// a marketing image for a service. People like it better when they see what
	// they are paying for.
	//
	// optional
	PhotoURL string `json:"photo_url,omitempty"`
	// Photo size
	//
	// optional
	PhotoSize int `json:"photo_size,omitempty"`
	// Photo width
	//
	// optional
	PhotoWidth int `json:"photo_width,omitempty"`
	// Photo height
	//
	// optional
	PhotoHeight int `json:"photo_height,omitempty"`
	// Pass True, if you require the user's full name to complete the order
	//
	// optional
	NeedName bool `json:"need_name,omitempty"`
	// Pass True, if you require the user's phone number to complete the order
	//
	// optional
	NeedPhoneNumber bool `json:"need_phone_number,omitempty"`
	// Pass True, if you require the user's email address to complete the order
	//
	// optional
	NeedEmail bool `json:"need_email,omitempty"`
	// Pass True, if you require the user's shipping address to complete the order
	//
	// optional
	NeedShippingAddress bool `json:"need_shipping_address,omitempty"`
	// Pass True, if user's phone number should be sent to provider
	//
	// optional
	SendPhoneNumberToProvider bool `json:"send_phone_number_to_provider,omitempty"`
	// Pass True, if user's email address should be sent to provider
	//
	// optional
	SendEmailToProvider bool `json:"send_email_to_provider,omitempty"`
	// Pass True, if the final price depends on the shipping method
	//
	// optional
	IsFlexible bool `json:"is_flexible,omitempty"`
}

// LabeledPrice represents a portion of the price for goods or services.
type LabeledPrice struct {
	// Label portion label
	Label string `json:"label"`
	// Amount price of the product in the smallest units of the currency (integer, not float/double).
	// For example, for a price of US$ 1.45 pass amount = 145.
	// See the exp parameter in currencies.json
	// (https://core.telegram.org/bots/payments/currencies.json),
	// it shows the number of digits past the decimal point
	// for each currency (2 for the majority of currencies).
	Amount int `json:"amount"`
}

// Invoice contains basic information about an invoice.
type Invoice struct {
	// Title product name
	Title string `json:"title"`
	// Description product description
	Description string `json:"description"`
	// StartParameter unique bot deep-linking parameter that can be used to generate this invoice
	StartParameter string `json:"start_parameter"`
	// Currency three-letter ISO 4217 currency code
	// (see https://core.telegram.org/bots/payments#supported-currencies)
	Currency string `json:"currency"`
	// TotalAmount total price in the smallest units of the currency (integer, not float/double).
	// For example, for a price of US$ 1.45 pass amount = 145.
	// See the exp parameter in currencies.json
	// (https://core.telegram.org/bots/payments/currencies.json),
	// it shows the number of digits past the decimal point
	// for each currency (2 for the majority of currencies).
	TotalAmount int `json:"total_amount"`
}

// ShippingAddress represents a shipping address.
type ShippingAddress struct {
	// CountryCode ISO 3166-1 alpha-2 country code
	CountryCode string `json:"country_code"`
	// State if applicable
	State string `json:"state"`
	// City city
	City string `json:"city"`
	// StreetLine1 first line for the address
	StreetLine1 string `json:"street_line1"`
	// StreetLine2 second line for the address
	StreetLine2 string `json:"street_line2"`
	// PostCode address post code
	PostCode string `json:"post_code"`
}

// OrderInfo represents information about an order.
type OrderInfo struct {
	// Name user name
	//
	// optional
	Name string `json:"name,omitempty"`
	// PhoneNumber user's phone number
	//
	// optional
	PhoneNumber string `json:"phone_number,omitempty"`
	// Email user email
	//
	// optional
	Email string `json:"email,omitempty"`
	// ShippingAddress user shipping address
	//
	// optional
	ShippingAddress *ShippingAddress `json:"shipping_address,omitempty"`
}

// ShippingOption represents one shipping option.
type ShippingOption struct {
	// ID shipping option identifier
	ID string `json:"id"`
	// Title option title
	Title string `json:"title"`
	// Prices list of price portions
	Prices []LabeledPrice `json:"prices"`
}

// SuccessfulPayment contains basic information about a successful payment.
type SuccessfulPayment struct {
	// Currency three-letter ISO 4217 currency code
	// (see https://core.telegram.org/bots/payments#supported-currencies)
	Currency string `json:"currency"`
	// TotalAmount total price in the smallest units of the currency (integer, not float/double).
	// For example, for a price of US$ 1.45 pass amount = 145.
	// See the exp parameter in currencies.json,
	// (https://core.telegram.org/bots/payments/currencies.json)
	// it shows the number of digits past the decimal point
	// for each currency (2 for the majority of currencies).
	TotalAmount int `json:"total_amount"`
	// InvoicePayload bot specified invoice payload
	InvoicePayload string `json:"invoice_payload"`
	// ShippingOptionID identifier of the shipping option chosen by the user
	//
	// optional
	ShippingOptionID string `json:"shipping_option_id,omitempty"`
	// OrderInfo order info provided by the user
	//
	// optional
	OrderInfo *OrderInfo `json:"order_info,omitempty"`
	// TelegramPaymentChargeID telegram payment identifier
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	// ProviderPaymentChargeID provider payment identifier
	ProviderPaymentChargeID string `json:"provider_payment_charge_id"`
}

// ShippingQuery contains information about an incoming shipping query.
type ShippingQuery struct {
	// ID unique query identifier
	ID string `json:"id"`
	// From user who sent the query
	From *User `json:"from"`
	// InvoicePayload bot specified invoice payload
	InvoicePayload string `json:"invoice_payload"`
	// ShippingAddress user specified shipping address
	ShippingAddress *ShippingAddress `json:"shipping_address"`
}

// PreCheckoutQuery contains information about an incoming pre-checkout query.
type PreCheckoutQuery struct {
	// ID unique query identifier
	ID string `json:"id"`
	// From user who sent the query
	From *User `json:"from"`
	// Currency three-letter ISO 4217 currency code
	//	// (see https://core.telegram.org/bots/payments#supported-currencies)
	Currency string `json:"currency"`
	// TotalAmount total price in the smallest units of the currency (integer, not float/double).
	//	// For example, for a price of US$ 1.45 pass amount = 145.
	//	// See the exp parameter in currencies.json,
	//	// (https://core.telegram.org/bots/payments/currencies.json)
	//	// it shows the number of digits past the decimal point
	//	// for each currency (2 for the majority of currencies).
	TotalAmount int `json:"total_amount"`
	// InvoicePayload bot specified invoice payload
	InvoicePayload string `json:"invoice_payload"`
	// ShippingOptionID identifier of the shipping option chosen by the user
	//
	// optional
	ShippingOptionID string `json:"shipping_option_id,omitempty"`
	// OrderInfo order info provided by the user
	//
	// optional
	OrderInfo *OrderInfo `json:"order_info,omitempty"`
}

// PassportRequestInfoConfig allows you to request passport info
type PassportRequestInfoConfig struct {
	BotID     int            `json:"bot_id"`
	Scope     *PassportScope `json:"scope"`
	Nonce     string         `json:"nonce"`
	PublicKey string         `json:"public_key"`
}

// PassportScopeElement supports using one or one of several elements.
type PassportScopeElement interface {
	ScopeType() string
}

// PassportScope is the requested scopes of data.
type PassportScope struct {
	V    int                    `json:"v"`
	Data []PassportScopeElement `json:"data"`
}

// PassportScopeElementOneOfSeveral allows you to request any one of the
// requested documents.
type PassportScopeElementOneOfSeveral struct {
}

// ScopeType is the scope type.
func (eo *PassportScopeElementOneOfSeveral) ScopeType() string {
	return "one_of"
}

// PassportScopeElementOne requires the specified element be provided.
type PassportScopeElementOne struct {
	Type        string `json:"type"` // One of "personal_details", "passport", "driver_license", "identity_card", "internal_passport", "address", "utility_bill", "bank_statement", "rental_agreement", "passport_registration", "temporary_registration", "phone_number", "email"
	Selfie      bool   `json:"selfie"`
	Translation bool   `json:"translation"`
	NativeNames bool   `json:"native_name"`
}

// ScopeType is the scope type.
func (eo *PassportScopeElementOne) ScopeType() string {
	return "one"
}

type (
	// PassportData contains information about Telegram Passport data shared with
	// the bot by the user.
	PassportData struct {
		// Array with information about documents and other Telegram Passport
		// elements that was shared with the bot
		Data []EncryptedPassportElement `json:"data"`

		// Encrypted credentials required to decrypt the data
		Credentials *EncryptedCredentials `json:"credentials"`
	}

	// PassportFile represents a file uploaded to Telegram Passport. Currently, all
	// Telegram Passport files are in JPEG format when decrypted and don't exceed
	// 10MB.
	PassportFile struct {
		// Unique identifier for this file
		FileID string `json:"file_id"`

		FileUniqueID string `json:"file_unique_id"`

		// File size
		FileSize int `json:"file_size"`

		// Unix time when the file was uploaded
		FileDate int64 `json:"file_date"`
	}

	// EncryptedPassportElement contains information about documents or other
	// Telegram Passport elements shared with the bot by the user.
	EncryptedPassportElement struct {
		// Element type.
		Type string `json:"type"`

		// Base64-encoded encrypted Telegram Passport element data provided by
		// the user, available for "personal_details", "passport",
		// "driver_license", "identity_card", "identity_passport" and "address"
		// types. Can be decrypted and verified using the accompanying
		// EncryptedCredentials.
		Data string `json:"data,omitempty"`

		// User's verified phone number, available only for "phone_number" type
		PhoneNumber string `json:"phone_number,omitempty"`

		// User's verified email address, available only for "email" type
		Email string `json:"email,omitempty"`

		// Array of encrypted files with documents provided by the user,
		// available for "utility_bill", "bank_statement", "rental_agreement",
		// "passport_registration" and "temporary_registration" types. Files can
		// be decrypted and verified using the accompanying EncryptedCredentials.
		Files []PassportFile `json:"files,omitempty"`

		// Encrypted file with the front side of the document, provided by the
		// user. Available for "passport", "driver_license", "identity_card" and
		// "internal_passport". The file can be decrypted and verified using the
		// accompanying EncryptedCredentials.
		FrontSide *PassportFile `json:"front_side,omitempty"`

		// Encrypted file with the reverse side of the document, provided by the
		// user. Available for "driver_license" and "identity_card". The file can
		// be decrypted and verified using the accompanying EncryptedCredentials.
		ReverseSide *PassportFile `json:"reverse_side,omitempty"`

		// Encrypted file with the selfie of the user holding a document,
		// provided by the user; available for "passport", "driver_license",
		// "identity_card" and "internal_passport". The file can be decrypted
		// and verified using the accompanying EncryptedCredentials.
		Selfie *PassportFile `json:"selfie,omitempty"`
	}

	// EncryptedCredentials contains data required for decrypting and
	// authenticating EncryptedPassportElement. See the Telegram Passport
	// Documentation for a complete description of the data decryption and
	// authentication processes.
	EncryptedCredentials struct {
		// Base64-encoded encrypted JSON-serialized data with unique user's
		// payload, data hashes and secrets required for EncryptedPassportElement
		// decryption and authentication
		Data string `json:"data"`

		// Base64-encoded data hash for data authentication
		Hash string `json:"hash"`

		// Base64-encoded secret, encrypted with the bot's public RSA key,
		// required for data decryption
		Secret string `json:"secret"`
	}

	// PassportElementError represents an error in the Telegram Passport element
	// which was submitted that should be resolved by the user.
	PassportElementError interface{}

	// PassportElementErrorDataField represents an issue in one of the data
	// fields that was provided by the user. The error is considered resolved
	// when the field's value changes.
	PassportElementErrorDataField struct {
		// Error source, must be data
		Source string `json:"source"`

		// The section of the user's Telegram Passport which has the error, one
		// of "personal_details", "passport", "driver_license", "identity_card",
		// "internal_passport", "address"
		Type string `json:"type"`

		// Name of the data field which has the error
		FieldName string `json:"field_name"`

		// Base64-encoded data hash
		DataHash string `json:"data_hash"`

		// Error message
		Message string `json:"message"`
	}

	// PassportElementErrorFrontSide represents an issue with the front side of
	// a document. The error is considered resolved when the file with the front
	// side of the document changes.
	PassportElementErrorFrontSide struct {
		// Error source, must be front_side
		Source string `json:"source"`

		// The section of the user's Telegram Passport which has the issue, one
		// of "passport", "driver_license", "identity_card", "internal_passport"
		Type string `json:"type"`

		// Base64-encoded hash of the file with the front side of the document
		FileHash string `json:"file_hash"`

		// Error message
		Message string `json:"message"`
	}

	// PassportElementErrorReverseSide represents an issue with the reverse side
	// of a document. The error is considered resolved when the file with reverse
	// side of the document changes.
	PassportElementErrorReverseSide struct {
		// Error source, must be reverse_side
		Source string `json:"source"`

		// The section of the user's Telegram Passport which has the issue, one
		// of "driver_license", "identity_card"
		Type string `json:"type"`

		// Base64-encoded hash of the file with the reverse side of the document
		FileHash string `json:"file_hash"`

		// Error message
		Message string `json:"message"`
	}

	// PassportElementErrorSelfie represents an issue with the selfie with a
	// document. The error is considered resolved when the file with the selfie
	// changes.
	PassportElementErrorSelfie struct {
		// Error source, must be selfie
		Source string `json:"source"`

		// The section of the user's Telegram Passport which has the issue, one
		// of "passport", "driver_license", "identity_card", "internal_passport"
		Type string `json:"type"`

		// Base64-encoded hash of the file with the selfie
		FileHash string `json:"file_hash"`

		// Error message
		Message string `json:"message"`
	}

	// PassportElementErrorFile represents an issue with a document scan. The
	// error is considered resolved when the file with the document scan changes.
	PassportElementErrorFile struct {
		// Error source, must be a file
		Source string `json:"source"`

		// The section of the user's Telegram Passport which has the issue, one
		// of "utility_bill", "bank_statement", "rental_agreement",
		// "passport_registration", "temporary_registration"
		Type string `json:"type"`

		// Base64-encoded file hash
		FileHash string `json:"file_hash"`

		// Error message
		Message string `json:"message"`
	}

	// PassportElementErrorFiles represents an issue with a list of scans. The
	// error is considered resolved when the list of files containing the scans
	// changes.
	PassportElementErrorFiles struct {
		// Error source, must be files
		Source string `json:"source"`

		// The section of the user's Telegram Passport which has the issue, one
		// of "utility_bill", "bank_statement", "rental_agreement",
		// "passport_registration", "temporary_registration"
		Type string `json:"type"`

		// List of base64-encoded file hashes
		FileHashes []string `json:"file_hashes"`

		// Error message
		Message string `json:"message"`
	}

	// Credentials contains encrypted data.
	Credentials struct {
		Data SecureData `json:"secure_data"`
		// Nonce the same nonce given in the request
		Nonce string `json:"nonce"`
	}

	// SecureData is a map of the fields and their encrypted values.
	SecureData map[string]*SecureValue
	// PersonalDetails       *SecureValue `json:"personal_details"`
	// Passport              *SecureValue `json:"passport"`
	// InternalPassport      *SecureValue `json:"internal_passport"`
	// DriverLicense         *SecureValue `json:"driver_license"`
	// IdentityCard          *SecureValue `json:"identity_card"`
	// Address               *SecureValue `json:"address"`
	// UtilityBill           *SecureValue `json:"utility_bill"`
	// BankStatement         *SecureValue `json:"bank_statement"`
	// RentalAgreement       *SecureValue `json:"rental_agreement"`
	// PassportRegistration  *SecureValue `json:"passport_registration"`
	// TemporaryRegistration *SecureValue `json:"temporary_registration"`

	// SecureValue contains encrypted values for a SecureData item.
	SecureValue struct {
		Data        *DataCredentials   `json:"data"`
		FrontSide   *FileCredentials   `json:"front_side"`
		ReverseSide *FileCredentials   `json:"reverse_side"`
		Selfie      *FileCredentials   `json:"selfie"`
		Translation []*FileCredentials `json:"translation"`
		Files       []*FileCredentials `json:"files"`
	}

	// DataCredentials contains information required to decrypt data.
	DataCredentials struct {
		// DataHash checksum of encrypted data
		DataHash string `json:"data_hash"`
		// Secret of encrypted data
		Secret string `json:"secret"`
	}

	// FileCredentials contains information required to decrypt files.
	FileCredentials struct {
		// FileHash checksum of encrypted data
		FileHash string `json:"file_hash"`
		// Secret of encrypted data
		Secret string `json:"secret"`
	}

	// PersonalDetails https://core.telegram.org/passport#personaldetails
	PersonalDetails struct {
		FirstName            string `json:"first_name"`
		LastName             string `json:"last_name"`
		MiddleName           string `json:"middle_name"`
		BirthDate            string `json:"birth_date"`
		Gender               string `json:"gender"`
		CountryCode          string `json:"country_code"`
		ResidenceCountryCode string `json:"residence_country_code"`
		FirstNameNative      string `json:"first_name_native"`
		LastNameNative       string `json:"last_name_native"`
		MiddleNameNative     string `json:"middle_name_native"`
	}

	// IDDocumentData https://core.telegram.org/passport#iddocumentdata
	IDDocumentData struct {
		DocumentNumber string `json:"document_no"`
		ExpiryDate     string `json:"expiry_date"`
	}
)

// Telegram constants
const (
	// APIEndpoint is the endpoint for all API methods,
	// with formatting for Sprintf.
	APIEndpoint = "https://api.telegram.org/bot%s/%s"
	// FileEndpoint is the endpoint for downloading a file from Telegram.
	FileEndpoint = "https://api.telegram.org/file/bot%s/%s"
)

// Constant values for ChatActions
const (
	ChatTyping          = "typing"
	ChatUploadPhoto     = "upload_photo"
	ChatRecordVideo     = "record_video"
	ChatUploadVideo     = "upload_video"
	ChatRecordVoice     = "record_voice"
	ChatUploadVoice     = "upload_voice"
	ChatUploadDocument  = "upload_document"
	ChatChooseSticker   = "choose_sticker"
	ChatFindLocation    = "find_location"
	ChatRecordVideoNote = "record_video_note"
	ChatUploadVideoNote = "upload_video_note"
)

// API errors
const (
	// ErrAPIForbidden happens when a token is bad
	ErrAPIForbidden = "forbidden"
)

// Constant values for ParseMode in MessageConfig
const (
	ModeMarkdown   = "Markdown"
	ModeMarkdownV2 = "MarkdownV2"
	ModeHTML       = "HTML"
)

// Constant values for update types
const (
	// UpdateTypeMessage is new incoming message of any kind — text, photo, sticker, etc.
	UpdateTypeMessage = "message"

	// UpdateTypeEditedMessage is new version of a message that is known to the bot and was edited
	UpdateTypeEditedMessage = "edited_message"

	// UpdateTypeChannelPost is new incoming channel post of any kind — text, photo, sticker, etc.
	UpdateTypeChannelPost = "channel_post"

	// UpdateTypeEditedChannelPost is new version of a channel post that is known to the bot and was edited
	UpdateTypeEditedChannelPost = "edited_channel_post"

	// UpdateTypeInlineQuery is new incoming inline query
	UpdateTypeInlineQuery = "inline_query"

	// UpdateTypeChosenInlineResult i the result of an inline query that was chosen by a user and sent to their
	// chat partner. Please see the documentation on the feedback collecting for
	// details on how to enable these updates for your bot.
	UpdateTypeChosenInlineResult = "chosen_inline_result"

	// UpdateTypeCallbackQuery is new incoming callback query
	UpdateTypeCallbackQuery = "callback_query"

	// UpdateTypeShippingQuery is new incoming shipping query. Only for invoices with flexible price
	UpdateTypeShippingQuery = "shipping_query"

	// UpdateTypePreCheckoutQuery is new incoming pre-checkout query. Contains full information about checkout
	UpdateTypePreCheckoutQuery = "pre_checkout_query"

	// UpdateTypePoll is new poll state. Bots receive only updates about stopped polls and polls
	// which are sent by the bot
	UpdateTypePoll = "poll"

	// UpdateTypePollAnswer is when user changed their answer in a non-anonymous poll. Bots receive new votes
	// only in polls that were sent by the bot itself.
	UpdateTypePollAnswer = "poll_answer"

	// UpdateTypeMyChatMember is when the bot's chat member status was updated in a chat. For private chats, this
	// update is received only when the bot is blocked or unblocked by the user.
	UpdateTypeMyChatMember = "my_chat_member"

	// UpdateTypeChatMember is when the bot must be an administrator in the chat and must explicitly specify
	// this update in the list of allowed_updates to receive these updates.
	UpdateTypeChatMember = "chat_member"
)

// Library errors
const (
	ErrBadURL = "bad or empty url"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	params() (Params, error)
	Method() string
}

// Fileable is any config type that can be sent that includes a file.
type Fileable interface {
	Chattable
	files() []RequestFile
}

// RequestFile represents a file associated with a field name.
type RequestFile struct {
	// The file field name.
	Name string `json:"name"`
	// The file data to include.
	Data RequestFileData `json:"data"`
}

// RequestFileData represents the data to be used for a file.
type RequestFileData interface {
	// NeedsUpload shows if the file needs to be uploaded.
	NeedsUpload() bool

	// UploadData gets the file name and an `io.Reader` for the file to be uploaded. This
	// must only be called when the file needs to be uploaded.
	UploadData() (string, io.Reader, error)
	// SendData gets the file data to send when a file does not need to be uploaded. This
	// must only be called when the file does not need to be uploaded.
	SendData() string
}

// FileBytes contains information about a set of bytes to upload
// as a File.
type FileBytes struct {
	Name  string `json:"name"`
	Bytes []byte `json:"bytes"`
}

func (fb FileBytes) NeedsUpload() bool {
	return true
}

func (fb FileBytes) UploadData() (string, io.Reader, error) {
	return fb.Name, bytes.NewReader(fb.Bytes), nil
}

func (fb FileBytes) SendData() string {
	panic("FileBytes must be uploaded")
}

// FileReader contains information about a reader to upload as a File.
type FileReader struct {
	Name   string    `json:"name"`
	Reader io.Reader `json:"reader"`
}

func (fr FileReader) NeedsUpload() bool {
	return true
}

func (fr FileReader) UploadData() (string, io.Reader, error) {
	return fr.Name, fr.Reader, nil
}

func (fr FileReader) SendData() string {
	panic("FileReader must be uploaded")
}

// FilePath is a path to a local file.
type FilePath string

func (fp FilePath) NeedsUpload() bool {
	return true
}

func (fp FilePath) UploadData() (string, io.Reader, error) {
	fileHandle, err := os.Open(string(fp))
	if err != nil {
		return "", nil, err
	}

	name := fileHandle.Name()
	return name, fileHandle, err
}

func (fp FilePath) SendData() string {
	panic("FilePath must be uploaded")
}

// FileURL is a URL to use as a file for a request.
type FileURL string

func (fu FileURL) NeedsUpload() bool {
	return false
}

func (fu FileURL) UploadData() (string, io.Reader, error) {
	panic("FileURL cannot be uploaded")
}

func (fu FileURL) SendData() string {
	return string(fu)
}

// FileID is an ID of a file already uploaded to Telegram.
type FileID string

func (fi FileID) NeedsUpload() bool {
	return false
}

func (fi FileID) UploadData() (string, io.Reader, error) {
	panic("FileID cannot be uploaded")
}

func (fi FileID) SendData() string {
	return string(fi)
}

// fileAttach is an internal file type used for processed media groups.
type fileAttach string

func (fa fileAttach) NeedsUpload() bool {
	return false
}

func (fa fileAttach) UploadData() (string, io.Reader, error) {
	panic("fileAttach cannot be uploaded")
}

func (fa fileAttach) SendData() string {
	return string(fa)
}

// LogOutConfig is a request to log out of the cloud Bot API server.
//
// Note that you may not log back in for at least 10 minutes.
type LogOutConfig struct{}

func (LogOutConfig) Method() string {
	return "logOut"
}

func (LogOutConfig) params() (Params, error) {
	return nil, nil
}

// CloseConfig is a request to close the bot instance on a local server.
//
// Note that you may not close an instance for the first 10 minutes after the
// bot has started.
type CloseConfig struct{}

func (CloseConfig) Method() string {
	return "close"
}

func (CloseConfig) params() (Params, error) {
	return nil, nil
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	// user_id(int64) or username
	ChatID                   int64       `json:"chat_id"`                       // required
	ChannelUsername          string      `json:"channel_username,omitempty"`    // optional
	ProtectContent           bool        `json:"protect_content,omitempty"`     // optional
	ReplyToMessageID         int         `json:"reply_to_message_id,omitempty"` // optional
	ReplyMarkup              interface{} `json:"reply_markup,omitempty"`        // optional
	DisableNotification      bool        `json:"disable_notification,omitempty"`
	AllowSendingWithoutReply bool        `json:"allow_sending_without_reply,omitempty"`
}

// BaseFile is a base type for all file config types.
type BaseFile struct {
	BaseChat
	File RequestFileData `json:"file"`
}

// BaseEdit is base type of all chat edits.
type BaseEdit struct {
	ChatID          int64                 `json:"chat_id"`
	ChannelUsername string                `json:"channel_username,omitempty"`
	MessageID       int                   `json:"message_id"`
	InlineMessageID string                `json:"inline_message_id,omitempty"`
	ReplyMarkup     *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	Text                  string          `json:"text"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	Entities              []MessageEntity `json:"entities,omitempty"`
	DisableWebPagePreview bool            `json:"disable_web_page_preview,omitempty"`
}

func (config MessageConfig) Method() string {
	return "sendMessage"
}

// ForwardConfig contains information about a ForwardMessage request.
type ForwardConfig struct {
	BaseChat
	FromChatID          int64  `json:"from_chat_id"` // required
	FromChannelUsername string `json:"from_channel_username,omitempty"`
	MessageID           int    `json:"message_id"` // required
}

func (config ForwardConfig) Method() string {
	return "forwardMessage"
}

// CopyMessageConfig contains information about a copyMessage request.
type CopyMessageConfig struct {
	BaseChat
	FromChatID          int64           `json:"from_chat_id"`
	FromChannelUsername string          `json:"from_channel_username,omitempty"`
	MessageID           int             `json:"message_id"`
	Caption             string          `json:"caption"`
	ParseMode           string          `json:"parse_mode,omitempty"`
	CaptionEntities     []MessageEntity `json:"caption_entities,omitempty"`
}

func (config CopyMessageConfig) Method() string {
	return "copyMessage"
}

// PhotoConfig contains information about a SendPhoto request.
type PhotoConfig struct {
	BaseFile
	Thumb           RequestFileData `json:"thumb,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
}

func (config PhotoConfig) Method() string {
	return "sendPhoto"
}

func (config PhotoConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "photo",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// AudioConfig contains information about a SendAudio request.
type AudioConfig struct {
	BaseFile
	Thumb           RequestFileData `json:"thumb,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Duration        int             `json:"duration,omitempty"`
	Performer       string          `json:"performer,omitempty"`
	Title           string          `json:"title,omitempty"`
}

func (config AudioConfig) Method() string {
	return "sendAudio"
}

func (config AudioConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "audio",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// DocumentConfig contains information about a SendDocument request.
type DocumentConfig struct {
	BaseFile
	Thumb                       RequestFileData `json:"thumb,omitempty"`
	Caption                     string          `json:"caption,omitempty"`
	ParseMode                   string          `json:"parse_mode,omitempty"`
	CaptionEntities             []MessageEntity `json:"caption_entities,omitempty"`
	DisableContentTypeDetection bool            `json:"disable_content_type_detection,omitempty"`
}

func (config DocumentConfig) Method() string {
	return "sendDocument"
}

func (config DocumentConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "document",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// StickerConfig contains information about a SendSticker request.
type StickerConfig struct {
	BaseFile
}

func (config StickerConfig) Method() string {
	return "sendSticker"
}

func (config StickerConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "sticker",
		Data: config.File,
	}}
}

// VideoConfig contains information about a SendVideo request.
type VideoConfig struct {
	BaseFile
	Thumb             RequestFileData `json:"thumb,omitempty"`
	Duration          int             `json:"duration,omitempty"`
	Caption           string          `json:"caption,omitempty"`
	ParseMode         string          `json:"parse_mode,omitempty"`
	CaptionEntities   []MessageEntity `json:"caption_entities,omitempty"`
	SupportsStreaming bool            `json:"supports_streaming,omitempty"`
}

func (config VideoConfig) Method() string {
	return "sendVideo"
}

func (config VideoConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "video",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// AnimationConfig contains information about a SendAnimation request.
type AnimationConfig struct {
	BaseFile
	Duration        int             `json:"duration,omitempty"`
	Thumb           RequestFileData `json:"thumb,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
}

func (config AnimationConfig) Method() string {
	return "sendAnimation"
}

func (config AnimationConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "animation",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// VideoNoteConfig contains information about a SendVideoNote request.
type VideoNoteConfig struct {
	BaseFile
	Thumb    RequestFileData `json:"thumb,omitempty"`
	Duration int             `json:"duration,omitempty"`
	Length   int             `json:"length,omitempty"`
}

func (config VideoNoteConfig) Method() string {
	return "sendVideoNote"
}

func (config VideoNoteConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "video_note",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// VoiceConfig contains information about a SendVoice request.
type VoiceConfig struct {
	BaseFile
	Thumb           RequestFileData `json:"thumb,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Duration        int             `json:"duration,omitempty"`
}

func (config VoiceConfig) Method() string {
	return "sendVoice"
}

func (config VoiceConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "voice",
		Data: config.File,
	}}

	if config.Thumb != nil {
		files = append(files, RequestFile{
			Name: "thumb",
			Data: config.Thumb,
		})
	}

	return files
}

// LocationConfig contains information about a SendLocation request.
type LocationConfig struct {
	BaseChat
	Latitude             float64 `json:"latitude"`                         // required
	Longitude            float64 `json:"longitude"`                        // required
	HorizontalAccuracy   float64 `json:"horizontal_accuracy,omitempty"`    // optional
	LivePeriod           int     `json:"live_period,omitempty"`            // optional
	Heading              int     `json:"heading,omitempty"`                // optional
	ProximityAlertRadius int     `json:"proximity_alert_radius,omitempty"` // optional
}

func (config LocationConfig) Method() string {
	return "sendLocation"
}

// EditMessageLiveLocationConfig allows you to update a live location.
type EditMessageLiveLocationConfig struct {
	BaseEdit
	Latitude             float64 `json:"latitude"`                         // required
	Longitude            float64 `json:"longitude"`                        // required
	HorizontalAccuracy   float64 `json:"horizontal_accuracy,omitempty"`    // optional
	Heading              int     `json:"heading,omitempty"`                // optional
	ProximityAlertRadius int     `json:"proximity_alert_radius,omitempty"` // optional
}

func (config EditMessageLiveLocationConfig) Method() string {
	return "editMessageLiveLocation"
}

// StopMessageLiveLocationConfig stops updating a live location.
type StopMessageLiveLocationConfig struct {
	BaseEdit
}

func (config StopMessageLiveLocationConfig) Method() string {
	return "stopMessageLiveLocation"
}

// VenueConfig contains information about a SendVenue request.
type VenueConfig struct {
	BaseChat
	Latitude        float64 `json:"latitude"`  // required
	Longitude       float64 `json:"longitude"` // required
	Title           string  `json:"title"`     // required
	Address         string  `json:"address"`   // required
	FoursquareID    string  `json:"foursquare_id,omitempty"`
	FoursquareType  string  `json:"foursquare_type,omitempty"`
	GooglePlaceID   string  `json:"google_place_id,omitempty"`
	GooglePlaceType string  `json:"google_place_type,omitempty"`
}

func (config VenueConfig) Method() string {
	return "sendVenue"
}

// ContactConfig allows you to send a contact.
type ContactConfig struct {
	BaseChat
	PhoneNumber string
	FirstName   string
	LastName    string
	VCard       string
}

func (config ContactConfig) Method() string {
	return "sendContact"
}

// SendPollConfig allows you to send a poll.
type SendPollConfig struct {
	BaseChat
	Question              string          `json:"question"`     // required
	Options               []string        `json:"options"`      // required
	IsAnonymous           bool            `json:"is_anonymous"` // required
	Type                  string          `json:"type"`         // required
	AllowsMultipleAnswers bool            `json:"allows_multiple_answers"`
	CorrectOptionID       int64           `json:"correct_option_id,omitempty"`
	Explanation           string          `json:"explanation,omitempty"`
	ExplanationParseMode  string          `json:"explanation_parse_mode,omitempty"`
	ExplanationEntities   []MessageEntity `json:"explanation_entities,omitempty"`
	OpenPeriod            int             `json:"open_period,omitempty"`
	CloseDate             int             `json:"close_date,omitempty"`
	IsClosed              bool            `json:"is_closed,omitempty"`
}

func (SendPollConfig) Method() string {
	return "sendPoll"
}

// GameConfig allows you to send a game.
type GameConfig struct {
	BaseChat
	GameShortName string
}

func (config GameConfig) Method() string {
	return "sendGame"
}

// SetGameScoreConfig allows you to update the game score in a chat.
type SetGameScoreConfig struct {
	UserID             int64  `json:"user_id,omitempty"`              // required
	Score              int    `json:"score,omitempty"`                // required
	Force              bool   `json:"force,omitempty"`                // optional
	DisableEditMessage bool   `json:"disable_edit_message,omitempty"` // optional
	ChatID             int64  `json:"chat_id,omitempty"`
	ChannelUsername    string `json:"channel_username,omitempty"`
	MessageID          int    `json:"message_id,omitempty"`
	InlineMessageID    string `json:"inline_message_id,omitempty"`
}

func (config SetGameScoreConfig) Method() string {
	return "setGameScore"
}

// GetGameHighScoresConfig allows you to fetch the high scores for a game.
type GetGameHighScoresConfig struct {
	UserID          int64  `json:"user_id,omitempty"`
	ChatID          int64  `json:"chat_id,omitempty"`
	ChannelUsername string `json:"channel_username,omitempty"`
	MessageID       int    `json:"message_id,omitempty"`
	InlineMessageID string `json:"inline_message_id,omitempty"`
}

func (config GetGameHighScoresConfig) Method() string {
	return "getGameHighScores"
}

// ChatActionConfig contains information about a SendChatAction request.
type ChatActionConfig struct {
	BaseChat
	Action string // required
}

func (config ChatActionConfig) Method() string {
	return "sendChatAction"
}

// EditMessageTextConfig allows you to modify the text in a message.
type EditMessageTextConfig struct {
	BaseEdit
	Text                  string          `json:"text"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	Entities              []MessageEntity `json:"entities,omitempty"`
	DisableWebPagePreview bool            `json:"disable_web_page_preview,omitempty"`
}

func (config EditMessageTextConfig) Method() string {
	return "editMessageText"
}

// EditMessageCaptionConfig allows you to modify the caption of a message.
type EditMessageCaptionConfig struct {
	BaseEdit
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
}

func (config EditMessageCaptionConfig) Method() string {
	return "editMessageCaption"
}

// EditMessageMediaConfig allows you to make an editMessageMedia request.
type EditMessageMediaConfig struct {
	BaseEdit

	Media interface{} `json:"media"`
}

func (EditMessageMediaConfig) Method() string {
	return "editMessageMedia"
}

func (config EditMessageMediaConfig) files() []RequestFile {
	return prepareInputMediaFile(config.Media, 0)
}

// EditMessageReplyMarkupConfig allows you to modify the reply markup
// of a message.
type EditMessageReplyMarkupConfig struct {
	BaseEdit
}

func (config EditMessageReplyMarkupConfig) Method() string {
	return "editMessageReplyMarkup"
}

// StopPollConfig allows you to stop a poll sent by the bot.
type StopPollConfig struct {
	BaseEdit
}

func (StopPollConfig) Method() string {
	return "stopPoll"
}

// UserProfilePhotosConfig contains information about a
// GetUserProfilePhotos request.
type UserProfilePhotosConfig struct {
	UserID int64 `json:"user_id"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

func (UserProfilePhotosConfig) Method() string {
	return "getUserProfilePhotos"
}

// FileConfig has information about a file hosted on Telegram.
type FileConfig struct {
	FileID string `json:"file_id"`
}

func (FileConfig) Method() string {
	return "getFile"
}

func (config FileConfig) params() (Params, error) {
	params := make(Params)

	params["file_id"] = config.FileID

	return params, nil
}

// UpdateConfig contains information about a GetUpdates request.
type UpdateConfig struct {
	Offset         int      `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

func (UpdateConfig) Method() string {
	return "getUpdates"
}

// WebhookConfig contains information about a SetWebhook request.
type WebhookConfig struct {
	URL                *url.URL        `json:"url"`
	Certificate        RequestFileData `json:"certificate,omitempty"`
	IPAddress          string          `json:"ip_address,omitempty"`
	MaxConnections     int             `json:"max_connections,omitempty"`
	AllowedUpdates     []string        `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool            `json:"drop_pending_updates,omitempty"`
}

func (config WebhookConfig) Method() string {
	return "setWebhook"
}

func (config WebhookConfig) files() []RequestFile {
	if config.Certificate != nil {
		return []RequestFile{{
			Name: "certificate",
			Data: config.Certificate,
		}}
	}

	return nil
}

// DeleteWebhookConfig is a helper to delete a webhook.
type DeleteWebhookConfig struct {
	DropPendingUpdates bool `json:"drop_pending_updates,omitempty"`
}

func (config DeleteWebhookConfig) Method() string {
	return "deleteWebhook"
}

// InlineConfig contains information on making an InlineQuery response.
type InlineConfig struct {
	InlineQueryID     string        `json:"inline_query_id"`
	Results           []interface{} `json:"results"`
	CacheTime         int           `json:"cache_time"`
	IsPersonal        bool          `json:"is_personal"`
	NextOffset        string        `json:"next_offset"`
	SwitchPMText      string        `json:"switch_pm_text"`
	SwitchPMParameter string        `json:"switch_pm_parameter"`
}

func (config InlineConfig) Method() string {
	return "answerInlineQuery"
}

// AnswerWebAppQueryConfig is used to set the result of an interaction with a
// Web App and send a corresponding message on behalf of the user to the chat
// from which the query originated.
type AnswerWebAppQueryConfig struct {
	// WebAppQueryID is the unique identifier for the query to be answered.
	WebAppQueryID string `json:"web_app_query_id"`
	// Result is an InlineQueryResult object describing the message to be sent.
	Result interface{} `json:"result"`
}

func (config AnswerWebAppQueryConfig) Method() string {
	return "answerWebAppQuery"
}

// CallbackConfig contains information on making a CallbackQuery response.
type CallbackConfig struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text"`
	ShowAlert       bool   `json:"show_alert"`
	URL             string `json:"url"`
	CacheTime       int    `json:"cache_time"`
}

func (config CallbackConfig) Method() string {
	return "answerCallbackQuery"
}

// ChatMemberConfig contains information about a user in a chat for use
// with administrative functions such as kicking or unbanning a user.
type ChatMemberConfig struct {
	ChatID             int64
	SuperGroupUsername string
	ChannelUsername    string
	UserID             int64
}

// UnbanChatMemberConfig allows you to unban a user.
type UnbanChatMemberConfig struct {
	ChatMemberConfig
	OnlyIfBanned bool
}

func (config UnbanChatMemberConfig) Method() string {
	return "unbanChatMember"
}

// BanChatMemberConfig contains extra fields to kick user.
type BanChatMemberConfig struct {
	ChatMemberConfig
	UntilDate      int64
	RevokeMessages bool
}

func (config BanChatMemberConfig) Method() string {
	return "banChatMember"
}

// KickChatMemberConfig contains extra fields to ban user.
//
// This was renamed to BanChatMember in later versions of the Telegram Bot API.
type KickChatMemberConfig = BanChatMemberConfig

// RestrictChatMemberConfig contains fields to restrict members of chat
type RestrictChatMemberConfig struct {
	ChatMemberConfig
	UntilDate   int64
	Permissions *ChatPermissions
}

func (config RestrictChatMemberConfig) Method() string {
	return "restrictChatMember"
}

// PromoteChatMemberConfig contains fields to promote members of chat
type PromoteChatMemberConfig struct {
	ChatMemberConfig
	IsAnonymous         bool
	CanManageChat       bool
	CanChangeInfo       bool
	CanPostMessages     bool
	CanEditMessages     bool
	CanDeleteMessages   bool
	CanManageVideoChats bool
	CanInviteUsers      bool
	CanRestrictMembers  bool
	CanPinMessages      bool
	CanPromoteMembers   bool
}

func (config PromoteChatMemberConfig) Method() string {
	return "promoteChatMember"
}

// SetChatAdministratorCustomTitle sets the title of an administrative user
// promoted by the bot for a chat.
type SetChatAdministratorCustomTitle struct {
	ChatMemberConfig
	CustomTitle string
}

func (SetChatAdministratorCustomTitle) Method() string {
	return "setChatAdministratorCustomTitle"
}

// BanChatSenderChatConfig bans a channel chat in a supergroup or a channel. The
// owner of the chat will not be able to send messages and join live streams on
// behalf of the chat, unless it is unbanned first. The bot must be an
// administrator in the supergroup or channel for this to work and must have the
// appropriate administrator rights.
type BanChatSenderChatConfig struct {
	ChatID          int64
	ChannelUsername string
	SenderChatID    int64
	UntilDate       int
}

func (config BanChatSenderChatConfig) Method() string {
	return "banChatSenderChat"
}

// UnbanChatSenderChatConfig unbans a previously banned channel chat in a
// supergroup or channel. The bot must be an administrator for this to work and
// must have the appropriate administrator rights.
type UnbanChatSenderChatConfig struct {
	ChatID          int64
	ChannelUsername string
	SenderChatID    int64
}

func (config UnbanChatSenderChatConfig) Method() string {
	return "unbanChatSenderChat"
}

// ChatConfig contains information about getting information on a chat.
type ChatConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

// ChatInfoConfig contains information about getting chat information.
type ChatInfoConfig struct {
	ChatConfig
}

func (ChatInfoConfig) Method() string {
	return "getChat"
}

// ChatMemberCountConfig contains information about getting the number of users in a chat.
type ChatMemberCountConfig struct {
	ChatConfig
}

func (ChatMemberCountConfig) Method() string {
	return "getChatMembersCount"
}

// ChatAdministratorsConfig contains information about getting chat administrators.
type ChatAdministratorsConfig struct {
	ChatConfig
}

func (ChatAdministratorsConfig) Method() string {
	return "getChatAdministrators"
}

// SetChatPermissionsConfig allows you to set default permissions for the
// members in a group. The bot must be an administrator and have rights to
// restrict members.
type SetChatPermissionsConfig struct {
	ChatConfig
	Permissions *ChatPermissions
}

func (SetChatPermissionsConfig) Method() string {
	return "setChatPermissions"
}

// ChatInviteLinkConfig contains information about getting a chat link.
//
// Note that generating a new link will revoke any previous links.
type ChatInviteLinkConfig struct {
	ChatConfig
}

func (ChatInviteLinkConfig) Method() string {
	return "exportChatInviteLink"
}

// CreateChatInviteLinkConfig allows you to create an additional invite link for
// a chat. The bot must be an administrator in the chat for this to work and
// must have the appropriate admin rights. The link can be revoked using the
// RevokeChatInviteLinkConfig.
type CreateChatInviteLinkConfig struct {
	ChatConfig
	Name               string
	ExpireDate         int
	MemberLimit        int
	CreatesJoinRequest bool
}

func (CreateChatInviteLinkConfig) Method() string {
	return "createChatInviteLink"
}

// EditChatInviteLinkConfig allows you to edit a non-primary invite link created
// by the bot. The bot must be an administrator in the chat for this to work and
// must have the appropriate admin rights.
type EditChatInviteLinkConfig struct {
	ChatConfig
	InviteLink         string
	Name               string
	ExpireDate         int
	MemberLimit        int
	CreatesJoinRequest bool
}

func (EditChatInviteLinkConfig) Method() string {
	return "editChatInviteLink"
}

// RevokeChatInviteLinkConfig allows you to revoke an invite link created by the
// bot. If the primary link is revoked, a new link is automatically generated.
// The bot must be an administrator in the chat for this to work and must have
// the appropriate admin rights.
type RevokeChatInviteLinkConfig struct {
	ChatConfig
	InviteLink string
}

func (RevokeChatInviteLinkConfig) Method() string {
	return "revokeChatInviteLink"
}

// ApproveChatJoinRequestConfig allows you to approve a chat join request.
type ApproveChatJoinRequestConfig struct {
	ChatConfig
	UserID int64
}

func (ApproveChatJoinRequestConfig) Method() string {
	return "approveChatJoinRequest"
}

// DeclineChatJoinRequest allows you to decline a chat join request.
type DeclineChatJoinRequest struct {
	ChatConfig
	UserID int64
}

func (DeclineChatJoinRequest) Method() string {
	return "declineChatJoinRequest"
}

// LeaveChatConfig allows you to leave a chat.
type LeaveChatConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config LeaveChatConfig) Method() string {
	return "leaveChat"
}

// ChatConfigWithUser contains information about a chat and a user.
type ChatConfigWithUser struct {
	ChatID             int64
	SuperGroupUsername string
	UserID             int64
}

// GetChatMemberConfig is information about getting a specific member in a chat.
type GetChatMemberConfig struct {
	ChatConfigWithUser
}

func (GetChatMemberConfig) Method() string {
	return "getChatMember"
}

// InvoiceConfig contains information for sendInvoice request.
type InvoiceConfig struct {
	BaseChat
	Title                     string         // required
	Description               string         // required
	Payload                   string         // required
	ProviderToken             string         // required
	Currency                  string         // required
	Prices                    []LabeledPrice // required
	MaxTipAmount              int
	SuggestedTipAmounts       []int
	StartParameter            string
	ProviderData              string
	PhotoURL                  string
	PhotoSize                 int
	PhotoWidth                int
	PhotoHeight               int
	NeedName                  bool
	NeedPhoneNumber           bool
	NeedEmail                 bool
	NeedShippingAddress       bool
	SendPhoneNumberToProvider bool
	SendEmailToProvider       bool
	IsFlexible                bool
}

func (config InvoiceConfig) Method() string {
	return "sendInvoice"
}

// ShippingConfig contains information for answerShippingQuery request.
type ShippingConfig struct {
	ShippingQueryID string // required
	OK              bool   // required
	ShippingOptions []ShippingOption
	ErrorMessage    string
}

func (config ShippingConfig) Method() string {
	return "answerShippingQuery"
}

// PreCheckoutConfig contains information for answerPreCheckoutQuery request.
type PreCheckoutConfig struct {
	PreCheckoutQueryID string // required
	OK                 bool   // required
	ErrorMessage       string
}

func (config PreCheckoutConfig) Method() string {
	return "answerPreCheckoutQuery"
}

// DeleteMessageConfig contains information of a message in a chat to delete.
type DeleteMessageConfig struct {
	ChannelUsername string
	ChatID          int64
	MessageID       int
}

func (config DeleteMessageConfig) Method() string {
	return "deleteMessage"
}

// PinChatMessageConfig contains information of a message in a chat to pin.
type PinChatMessageConfig struct {
	ChatID              int64
	ChannelUsername     string
	MessageID           int
	DisableNotification bool
}

func (config PinChatMessageConfig) Method() string {
	return "pinChatMessage"
}

// UnpinChatMessageConfig contains information of a chat message to unpin.
//
// If MessageID is not specified, it will unpin the most recent pin.
type UnpinChatMessageConfig struct {
	ChatID          int64
	ChannelUsername string
	MessageID       int
}

func (config UnpinChatMessageConfig) Method() string {
	return "unpinChatMessage"
}

// UnpinAllChatMessagesConfig contains information of all messages to unpin in
// a chat.
type UnpinAllChatMessagesConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config UnpinAllChatMessagesConfig) Method() string {
	return "unpinAllChatMessages"
}

// SetChatPhotoConfig allows you to set a group, supergroup, or channel's photo.
type SetChatPhotoConfig struct {
	BaseFile
}

func (config SetChatPhotoConfig) Method() string {
	return "setChatPhoto"
}

func (config SetChatPhotoConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "photo",
		Data: config.File,
	}}
}

// DeleteChatPhotoConfig allows you to delete a group, supergroup, or channel's photo.
type DeleteChatPhotoConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config DeleteChatPhotoConfig) Method() string {
	return "deleteChatPhoto"
}

// SetChatTitleConfig allows you to set the title of something other than a private chat.
type SetChatTitleConfig struct {
	ChatID          int64
	ChannelUsername string

	Title string
}

func (config SetChatTitleConfig) Method() string {
	return "setChatTitle"
}

// SetChatDescriptionConfig allows you to set the description of a supergroup or channel.
type SetChatDescriptionConfig struct {
	ChatID          int64
	ChannelUsername string

	Description string
}

func (config SetChatDescriptionConfig) Method() string {
	return "setChatDescription"
}

// GetStickerSetConfig allows you to get the stickers in a set.
type GetStickerSetConfig struct {
	Name string
}

func (config GetStickerSetConfig) Method() string {
	return "getStickerSet"
}

func (config GetStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name

	return params, nil
}

// UploadStickerConfig allows you to upload a sticker for use in a set later.
type UploadStickerConfig struct {
	UserID     int64
	PNGSticker RequestFileData
}

func (config UploadStickerConfig) Method() string {
	return "uploadStickerFile"
}

func (config UploadStickerConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "png_sticker",
		Data: config.PNGSticker,
	}}
}

// NewStickerSetConfig allows creating a new sticker set.
//
// You must set either PNGSticker or TGSSticker.
type NewStickerSetConfig struct {
	UserID        int64
	Name          string
	Title         string
	PNGSticker    RequestFileData
	TGSSticker    RequestFileData
	Emojis        string
	ContainsMasks bool
	MaskPosition  *MaskPosition
}

func (config NewStickerSetConfig) Method() string {
	return "createNewStickerSet"
}

func (config NewStickerSetConfig) files() []RequestFile {
	if config.PNGSticker != nil {
		return []RequestFile{{
			Name: "png_sticker",
			Data: config.PNGSticker,
		}}
	}

	return []RequestFile{{
		Name: "tgs_sticker",
		Data: config.TGSSticker,
	}}
}

// AddStickerConfig allows you to add a sticker to a set.
type AddStickerConfig struct {
	UserID       int64
	Name         string
	PNGSticker   RequestFileData
	TGSSticker   RequestFileData
	Emojis       string
	MaskPosition *MaskPosition
}

func (config AddStickerConfig) Method() string {
	return "addStickerToSet"
}

func (config AddStickerConfig) files() []RequestFile {
	if config.PNGSticker != nil {
		return []RequestFile{{
			Name: "png_sticker",
			Data: config.PNGSticker,
		}}
	}

	return []RequestFile{{
		Name: "tgs_sticker",
		Data: config.TGSSticker,
	}}

}

// SetStickerPositionConfig allows you to change the position of a sticker in a set.
type SetStickerPositionConfig struct {
	Sticker  string
	Position int
}

func (config SetStickerPositionConfig) Method() string {
	return "setStickerPositionInSet"
}

// DeleteStickerConfig allows you to delete a sticker from a set.
type DeleteStickerConfig struct {
	Sticker string
}

func (config DeleteStickerConfig) Method() string {
	return "deleteStickerFromSet"
}

func (config DeleteStickerConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker

	return params, nil
}

// SetStickerSetThumbConfig allows you to set the thumbnail for a sticker set.
type SetStickerSetThumbConfig struct {
	Name   string
	UserID int64
	Thumb  RequestFileData
}

func (config SetStickerSetThumbConfig) Method() string {
	return "setStickerSetThumb"
}

func (config SetStickerSetThumbConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "thumb",
		Data: config.Thumb,
	}}
}

// SetChatStickerSetConfig allows you to set the sticker set for a supergroup.
type SetChatStickerSetConfig struct {
	ChatID             int64
	SuperGroupUsername string

	StickerSetName string
}

func (config SetChatStickerSetConfig) Method() string {
	return "setChatStickerSet"
}

// DeleteChatStickerSetConfig allows you to remove a supergroup's sticker set.
type DeleteChatStickerSetConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config DeleteChatStickerSetConfig) Method() string {
	return "deleteChatStickerSet"
}

// MediaGroupConfig allows you to send a group of media.
//
// Media consist of InputMedia items (InputMediaPhoto, InputMediaVideo).
type MediaGroupConfig struct {
	ChatID          int64
	ChannelUsername string

	Media               []interface{}
	DisableNotification bool
	ReplyToMessageID    int
}

func (config MediaGroupConfig) Method() string {
	return "sendMediaGroup"
}

func (config MediaGroupConfig) files() []RequestFile {
	return prepareInputMediaForFiles(config.Media)
}

// DiceConfig contains information about a sendDice request.
type DiceConfig struct {
	BaseChat
	// Emoji on which the dice throw animation is based.
	// Currently, must be one of 🎲, 🎯, 🏀, ⚽, 🎳, or 🎰.
	// Dice can have values 1-6 for 🎲, 🎯, and 🎳, values 1-5 for 🏀 and ⚽,
	// and values 1-64 for 🎰.
	// Defaults to "🎲"
	Emoji string `json:"emoji,omitempty"`
}

func (config DiceConfig) Method() string {
	return "sendDice"
}

// GetMyCommandsConfig gets a list of the currently registered commands.
type GetMyCommandsConfig struct {
	Scope        *BotCommandScope
	LanguageCode string
}

func (config GetMyCommandsConfig) Method() string {
	return "getMyCommands"
}

// SetMyCommandsConfig sets a list of commands the bot understands.
type SetMyCommandsConfig struct {
	Commands     []BotCommand
	Scope        *BotCommandScope
	LanguageCode string
}

func (config SetMyCommandsConfig) Method() string {
	return "setMyCommands"
}

type DeleteMyCommandsConfig struct {
	Scope        *BotCommandScope
	LanguageCode string
}

func (config DeleteMyCommandsConfig) Method() string {
	return "deleteMyCommands"
}

// SetChatMenuButtonConfig changes the bot's menu button in a private chat,
// or the default menu button.
type SetChatMenuButtonConfig struct {
	ChatID          int64
	ChannelUsername string

	MenuButton *MenuButton
}

func (config SetChatMenuButtonConfig) Method() string {
	return "setChatMenuButton"
}

type GetChatMenuButtonConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config GetChatMenuButtonConfig) Method() string {
	return "getChatMenuButton"
}

type SetMyDefaultAdministratorRightsConfig struct {
	Rights      ChatAdministratorRights
	ForChannels bool
}

func (config SetMyDefaultAdministratorRightsConfig) Method() string {
	return "setMyDefaultAdministratorRights"
}

type GetMyDefaultAdministratorRightsConfig struct {
	ForChannels bool
}

func (config GetMyDefaultAdministratorRightsConfig) Method() string {
	return "getMyDefaultAdministratorRights"
}

// prepareInputMediaParam evaluates a single InputMedia and determines if it
// needs to be modified for a successful upload. If it returns nil, then the
// value does not need to be included in the params. Otherwise, it will return
// the same type as was originally provided.
//
// The idx is used to calculate the file field name. If you only have a single
// file, 0 may be used. It is formatted into "attach://file-%d" for the primary
// media and "attach://file-%d-thumb" for thumbnails.
//
// It is expected to be used in conjunction with prepareInputMediaFile.
func prepareInputMediaParam(inputMedia interface{}, idx int) interface{} {
	switch m := inputMedia.(type) {
	case InputMediaPhoto:
		// TODO: need fix
		panic("InputMediaPhoto is not implemented,need fix")
		// if m.Media.NeedsUpload() {
		// 	m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		// }

		// return m
	case InputMediaVideo:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			m.Thumb = fileAttach(fmt.Sprintf("attach://file-%d-thumb", idx))
		}

		return m
	case InputMediaAudio:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			m.Thumb = fileAttach(fmt.Sprintf("attach://file-%d-thumb", idx))
		}

		return m
	case InputMediaDocument:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			m.Thumb = fileAttach(fmt.Sprintf("attach://file-%d-thumb", idx))
		}

		return m
	}

	return nil
}

// prepareInputMediaFile generates an array of RequestFile to provide for
// Fileable's files method. It returns an array as a single InputMedia may have
// multiple files, for the primary media and a thumbnail.
//
// The idx parameter is used to generate file field names. It uses the names
// "file-%d" for the main file and "file-%d-thumb" for the thumbnail.
//
// It is expected to be used in conjunction with prepareInputMediaParam.
func prepareInputMediaFile(inputMedia interface{}, idx int) []RequestFile {
	files := []RequestFile{}

	switch m := inputMedia.(type) {
	case InputMediaPhoto:
		// TODO: need fix
		panic("InputMediaPhoto is not implemented,need fix")
		// if m.Media.NeedsUpload() {
		// 	files = append(files, RequestFile{
		// 		Name: fmt.Sprintf("file-%d", idx),
		// 		Data: m.Media,
		// 	})
		// }
	case InputMediaVideo:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumb,
			})
		}
	case InputMediaDocument:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumb,
			})
		}
	case InputMediaAudio:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumb != nil && m.Thumb.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumb,
			})
		}
	}

	return files
}

// prepareInputMediaForParams calls prepareInputMediaParam for each item
// provided and returns a new array with the correct params for a request.
//
// It is expected that files will get data from the associated function,
// prepareInputMediaForFiles.
func prepareInputMediaForParams(inputMedia []interface{}) []interface{} {
	newMedia := make([]interface{}, len(inputMedia))
	copy(newMedia, inputMedia)

	for idx, media := range inputMedia {
		if param := prepareInputMediaParam(media, idx); param != nil {
			newMedia[idx] = param
		}
	}

	return newMedia
}

// prepareInputMediaForFiles calls prepareInputMediaFile for each item
// provided and returns a new array with the correct files for a request.
//
// It is expected that params will get data from the associated function,
// prepareInputMediaForParams.
func prepareInputMediaForFiles(inputMedia []interface{}) []RequestFile {
	files := []RequestFile{}

	for idx, media := range inputMedia {
		if file := prepareInputMediaFile(media, idx); file != nil {
			files = append(files, file...)
		}
	}

	return files
}
