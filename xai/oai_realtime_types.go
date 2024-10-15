package xai

// Event type constants
const (
	// Client event types
	OAIRealtimeEventTypeSessionUpdate            = "session.update"
	OAIRealtimeEventTypeInputAudioBufferAppend   = "input_audio_buffer.append"
	OAIRealtimeEventTypeInputAudioBufferCommit   = "input_audio_buffer.commit"
	OAIRealtimeEventTypeInputAudioBufferClear    = "input_audio_buffer.clear"
	OAIRealtimeEventTypeConversationItemCreate   = "conversation.item.create"
	OAIRealtimeEventTypeConversationItemTruncate = "conversation.item.truncate"
	OAIRealtimeEventTypeConversationItemDelete   = "conversation.item.delete"
	OAIRealtimeEventTypeResponseCreate           = "response.create"
	OAIRealtimeEventTypeResponseCancel           = "response.cancel"
	// Server event types
	OAIRealtimeEventTypeError                                            = "error"
	OAIRealtimeEventTypeSessionCreated                                   = "session.created"
	OAIRealtimeEventTypeSessionUpdated                                   = "session.updated"
	OAIRealtimeEventTypeConversationCreated                              = "conversation.created"
	OAIRealtimeEventTypeConversationItemCreated                          = "conversation.item.created"
	OAIRealtimeEventTypeConversationItemInputAudioTranscriptionCompleted = "conversation.item.input_audio_transcription.completed"
	OAIRealtimeEventTypeConversationItemInputAudioTranscriptionFailed    = "conversation.item.input_audio_transcription.failed"
	OAIRealtimeEventTypeConversationItemTruncated                        = "conversation.item.truncated"
	OAIRealtimeEventTypeConversationItemDeleted                          = "conversation.item.deleted"
	OAIRealtimeEventTypeInputAudioBufferCommitted                        = "input_audio_buffer.committed"
	OAIRealtimeEventTypeInputAudioBufferCleared                          = "input_audio_buffer.cleared"
	OAIRealtimeEventTypeInputAudioBufferSpeechStarted                    = "input_audio_buffer.speech_started"
	OAIRealtimeEventTypeInputAudioBufferSpeechStopped                    = "input_audio_buffer.speech_stopped"
	OAIRealtimeEventTypeResponseCreated                                  = "response.created"
	OAIRealtimeEventTypeResponseDone                                     = "response.done"
	OAIRealtimeEventTypeResponseOutputItemAdded                          = "response.output_item.added"
	OAIRealtimeEventTypeResponseOutputItemDone                           = "response.output_item.done"
	OAIRealtimeEventTypeResponseContentPartAdded                         = "response.content_part.added"
	OAIRealtimeEventTypeResponseContentPartDone                          = "response.content_part.done"
	OAIRealtimeEventTypeResponseTextDelta                                = "response.text.delta"
	OAIRealtimeEventTypeResponseTextDone                                 = "response.text.done"
	OAIRealtimeEventTypeResponseAudioTranscriptDelta                     = "response.audio_transcript.delta"
	OAIRealtimeEventTypeResponseAudioTranscriptDone                      = "response.audio_transcript.done"
	OAIRealtimeEventTypeResponseAudioDelta                               = "response.audio.delta"
	OAIRealtimeEventTypeResponseFunctionCallArgumentsDelta               = "response.function_call_arguments.delta"
	OAIRealtimeEventTypeResponseFunctionCallArgumentsDone                = "response.function_call_arguments.done"
	OAIRealtimeEventTypeRateLimitsUpdated                                = "rate_limits.updated"

	// Additional constants
	// Session configuration constants
	OAIRealtimeVoiceAlloy          = "alloy"
	OAIRealtimeVoiceEcho           = "echo"
	OAIRealtimeVoiceShimmer        = "shimmer"
	OAIRealtimeAudioFormatPCM16    = "pcm16"
	OAIRealtimeAudioFormatG711ULaw = "g711_ulaw"
	OAIRealtimeAudioFormatG711ALaw = "g711_alaw"

	// Item status constants
	OAIRealtimeItemStatusCompleted  = "completed"
	OAIRealtimeItemStatusInProgress = "in_progress"
	OAIRealtimeItemStatusIncomplete = "incomplete"

	// Item role constants
	OAIRealtimeItemRoleUser      = "user"
	OAIRealtimeItemRoleAssistant = "assistant"
	OAIRealtimeItemRoleSystem    = "system"

	// Item type constants
	OAIRealtimeItemTypeMessage            = "message"
	OAIRealtimeItemTypeFunctionCall       = "function_call"
	OAIRealtimeItemTypeFunctionCallOutput = "function_call_output"

	// Content type constants
	OAIRealtimeContentTypeInputText  = "input_text"
	OAIRealtimeContentTypeInputAudio = "input_audio"
	OAIRealtimeContentTypeText       = "text"
	OAIRealtimeContentTypeAudio      = "audio"

	// Error type constants
	OAIRealtimeErrorTypeInvalidRequest = "invalid_request_error"
	OAIRealtimeErrorTypeServerError    = "server_error"

	// Rate limit name constants
	OAIRealtimeRateLimitRequests     = "requests"
	OAIRealtimeRateLimitTokens       = "tokens"
	OAIRealtimeRateLimitInputTokens  = "input_tokens"
	OAIRealtimeRateLimitOutputTokens = "output_tokens"
)

