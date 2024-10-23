package xai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
)

// OpenAIClientOption is a function type for configuring the OpenAIClient
type OpenAIClientOption func(*OpenAIClient)

// TextGenerationOptions contains options for generating text
type TextGenerationOptions struct {
	Model        string                  `json:"model"`
	Prompt       string                  `json:"prompt"`
	SystemPrompt string                  `json:"system"`
	Messages     []ChatCompletionMessage `json:"messages"`
	IsStreaming  bool                    `json:"stream"`
	ObjectSchema string                  `json:"object_schema"`
	Temperature  float64                 `json:"temperature,omitempty"`
	MaxTokens    int                     `json:"max_tokens,omitempty"`
	TopP         float64                 `json:"top_p,omitempty"`
	N            int                     `json:"n,omitempty"`
	ChunkSize    int                     `json:"chunk_size,omitempty"`
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Image   string `json:"image,omitempty"`
}

// ChatCompletionResponse represents the response from the chat completion API
type ChatCompletionResponse struct {
	ID         string   `json:"id"`
	ObjectType string   `json:"object"`
	CreatedAt  int64    `json:"created"`
	ModelName  string   `json:"model"`
	UsageInfo  Usage    `json:"usage"`
	Choices    []Choice `json:"choices"`
}

// StreamChoice represents a single choice in the streaming response
type StreamChoice struct {
	Delta        StreamDelta `json:"delta"`
	Index        int         `json:"index"`
	FinishReason string      `json:"finish_reason"`
}

// StreamDelta represents the delta content in a streaming response
type StreamDelta struct {
	Content string `json:"content"`
}

// EmbeddingRequest represents the request for creating embeddings
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents the response from the embeddings API
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// ImageGenerationRequest represents the request for generating images
type ImageGenerationRequest struct {
	Model                   string `json:"model"`
	Prompt                  string `json:"prompt"`
	N                       int    `json:"n,omitempty"`
	Size                    string `json:"size,omitempty"`
	Quality                 string `json:"quality,omitempty"`
	ResponseFormat          string `json:"response_format,omitempty"`
	UseMidjourney           bool   `json:"use_midjourney,omitempty"`
	MidjourneyAction        string `json:"midjourney_action,omitempty"`
	MidjourneyActionContent string `json:"midjourney_action_content,omitempty"`
}

// ImageGenerationResponse represents the response from the image generation API
type ImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url,omitempty"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
}

// Constants for message roles and models
const (
	MessageRoleSystem    = "system"
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"

	DefaultModel        = "gpt-3.5-turbo"
	ModelGPT4o          = "gpt-4o"
	ModelClaude35Sonnet = "claude-3-5-sonnet-20240620"
)

// Environment variable keys
const (
	EnvOpenAIAPIKey  = "OPENAI_API_KEY"
	EnvOpenAIBaseURL = "OPENAI_API_BASE"
)

// API endpoints
const (
	MidjourneyURL         = "/mj/submit/imagine"
	MidjourneyStatusURL   = "/mj/task/{id}/fetch"
	MidjourneyActionURL   = "/mj/submit/action"
	DefaultEmbeddingModel = "text-embedding-ada-002"
	DefaultImageModel     = "dall-e-3"
	DefaultChunkSize      = 100 // Default chunk size for streaming
)

const (
	MidjourneyActionSubmit = "submit"
	MidjourneyActionStatus = "status"
	MidjourneyActionAction = "action"
)

// WithBaseURL sets the base URL for the OpenAI API
func WithBaseURL(url string) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.baseURL = processBaseURL(url)
	}
}

func processBaseURL(url string) string {
	return x.TrimSuffixes(url, "/", "/v1")
}

// WithAPIKey sets the API key for authentication
func WithAPIKey(key string) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.apiKey = key
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *xhttpc.Client) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.httpClient = client
	}
}

// WithModel sets the default model for the OpenAI client
func WithModel(model string) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.model = model
	}
}

// WithDebug enables or disables debug mode for the OpenAI client
func WithDebug(debug bool) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.debug = debug
	}
}

// WithHttpClientDebug enables or disables debug mode for the HTTP client
func WithHttpClientDebug(debug bool) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.httpClient.SetDebug(debug)
	}
}

// NewOpenAIClient creates a new OpenAIClient with the given options
func NewOpenAIClient(options ...OpenAIClientOption) *OpenAIClient {
	client := &OpenAIClient{
		baseURL: DefaultBaseURL,
		httpClient: x.Must1(xhttpc.NewClient(
			xhttpc.WithTimeout(30 * time.Second),
		)),
		model: DefaultModel,
		debug: false,
	}

	client.loadEnvironmentVariables()

	for _, option := range options {
		option(client)
	}
	client.baseURL = processBaseURL(client.baseURL)

	return client
}