// OAIRealtimeSessionUpdateEvent represents the session.update event sent by the client
// to update the session's default configuration.
// This is a client event.
type OAIRealtimeSessionUpdateEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "session.update".
	Type string `json:"type"`
	// Session contains the session configuration to update.
	Session OAIRealtimeSession `json:"session"`
}

// OAIRealtimeSession represents the configuration for a session.
type OAIRealtimeSession struct {
	// Modalities is the set of modalities the model can respond with.
	Modalities []string `json:"modalities,omitempty"`
	// Instructions are the default system instructions prepended to model calls.
	Instructions string `json:"instructions,omitempty"`
	// Voice is the voice the model uses to respond (alloy, echo, or shimmer).
	Voice string `json:"voice,omitempty"`
	// InputAudioFormat is the format of input audio (pcm16, g711_ulaw, or g711_alaw).
	InputAudioFormat string `json:"input_audio_format,omitempty"`
	// OutputAudioFormat is the format of output audio (pcm16, g711_ulaw, or g711_alaw).
	OutputAudioFormat string `json:"output_audio_format,omitempty"`
	// InputAudioTranscription is the configuration for input audio transcription.
	InputAudioTranscription *OAIRealtimeInputAudioTranscription `json:"input_audio_transcription,omitempty"`
	// TurnDetection is the configuration for turn detection.
	TurnDetection *OAIRealtimeTurnDetection `json:"turn_detection,omitempty"`
	// Tools are the tools (functions) available to the model.
	Tools []OAIRealtimeTool `json:"tools,omitempty"`
	// ToolChoice specifies how the model chooses tools.
	ToolChoice OAIRealtimeToolChoice `json:"tool_choice,omitempty"`
	// Temperature is the sampling temperature for the model.
	Temperature float64 `json:"temperature,omitempty"`
	// MaxOutputTokens is the maximum number of output tokens for a single assistant response.
	MaxOutputTokens interface{} `json:"max_output_tokens,omitempty"`
}

// OAIRealtimeInputAudioTranscription represents the configuration for input audio transcription.
type OAIRealtimeInputAudioTranscription struct {
	// Enabled indicates whether input audio transcription is enabled.
	Enabled bool `json:"enabled"`
	// Model specifies the model used for audio transcription.
	Model string `json:"model,omitempty"`
}

// OAIRealtimeTurnDetection represents the configuration for turn detection.
type OAIRealtimeTurnDetection struct {
	// Type is the type of turn detection, currently only "server_vad" is supported.
	Type string `json:"type"`
	// Threshold is the activation threshold for VAD (0.0 to 1.0).
	Threshold float64 `json:"threshold"`
	// PrefixPaddingMs is the amount of audio to include before speech starts (in milliseconds).
	PrefixPaddingMs int `json:"prefix_padding_ms"`
	// SilenceDurationMs is the duration of silence to detect speech stop (in milliseconds).
	SilenceDurationMs int `json:"silence_duration_ms"`
}

// OAIRealtimeTool represents a tool (function) available to the model.
type OAIRealtimeTool struct {
	// Type is the type of the tool, e.g., "function".
	Type string `json:"type"`
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is the description of the function.
	Description string `json:"description"`
	// Parameters are the parameters of the function in JSON Schema.
	Parameters map[string]interface{} `json:"parameters"`
}

// OAIRealtimeInputAudioBufferAppendEvent represents the input_audio_buffer.append event
// sent by the client to append audio bytes to the input audio buffer.
// This is a client event.
type OAIRealtimeInputAudioBufferAppendEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "input_audio_buffer.append".
	Type string `json:"type"`
	// Audio contains the Base64-encoded audio bytes.
	Audio string `json:"audio"`
}

// OAIRealtimeInputAudioBufferCommitEvent represents the input_audio_buffer.commit event
// sent by the client to commit audio bytes to a user message.
// This is a client event.
type OAIRealtimeInputAudioBufferCommitEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "input_audio_buffer.commit".
	Type string `json:"type"`
}

// OAIRealtimeInputAudioBufferClearEvent represents the input_audio_buffer.clear event
// sent by the client to clear the audio bytes in the buffer.
// This is a client event.
type OAIRealtimeInputAudioBufferClearEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "input_audio_buffer.clear".
	Type string `json:"type"`
}

// OAIRealtimeConversationItemCreateEvent represents the conversation.item.create event
// sent by the client when adding an item to the conversation.
// This is a client event.
type OAIRealtimeConversationItemCreateEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "conversation.item.create".
	Type string `json:"type"`
	// PreviousItemID is the ID of the preceding item after which the new item will be inserted.
	PreviousItemID string `json:"previous_item_id"`
	// Item is the item to add to the conversation.
	Item OAIRealtimeConversationItem `json:"item"`
}

// OAIRealtimeConversationItem represents an item in the conversation.
type OAIRealtimeConversationItem struct {
	// ID is the unique ID of the item.
	ID string `json:"id"`
	// Type is the type of the item ("message", "function_call", "function_call_output").
	Type string `json:"type"`
	// Status is the status of the item ("completed", "in_progress", "incomplete").
	Status string `json:"status"`
	// Role is the role of the message sender ("user", "assistant", "system").
	Role string `json:"role"`
	// Content is the content of the message.
	Content []OAIRealtimeConversationItemContent `json:"content"`
	// CallID is the ID of the function call (for "function_call" items).
	CallID string `json:"call_id,omitempty"`
	// Name is the name of the function being called (for "function_call" items).
	Name string `json:"name,omitempty"`
	// Arguments are the arguments of the function call (for "function_call" items).
	Arguments string `json:"arguments,omitempty"`
	// Output is the output of the function call (for "function_call_output" items).
	Output string `json:"output,omitempty"`
}

// OAIRealtimeConversationItemContent represents the content of a conversation item.
type OAIRealtimeConversationItemContent struct {
	// Type is the content type ("input_text", "input_audio", "text", "audio").
	Type string `json:"type"`
	// Text is the text content.
	Text string `json:"text,omitempty"`
	// Audio contains Base64-encoded audio bytes.
	Audio string `json:"audio,omitempty"`
	// Transcript is the transcript of the audio.
	Transcript string `json:"transcript,omitempty"`
}

// OAIRealtimeConversationItemTruncateEvent represents the conversation.item.truncate event
// sent by the client to truncate a previous assistant message's audio.
// This is a client event.
type OAIRealtimeConversationItemTruncateEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "conversation.item.truncate".
	Type string `json:"type"`
	// ItemID is the ID of the assistant message item to truncate.
	ItemID string `json:"item_id"`
	// ContentIndex is the index of the content part to truncate.
	ContentIndex int `json:"content_index"`
	// AudioEndMs is the inclusive duration up to which audio is truncated, in milliseconds.
	AudioEndMs int `json:"audio_end_ms"`
}