func (c *OpenAIClient) loadEnvironmentVariables() {
	if apiKey := os.Getenv(EnvOpenAIAPIKey); apiKey != "" {
		c.apiKey = apiKey
	}
	if baseURL := os.Getenv(EnvOpenAIBaseURL); baseURL != "" {
		c.baseURL = baseURL
	}
}

// validateTextGenerationOptions checks if the provided options are valid
func validateTextGenerationOptions(options *TextGenerationOptions) error {
	hasMessages := len(options.Messages) > 0
	hasPromptOrSystem := options.Prompt != "" || options.SystemPrompt != ""

	if hasMessages && hasPromptOrSystem {
		return xerror.New("either 'Messages' or 'Prompt'/'SystemPrompt' should be provided, not both")
	}

	if !hasMessages && !hasPromptOrSystem {
		return xerror.New("either 'Messages' or 'Prompt'/'SystemPrompt' must be provided")
	}

	return nil
}

// GenerateText generates text based on the provided options
func (c *OpenAIClient) GenerateText(ctx context.Context, options TextGenerationOptions) (string, error) {
	if err := validateTextGenerationOptions(&options); err != nil {
		return "", xerror.Wrap(err, "invalid text generation options")
	}

	if x.IsEmpty(options.Model) {
		options.Model = c.model
	}

	requestBody := map[string]interface{}{
		"model":    options.Model,
		"messages": options.Messages,
	}

	x.SetNonZeroValuesWithKeys(requestBody, map[string]interface{}{
		"temperature": options.Temperature,
		"max_tokens":  options.MaxTokens,
		"top_p":       options.TopP,
		"n":           options.N,
		"stream":      options.IsStreaming,
	})

	resp, err := c.sendRequest(ctx, http.MethodPost, ChatCompletionsEndpoint, requestBody, false)
	if err != nil {
		return "", xerror.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", xerror.Wrap(err, "error decoding response")
	}

	if len(result.Choices) == 0 {
		return "", xerror.New("no choices returned from API")
	}

	return result.Choices[0].Message.Content, nil
}

// GenerateTextStream generates text in a streaming fashion
func (c *OpenAIClient) GenerateTextStream(ctx context.Context, options TextGenerationOptions) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		if err := validateTextGenerationOptions(&options); err != nil {
			errChan <- xerror.Wrap(err, "invalid text generation options")
			return
		}

		if x.IsEmpty(options.Model) {
			options.Model = c.model
		}

		requestBody := map[string]interface{}{
			"model":    options.Model,
			"messages": options.Messages,
			"stream":   true,
		}

		x.SetNonZeroValuesWithKeys(requestBody, map[string]interface{}{
			"temperature": options.Temperature,
			"max_tokens":  options.MaxTokens,
			"top_p":       options.TopP,
			"n":           options.N,
		})

		responseChan, responseErrChan := c.httpClient.StreamResponse(ctx, http.MethodPost, c.baseURL+ChatCompletionsEndpoint, requestBody, xhttpc.WithStreamContentType("application/json"))

		buffer := strings.Builder{}
		lastOutputTime := time.Now()
		const maxOutputInterval = 500 * time.Millisecond
		const minChunkSize = 10 // Minimum number of characters to output

		flushBuffer := func() {
			if buffer.Len() > 0 {
				textChan <- buffer.String()
				buffer.Reset()
				lastOutputTime = time.Now()
			}
		}

		for {
			select {
			case <-ctx.Done():
				flushBuffer()
				return
			case err := <-responseErrChan:
				flushBuffer()
				if err != nil {
					errChan <- xerror.Wrap(err, "error in stream response")
				}
				return
			case chunk, ok := <-responseChan:
				if !ok {
					flushBuffer()
					return
				}
				line := strings.TrimSpace(string(chunk))
				if line == "" || line == "data: [DONE]" {
					flushBuffer()
					continue
				}
				if !strings.HasPrefix(line, "data: ") {
					flushBuffer()
					errChan <- xerror.Newf("unexpected line format: %s", line)
					return
				}

				data := strings.TrimPrefix(line, "data: ")
				var streamResponse ChatCompletionChunk
				if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
					flushBuffer()
					errChan <- xerror.Wrap(err, "error unmarshaling stream data")
					return
				}

				if streamResponse.Response != nil &&
					len(streamResponse.Response.Choices) > 0 &&
					streamResponse.Response.Choices[0].Delta.Content != "" {
					buffer.WriteString(streamResponse.Response.Choices[0].Delta.Content)
					if buffer.Len() >= options.ChunkSize || time.Since(lastOutputTime) >= maxOutputInterval {
						flushBuffer()
					}
				}

				if buffer.Len() >= minChunkSize && time.Since(lastOutputTime) >= maxOutputInterval {
					flushBuffer()
				}
			}
		}
	}()

	return textChan, errChan
}