// OAIRealtimeConversationItemDeleteEvent represents the conversation.item.delete event
// sent by the client to remove an item from the conversation history.
// This is a client event.
type OAIRealtimeConversationItemDeleteEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "conversation.item.delete".
	Type string `json:"type"`
	// ItemID is the ID of the item to delete.
	ItemID string `json:"item_id"`
}

// OAIRealtimeResponseCreateEvent represents the response.create event
// sent by the client to trigger a response generation.
// This is a client event.
type OAIRealtimeResponseCreateEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "response.create".
	Type string `json:"type"`
	// Response contains the configuration for the response.
	Response OAIRealtimeResponseConfig `json:"response"`
}

// OAIRealtimeResponseConfig represents the configuration for a response.
type OAIRealtimeResponseConfig struct {
	// Modalities are the modalities for the response.
	Modalities []string `json:"modalities,omitempty"`
	// Instructions are the instructions for the model.
	Instructions string `json:"instructions,omitempty"`
	// Voice is the voice the model uses to respond (alloy, echo, or shimmer).
	Voice string `json:"voice,omitempty"`
	// OutputAudioFormat is the format of output audio.
	OutputAudioFormat string `json:"output_audio_format,omitempty"`
	// Tools are the tools (functions) available to the model.
	Tools []OAIRealtimeTool `json:"tools,omitempty"`
	// ToolChoice specifies how the model chooses tools.
	ToolChoice OAIRealtimeToolChoice `json:"tool_choice,omitempty"`
	// Temperature is the sampling temperature for the model.
	Temperature float64 `json:"temperature,omitempty"`
	// MaxOutputTokens is the maximum number of output tokens for a single assistant response.
	MaxOutputTokens interface{} `json:"max_output_tokens,omitempty"`
}

// OAIRealtimeResponseCancelEvent represents the response.cancel event
// sent by the client to cancel an in-progress response.
// This is a client event.
type OAIRealtimeResponseCancelEvent struct {
	// EventID is an optional client-generated ID used to identify this event.
	EventID string `json:"event_id,omitempty"`
	// Type is the event type, must be "response.cancel".
	Type string `json:"type"`
}

// OAIRealtimeErrorEvent represents the error event emitted by the server
// when an error occurs.
// This is a server event.
type OAIRealtimeErrorEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "error".
	Type string `json:"type"`
	// Error contains details of the error.
	Error OAIRealtimeError `json:"error"`
}

// OAIRealtimeError represents the details of an error.
type OAIRealtimeError struct {
	// Type is the type of error (e.g., "invalid_request_error", "server_error").
	Type string `json:"type"`
	// Code is the error code, if any.
	Code string `json:"code,omitempty"`
	// Message is a human-readable error message.
	Message string `json:"message"`
	// Param is the parameter related to the error, if any.
	Param string `json:"param,omitempty"`
	// EventID is the event_id of the client event that caused the error, if applicable.
	EventID string `json:"event_id,omitempty"`
}

// OAIRealtimeSessionCreatedEvent represents the session.created event emitted by the server
// when a new session is created.
// This is a server event.
type OAIRealtimeSessionCreatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "session.created".
	Type string `json:"type"`
	// Session contains the session resource.
	Session OAIRealtimeSessionResource `json:"session"`
}

// OAIRealtimeSessionResource represents the session resource returned by the server.
type OAIRealtimeSessionResource struct {
	// ID is the unique ID of the session.
	ID string `json:"id"`
	// Object is the object type, must be "realtime.session".
	Object string `json:"object"`
	// Model is the default model used for this session.
	Model string `json:"model"`
	// Modalities is the set of modalities the model can respond with.
	Modalities []string `json:"modalities"`
	// Instructions are the default system instructions.
	Instructions string `json:"instructions"`
	// Voice is the voice the model uses to respond (alloy, echo, or shimmer).
	Voice string `json:"voice"`
	// InputAudioFormat is the format of input audio.
	InputAudioFormat string `json:"input_audio_format"`
	// OutputAudioFormat is the format of output audio.
	OutputAudioFormat string `json:"output_audio_format"`
	// InputAudioTranscription is the configuration for input audio transcription.
	InputAudioTranscription *OAIRealtimeInputAudioTranscription `json:"input_audio_transcription"`
	// TurnDetection is the configuration for turn detection.
	TurnDetection *OAIRealtimeTurnDetection `json:"turn_detection"`
	// Tools are the tools (functions) available to the model.
	Tools []OAIRealtimeTool `json:"tools"`
	// ToolChoice specifies how the model chooses tools.
	ToolChoice OAIRealtimeToolChoice `json:"tool_choice"`
	// Temperature is the sampling temperature for the model.
	Temperature float64 `json:"temperature"`
	// MaxOutputTokens is the maximum number of output tokens for a single assistant response.
	MaxOutputTokens interface{} `json:"max_output_tokens"`
}

// OAIRealtimeSessionUpdatedEvent represents the session.updated event emitted by the server
// when a session is updated.
// This is a server event.
type OAIRealtimeSessionUpdatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "session.updated".
	Type string `json:"type"`
	// Session contains the updated session resource.
	Session OAIRealtimeSessionResource `json:"session"`
}

// OAIRealtimeConversationCreatedEvent represents the conversation.created event emitted by the server
// when a new conversation is created.
// This is a server event.
type OAIRealtimeConversationCreatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "conversation.created".
	Type string `json:"type"`
	// Conversation contains the conversation resource.
	Conversation OAIRealtimeConversationResource `json:"conversation"`
}

// OAIRealtimeConversationResource represents the conversation resource returned by the server.
type OAIRealtimeConversationResource struct {
	// ID is the unique ID of the conversation.
	ID string `json:"id"`
	// Object is the object type, must be "realtime.conversation".
	Object string `json:"object"`
}

// OAIRealtimeConversationItemCreatedEvent represents the conversation.item.created event
// emitted by the server when a conversation item is created.
// This is a server event.
type OAIRealtimeConversationItemCreatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "conversation.item.created".
	Type string `json:"type"`
	// PreviousItemID is the ID of the preceding item.
	PreviousItemID string `json:"previous_item_id"`
	// Item is the item that was created.
	Item OAIRealtimeConversationItemResource `json:"item"`
}

// OAIRealtimeConversationItemResource represents the conversation item resource returned by the server.
type OAIRealtimeConversationItemResource struct {
	// ID is the unique ID of the item.
	ID string `json:"id"`
	// Object is the object type, must be "realtime.item".
	Object string `json:"object"`
	// Type is the type of the item ("message", "function_call", "function_call_output").
	Type string `json:"type"`
	// Status is the status of the item ("completed", "in_progress", "incomplete").
	Status string `json:"status"`
	// Role is the role associated with the item ("user", "assistant", "system").
	Role string `json:"role"`
	// Content is the content of the item.
	Content []OAIRealtimeConversationItemContent `json:"content"`
	// CallID is the ID of the function call (for "function_call" items).
	CallID string `json:"call_id,omitempty"`
	// Name is the name of the function being called.
	Name string `json:"name,omitempty"`
	// Arguments are the arguments of the function call.
	Arguments string `json:"arguments,omitempty"`
	// Output is the output of the function call (for "function_call_output" items).
	Output string `json:"output,omitempty"`
}

// OAIRealtimeConversationItemInputAudioTranscriptionCompletedEvent represents the
// conversation.item.input_audio_transcription.completed event emitted by the server
// when input audio transcription is enabled and a transcription succeeds.
// This is a server event.
type OAIRealtimeConversationItemInputAudioTranscriptionCompletedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "conversation.item.input_audio_transcription.completed".
	Type string `json:"type"`
	// ItemID is the ID of the user message item.
	ItemID string `json:"item_id"`
	// ContentIndex is the index of the content part containing the audio.
	ContentIndex int `json:"content_index"`
	// Transcript is the transcribed text.
	Transcript string `json:"transcript"`
}