func (c *OpenAIClient) sendRequest(ctx context.Context, method, endpoint string, body map[string]interface{}, isMidjourney bool) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var resp *http.Response
	var err error

	if isMidjourney {
		xlog.Debug("sending request to midjourney")
	}

	switch method {
	case http.MethodGet:
		resp, err = c.httpClient.
			SetBaseURL(c.baseURL).
			SetBearerToken(c.apiKey).
			Get(ctx, endpoint)
	case http.MethodPost:
		resp, err = c.httpClient.
			SetBaseURL(c.baseURL).
			SetBearerToken(c.apiKey).
			PostJSON(ctx, endpoint, body)
	default:
		return nil, xerror.Newf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, xerror.Wrap(err, "error sending request")
	}

	if resp == nil {
		return nil, xerror.New("received nil response")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, xerror.Newf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func (c *OpenAIClient) handleStreamResponse(ctx context.Context, resp *http.Response, textChan chan<- string, errChan chan<- error, chunkSize int) {
	if resp == nil || resp.Body == nil {
		errChan <- xerror.New("invalid response or response body")
		return
	}

	reader := bufio.NewReader(resp.Body)
	buffer := strings.Builder{}

	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	defer func() {
		resp.Body.Close()
		if r := recover(); r != nil {
			errChan <- xerror.Newf("panic in stream handler: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if buffer.Len() > 0 {
						textChan <- buffer.String()
					}
					return
				}
				errChan <- xerror.Wrap(err, "error reading stream")
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || line == "data: [DONE]" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				errChan <- xerror.Newf("unexpected line format: %s", line)
				return
			}

			data := strings.TrimPrefix(line, "data: ")
			var streamResponse ChatCompletionChunk
			if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
				errChan <- xerror.Wrap(err, "error unmarshaling stream data")
				return
			}

			if streamResponse.Response != nil &&
				len(streamResponse.Response.Choices) > 0 &&
				streamResponse.Response.Choices[0].Delta.Content != "" {
				buffer.WriteString(streamResponse.Response.Choices[0].Delta.Content)
				if buffer.Len() >= chunkSize {
					select {
					case textChan <- buffer.String():
						buffer.Reset()
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

// prepareTextGenerationOptions prepares the options for text generation
func (c *OpenAIClient) prepareTextGenerationOptions(prompt []string, options ...func(*TextGenerationOptions)) TextGenerationOptions {
	opts := TextGenerationOptions{
		Model:     c.model,
		Messages:  []ChatCompletionMessage{},
		ChunkSize: DefaultChunkSize,
	}

	for _, option := range options {
		option(&opts)
	}

	if opts.SystemPrompt != "" {
		opts.Messages = append(opts.Messages, ChatCompletionMessage{Role: MessageRoleSystem, Content: opts.SystemPrompt})
	}

	for i, content := range prompt {
		role := MessageRoleUser
		if i%2 != 0 {
			role = MessageRoleAssistant
		}

		message := ChatCompletionMessage{Role: role}
		if x.IsImageURL(content) || x.IsBase64(content) {
			message.Image = content
		} else {
			message.Content = content
		}
		opts.Messages = append(opts.Messages, message)
	}
	return opts
}

// QuickGenerateText is a convenience method for generating text
func (c *OpenAIClient) QuickGenerateText(ctx context.Context, prompt []string, options ...func(*TextGenerationOptions)) (string, error) {
	opts := c.prepareTextGenerationOptions(prompt, options...)
	return c.GenerateText(ctx, opts)
}

// QuickGenerateTextStream is a convenience method for generating text in a streaming fashion
func (c *OpenAIClient) QuickGenerateTextStream(ctx context.Context, prompt []string, options ...func(*TextGenerationOptions)) (<-chan string, <-chan error) {
	opts := c.prepareTextGenerationOptions(prompt, options...)
	return c.GenerateTextStream(ctx, opts)
}

// WithTextModel sets the model for text generation
func WithTextModel(model string) func(*TextGenerationOptions) {
	return func(opts *TextGenerationOptions) {
		opts.Model = model
	}
}

// WithChunkSize sets the chunk size for streaming text generation
func WithChunkSize(size int) func(*TextGenerationOptions) {
	return func(opts *TextGenerationOptions) {
		opts.ChunkSize = size
	}
}

// CreateEmbeddings generates embeddings for the given input
func (c *OpenAIClient) CreateEmbeddings(ctx context.Context, input []string, model string) ([][]float64, error) {
	if model == "" {
		model = DefaultEmbeddingModel
	}

	requestBody := map[string]interface{}{
		"model": model,
		"input": input,
	}

	resp, err := c.sendRequest(ctx, http.MethodPost, EmbeddingsEndpoint, requestBody, false)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to send embedding request")
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, xerror.Wrap(err, "error decoding embedding response")
	}

	embeddings := make([][]float64, len(result.Data))
	for i, data := range result.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// GetMidjourneyStatus gets the status of a Midjourney job
func (c *OpenAIClient) GetMidjourneyStatus(ctx context.Context, jobID string) (*MidjourneyResponse, error) {
	return c.GenerateImageWithMidjourney(ctx, jobID, WithMidjourneyAction(MidjourneyActionStatus))
}

// ActMidjourney is a convenience method for generating an image based on the provided prompt using Midjourney
func (c *OpenAIClient) ActMidjourney(ctx context.Context, actionContent string, jobID string) (*MidjourneyResponse, error) {
	return c.GenerateImageWithMidjourney(ctx, jobID, WithMidjourneyAction(MidjourneyActionAction), WithMidjourneyActionContent(actionContent))
}

// GetFileIDFromMidjourneySuccessResponse parses the file ID from a successful Midjourney response
func GetFileIDFromMidjourneySuccessResponse(response *MidjourneyResponse) (string, error) {
	if response == nil || len(response.Buttons) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	parts := strings.Split(response.Buttons[0].CustomID, "::")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid result format")
	}

	return parts[2], nil
}

// MidjourneyAction represents different actions that can be performed on Midjourney images
const (
	MJJOBUpsample      = "MJ::JOB::upsample"
	MJJOBVariation     = "MJ::JOB::variation"
	MJJOBReroll        = "MJ::JOB::reroll"
	MJJOBUpsampleV52x  = "MJ::JOB::upsample_v5_2x"
	MJJOBUpsampleV54x  = "MJ::JOB::upsample_v5_4x"
	MJJOBLowVariation  = "MJ::JOB::low_variation"
	MJJOBHighVariation = "MJ::JOB::high_variation"
	MJInpaint          = "MJ::Inpaint"
	MJOutpaint50       = "MJ::Outpaint::50"
	MJOutpaint75       = "MJ::Outpaint::75"
	MJCustomZoom       = "MJ::CustomZoom"
	MJJOBPanLeft       = "MJ::JOB::pan_left"
	MJJOBPanRight      = "MJ::JOB::pan_right"
	MJJOBPanUp         = "MJ::JOB::pan_up"
	MJJOBPanDown       = "MJ::JOB::pan_down"
	MJBOOKMARK         = "MJ::BOOKMARK"
)

// BuildMidjourneyActionContent builds the action content for Midjourney
func BuildMidjourneyActionContent(action, number, fileID string) string {
	switch action {
	case MJCustomZoom, MJBOOKMARK:
		return fmt.Sprintf("%s::%s", action, fileID)
	case MJJOBUpsample, MJJOBVariation:
		return fmt.Sprintf("%s::%s::%s", action, number, fileID)
	case MJJOBReroll:
		return fmt.Sprintf("%s::0::%s::SOLO", action, fileID)
	default:
		return fmt.Sprintf("%s::%s::%s::SOLO", action, number, fileID)
	}
}

// GenerateImageWithMidjourney generates an image based on the provided prompt using Midjourney
func (c *OpenAIClient) GenerateImageWithMidjourney(ctx context.Context, prompt string, options ...func(*ImageGenerationRequest)) (*MidjourneyResponse, error) {
	var opts = ImageGenerationRequest{
		Prompt:           prompt,
		UseMidjourney:    true,
		MidjourneyAction: MidjourneyActionSubmit,
	}

	for _, option := range options {
		option(&opts)
	}

	var requestBody map[string]interface{}
	var requestPath string
	var method string

	switch opts.MidjourneyAction {
	case MidjourneyActionSubmit:
		requestPath = MidjourneyURL
		method = http.MethodPost
		requestBody = map[string]interface{}{
			"prompt": prompt,
		}
	case MidjourneyActionStatus:
		requestPath = strings.Replace(MidjourneyStatusURL, "{id}", prompt, 1)
		method = http.MethodGet
		requestBody = nil
	case MidjourneyActionAction:
		requestPath = MidjourneyActionURL
		method = http.MethodPost
		requestBody = map[string]interface{}{
			"taskId":   prompt,
			"customId": opts.MidjourneyActionContent,
		}
	default:
		return nil, fmt.Errorf("invalid midjourney action: %s", opts.MidjourneyAction)
	}

	resp, err := c.sendRequest(ctx, method, requestPath, requestBody, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result MidjourneyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

type MidjourneyResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	Result      string `json:"result"`
	Properties  struct {
		DiscordChannelId  string `json:"discordChannelId"`
		DiscordInstanceId string `json:"discordInstanceId"`
	} `json:"properties"`
	ID         string `json:"id"`
	Action     string `json:"action"`
	CustomID   string `json:"customId"`
	BotType    string `json:"botType"`
	Prompt     string `json:"prompt"`
	PromptEn   string `json:"promptEn"`
	State      string `json:"state"`
	SubmitTime int64  `json:"submitTime"`
	StartTime  int64  `json:"startTime"`
	FinishTime int64  `json:"finishTime"`
	ImageURL   string `json:"imageUrl"`
	Status     string `json:"status"`
	Progress   string `json:"progress"`
	FailReason string `json:"failReason"`
	Buttons    []struct {
		CustomID string `json:"customId"`
		Emoji    string `json:"emoji"`
		Label    string `json:"label"`
		Type     int    `json:"type"`
		Style    int    `json:"style"`
	} `json:"buttons"`
	MaskBase64    string `json:"maskBase64"`
	FinalPrompt   string `json:"finalPrompt"`
	FinalZhPrompt string `json:"finalZhPrompt"`
}

// GenerateImage generates an image based on the provided prompt
func (c *OpenAIClient) GenerateImage(ctx context.Context, prompt string, options ...func(*ImageGenerationRequest)) ([]string, error) {
	requestBody := map[string]interface{}{
		"model":  DefaultImageModel,
		"prompt": prompt,
	}

	for _, option := range options {
		var opts ImageGenerationRequest
		option(&opts)
		if opts.N != 0 {
			requestBody["n"] = opts.N
		}
		if opts.Size != "" {
			requestBody["size"] = opts.Size
		}
		if opts.Quality != "" {
			requestBody["quality"] = opts.Quality
		}
		if opts.ResponseFormat != "" {
			requestBody["response_format"] = opts.ResponseFormat
		}
	}

	resp, err := c.sendRequest(ctx, http.MethodPost, ImagesGenerationsEndpoint, requestBody, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ImageGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	urls := make([]string, len(result.Data))
	for i, data := range result.Data {
		if data.URL != "" {
			urls[i] = data.URL
		} else {
			urls[i] = data.B64JSON
		}
	}

	return urls, nil
}

// WithImageModel sets the model for image generation
func WithImageModel(model string) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.Model = model
	}
}

// WithImageSize sets the size for image generation
func WithImageSize(size string) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.Size = size
	}
}

// WithImageQuality sets the quality for image generation
func WithImageQuality(quality string) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.Quality = quality
	}
}

// WithImageResponseFormat sets the response format for image generation
func WithImageResponseFormat(format string) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.ResponseFormat = format
	}
}

// WithImageCount sets the number of images to generate
func WithImageCount(n int) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.N = n
	}
}

// WithUseMidjourney sets the use of Midjourney for image generation
func WithUseMidjourney(useMidjourney bool) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.UseMidjourney = useMidjourney
	}
}

// WithMidjourneyAction sets the action for Midjourney image generation
func WithMidjourneyAction(action string) func(*ImageGenerationRequest) {
	if action != MidjourneyActionSubmit && action != MidjourneyActionStatus && action != MidjourneyActionAction {
		panic("invalid midjourney action")
	}
	return func(opts *ImageGenerationRequest) {
		opts.MidjourneyAction = action
	}
}

// WithMidjourneyActionContent sets the action content for Midjourney image generation
func WithMidjourneyActionContent(action_content string) func(*ImageGenerationRequest) {
	return func(opts *ImageGenerationRequest) {
		opts.MidjourneyActionContent = action_content
	}
}