// OAIRealtimeConversationItemInputAudioTranscriptionFailedEvent represents the
// conversation.item.input_audio_transcription.failed event emitted by the server
// when input audio transcription is configured, and a transcription request for a user message failed.
// This is a server event.
type OAIRealtimeConversationItemInputAudioTranscriptionFailedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "conversation.item.input_audio_transcription.failed".
	Type string `json:"type"`
	// ItemID is the ID of the user message item.
	ItemID string `json:"item_id"`
	// ContentIndex is the index of the content part containing the audio.
	ContentIndex int `json:"content_index"`
	// Error contains details of the transcription error.
	Error OAIRealtimeTranscriptionError `json:"error"`
}

// OAIRealtimeTranscriptionError represents the details of a transcription error.
type OAIRealtimeTranscriptionError struct {
	// Type is the type of error.
	Type string `json:"type"`
	// Code is the error code, if any.
	Code string `json:"code,omitempty"`
	// Message is a human-readable error message.
	Message string `json:"message"`
	// Param is the parameter related to the error, if any.
	Param string `json:"param,omitempty"`
}

// OAIRealtimeConversationItemTruncatedEvent represents the conversation.item.truncated event
// emitted by the server when an earlier assistant audio message item is truncated by the client.
// This is a server event.
type OAIRealtimeConversationItemTruncatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "conversation.item.truncated".
	Type string `json:"type"`
	// ItemID is the ID of the assistant message item that was truncated.
	ItemID string `json:"item_id"`
	// ContentIndex is the index of the content part that was truncated.
	ContentIndex int `json:"content_index"`
	// AudioEndMs is the duration up to which the audio was truncated, in milliseconds.
	AudioEndMs int `json:"audio_end_ms"`
}

// OAIRealtimeConversationItemDeletedEvent represents the conversation.item.deleted event
// emitted by the server when an item in the conversation is deleted.
// This is a server event.
type OAIRealtimeConversationItemDeletedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "conversation.item.deleted".
	Type string `json:"type"`
	// ItemID is the ID of the item that was deleted.
	ItemID string `json:"item_id"`
}

// OAIRealtimeInputAudioBufferCommittedEvent represents the input_audio_buffer.committed event
// emitted by the server when an input audio buffer is committed, either by the client or
// automatically in server VAD mode.
// This is a server event.
type OAIRealtimeInputAudioBufferCommittedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "input_audio_buffer.committed".
	Type string `json:"type"`
	// PreviousItemID is the ID of the preceding item after which the new item will be inserted.
	PreviousItemID string `json:"previous_item_id"`
	// ItemID is the ID of the user message item that will be created.
	ItemID string `json:"item_id"`
}

// OAIRealtimeInputAudioBufferClearedEvent represents the input_audio_buffer.cleared event
// emitted by the server when the input audio buffer is cleared by the client.
// This is a server event.
type OAIRealtimeInputAudioBufferClearedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "input_audio_buffer.cleared".
	Type string `json:"type"`
}

// OAIRealtimeInputAudioBufferSpeechStartedEvent represents the input_audio_buffer.speech_started event
// emitted by the server in server turn detection mode when speech is detected.
// This is a server event.
type OAIRealtimeInputAudioBufferSpeechStartedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "input_audio_buffer.speech_started".
	Type string `json:"type"`
	// AudioStartMs is the milliseconds since the session started when speech was detected.
	AudioStartMs int `json:"audio_start_ms"`
	// ItemID is the ID of the user message item that will be created when speech stops.
	ItemID string `json:"item_id"`
}

// OAIRealtimeInputAudioBufferSpeechStoppedEvent represents the input_audio_buffer.speech_stopped event
// emitted by the server in server turn detection mode when speech stops.
// This is a server event.
type OAIRealtimeInputAudioBufferSpeechStoppedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "input_audio_buffer.speech_stopped".
	Type string `json:"type"`
	// AudioEndMs is the milliseconds since the session started when speech stopped.
	AudioEndMs int `json:"audio_end_ms"`
	// ItemID is the ID of the user message item that will be created.
	ItemID string `json:"item_id"`
}

// OAIRealtimeResponseCreatedEvent represents the response.created event
// emitted by the server when a new Response is created.
// This is a server event.
type OAIRealtimeResponseCreatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.created".
	Type string `json:"type"`
	// Response contains the response resource.
	Response OAIRealtimeResponseResource `json:"response"`
}

// OAIRealtimeResponseResource represents the response resource returned by the server.
type OAIRealtimeResponseResource struct {
	// ID is the unique ID of the response.
	ID string `json:"id"`
	// Object is the object type, must be "realtime.response".
	Object string `json:"object"`
	// Status is the status of the response ("in_progress").
	Status string `json:"status"`
	// StatusDetails contains additional details about the status.
	StatusDetails interface{} `json:"status_details"`
	// Output is the list of output items generated by the response.
	Output []interface{} `json:"output"`
	// Usage contains usage statistics for the response.
	Usage interface{} `json:"usage"`
}

// OAIRealtimeResponseDoneEvent represents the response.done event
// emitted by the server when a Response is done streaming.
// This is a server event.
type OAIRealtimeResponseDoneEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.done".
	Type string `json:"type"`
	// Response contains the response resource.
	Response OAIRealtimeResponseResource `json:"response"`
}

// OAIRealtimeResponseUsage represents the usage statistics for a response.
type OAIRealtimeResponseUsage struct {
	// TotalTokens is the total number of tokens used in the response.
	TotalTokens int `json:"total_tokens"`
	// InputTokens is the number of input tokens used.
	InputTokens int `json:"input_tokens"`
	// OutputTokens is the number of output tokens generated.
	OutputTokens int `json:"output_tokens"`
}

// OAIRealtimeResponseOutputItemAddedEvent represents the response.output_item.added event
// emitted by the server when a new Item is created during response generation.
// This is a server event.
type OAIRealtimeResponseOutputItemAddedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.output_item.added".
	Type string `json:"type"`
	// ResponseID is the ID of the response to which the item belongs.
	ResponseID string `json:"response_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// Item is the item that was added.
	Item OAIRealtimeConversationItemResource `json:"item"`
}

// OAIRealtimeResponseOutputItemDoneEvent represents the response.output_item.done event
// emitted by the server when an Item is done streaming or when a Response is interrupted,
// incomplete, or cancelled.
// This is a server event.
type OAIRealtimeResponseOutputItemDoneEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.output_item.done".
	Type string `json:"type"`
	// ResponseID is the ID of the response to which the item belongs.
	ResponseID string `json:"response_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// Item is the completed item.
	Item OAIRealtimeConversationItemResource `json:"item"`
}

// OAIRealtimeResponseContentPartAddedEvent represents the response.content_part.added event
// emitted by the server when a new content part is added to an assistant message item during response generation.
// This is a server event.
type OAIRealtimeResponseContentPartAddedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.content_part.added".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item to which the content part was added.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Part is the content part that was added.
	Part OAIRealtimeConversationItemContent `json:"part"`
}

// OAIRealtimeResponseContentPartDoneEvent represents the response.content_part.done event
// emitted by the server when a content part is done streaming in an assistant message item.
// Also emitted when a Response is interrupted, incomplete, or cancelled.
// This is a server event.
type OAIRealtimeResponseContentPartDoneEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.content_part.done".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Part is the content part that is done.
	Part OAIRealtimeConversationItemContent `json:"part"`
}

// OAIRealtimeResponseTextDeltaEvent represents the response.text.delta event
// emitted by the server when the text value of a "text" content part is updated.
// This is a server event.
type OAIRealtimeResponseTextDeltaEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.text.delta".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Delta is the text delta.
	Delta string `json:"delta"`
}

// OAIRealtimeResponseTextDoneEvent represents the response.text.done event
// emitted by the server when the text value of a "text" content part is done streaming.
// Also emitted when a Response is interrupted, incomplete, or cancelled.
// This is a server event.
type OAIRealtimeResponseTextDoneEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.text.done".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Text is the final text content.
	Text string `json:"text"`
}

// OAIRealtimeResponseAudioTranscriptDeltaEvent represents the response.audio_transcript.delta event
// emitted by the server when the model-generated transcription of audio output is updated.
// This is a server event.
type OAIRealtimeResponseAudioTranscriptDeltaEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.audio_transcript.delta".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Delta is the transcript delta.
	Delta string `json:"delta"`
}

// OAIRealtimeResponseAudioTranscriptDoneEvent represents the response.audio_transcript.done event
// emitted by the server when the model-generated transcription of audio output is done streaming.
// Also emitted when a Response is interrupted, incomplete, or cancelled.
// This is a server event.
type OAIRealtimeResponseAudioTranscriptDoneEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.audio_transcript.done".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Transcript is the final transcript of the audio.
	Transcript string `json:"transcript"`
}

// OAIRealtimeResponseAudioDeltaEvent represents the response.audio.delta event
// emitted by the server when the model-generated audio is updated.
// This is a server event.
type OAIRealtimeResponseAudioDeltaEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.audio.delta".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// ContentIndex is the index of the content part in the item's content array.
	ContentIndex int `json:"content_index"`
	// Delta is the Base64-encoded audio data delta.
	Delta string `json:"delta"`
}

// OAIRealtimeResponseFunctionCallArgumentsDeltaEvent represents the response.function_call_arguments.delta event
// emitted by the server when the model-generated function call arguments are updated.
// This is a server event.
type OAIRealtimeResponseFunctionCallArgumentsDeltaEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.function_call_arguments.delta".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the function call item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// CallID is the ID of the function call.
	CallID string `json:"call_id"`
	// Delta is the arguments delta as a JSON string.
	Delta string `json:"delta"`
}

// OAIRealtimeResponseFunctionCallArgumentsDoneEvent represents the response.function_call_arguments.done event
// emitted by the server when the model-generated function call arguments are done streaming.
// Also emitted when a Response is interrupted, incomplete, or cancelled.
// This is a server event.
type OAIRealtimeResponseFunctionCallArgumentsDoneEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "response.function_call_arguments.done".
	Type string `json:"type"`
	// ResponseID is the ID of the response.
	ResponseID string `json:"response_id"`
	// ItemID is the ID of the function call item.
	ItemID string `json:"item_id"`
	// OutputIndex is the index of the output item in the response.
	OutputIndex int `json:"output_index"`
	// CallID is the ID of the function call.
	CallID string `json:"call_id"`
	// Arguments is the final arguments as a JSON string.
	Arguments string `json:"arguments"`
}

// OAIRealtimeRateLimitsUpdatedEvent represents the rate_limits.updated event
// emitted by the server after every "response.done" event to indicate the updated rate limits.
// This is a server event.
type OAIRealtimeRateLimitsUpdatedEvent struct {
	// EventID is the unique ID of the server event.
	EventID string `json:"event_id"`
	// Type is the event type, must be "rate_limits.updated".
	Type string `json:"type"`
	// RateLimits is the list of rate limit information.
	RateLimits []OAIRealtimeRateLimit `json:"rate_limits"`
}

// OAIRealtimeRateLimit represents the rate limit information for a specific limit type.
type OAIRealtimeRateLimit struct {
	// Name is the name of the rate limit ("requests", "tokens", "input_tokens", "output_tokens").
	Name string `json:"name"`
	// Limit is the maximum allowed value for the rate limit.
	Limit int `json:"limit"`
	// Remaining is the remaining value before the limit is reached.
	Remaining int `json:"remaining"`
	// ResetSeconds is the number of seconds until the rate limit resets.
	ResetSeconds float64 `json:"reset_seconds"`
}

// OAIRealtimeToolChoice represents the tool choice options
type OAIRealtimeToolChoice string

const (
	OAIRealtimeToolChoiceAuto    OAIRealtimeToolChoice = "auto"
	OAIRealtimeToolChoiceNone    OAIRealtimeToolChoice = "none"
	OAIRealtimeToolChoiceDefault OAIRealtimeToolChoice = "default"
)
