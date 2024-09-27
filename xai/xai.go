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
	"github.com/seefs001/xox/xhttpc"
)

// OpenAIClient represents a client for interacting with the OpenAI API
type OpenAIClient struct {
	baseURL    string
	apiKey     string
	httpClient *xhttpc.Client
	model      string
	debug      bool
}

// OpenAIClientOption is a function type for configuring the OpenAIClient
type OpenAIClientOption func(*OpenAIClient)

// TextGenerationOptions contains options for generating text
type TextGenerationOptions struct {
	Model        string    `json:"model"`
	Prompt       string    `json:"prompt"`
	SystemPrompt string    `json:"system"`
	Messages     []Message `json:"messages"`
	IsStreaming  bool      `json:"stream"`
	ObjectSchema string    `json:"object_schema"`
	Temperature  float64   `json:"temperature,omitempty"`
	MaxTokens    int       `json:"max_tokens,omitempty"`
	TopP         float64   `json:"top_p,omitempty"`
	N            int       `json:"n,omitempty"`
	ChunkSize    int       `json:"chunk_size,omitempty"`
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

// Usage represents the token usage information
type Usage struct {
	PromptTokenCount     int `json:"prompt_tokens"`
	CompletionTokenCount int `json:"completion_tokens"`
	TotalTokenCount      int `json:"total_tokens"`
}

// Choice represents a single choice in the API response
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

// StreamResponse represents a single chunk of the streaming response
type StreamResponse struct {
	ID         string         `json:"id"`
	ObjectType string         `json:"object"`
	CreatedAt  int64          `json:"created"`
	ModelName  string         `json:"model"`
	Choices    []StreamChoice `json:"choices"`
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
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
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
	DefaultBaseURL        = "https://api.openai.com/v1"
	ChatCompletionsURL    = "/chat/completions"
	EmbeddingsURL         = "/embeddings"
	ImageGenerationURL    = "/images/generations"
	DefaultEmbeddingModel = "text-embedding-ada-002"
	DefaultImageModel     = "dall-e-3"
	DefaultChunkSize      = 100 // Default chunk size for streaming
)

// WithBaseURL sets the base URL for the OpenAI API
func WithBaseURL(url string) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.baseURL = url
	}
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

// WithDebug enables or disables debug mode
func WithDebug(debug bool) OpenAIClientOption {
	return func(c *OpenAIClient) {
		c.debug = debug
	}
}

// NewOpenAIClient creates a new OpenAIClient with the given options
func NewOpenAIClient(options ...OpenAIClientOption) *OpenAIClient {
	client := &OpenAIClient{
		baseURL: DefaultBaseURL,
		httpClient: xhttpc.NewClient(
			xhttpc.WithTimeout(30 * time.Second),
		),
		model: DefaultModel,
		debug: false,
	}

	client.loadEnvironmentVariables()

	for _, option := range options {
		option(client)
	}

	if client.debug {
		client.httpClient.SetDebug(true)
	}

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
		return fmt.Errorf("either 'Messages' or 'Prompt'/'SystemPrompt' should be provided, not both")
	}

	if !hasMessages && !hasPromptOrSystem {
		return fmt.Errorf("either 'Messages' or 'Prompt'/'SystemPrompt' must be provided")
	}

	return nil
}

// GenerateText generates text based on the provided options
func (c *OpenAIClient) GenerateText(ctx context.Context, options TextGenerationOptions) (string, error) {
	if err := validateTextGenerationOptions(&options); err != nil {
		return "", err
	}

	if x.IsEmpty(options.Model) {
		options.Model = c.model
	}

	requestBody := map[string]interface{}{
		"model":    options.Model,
		"messages": options.Messages,
	}

	// Only add non-default parameters
	if options.Temperature != 0 {
		requestBody["temperature"] = options.Temperature
	}
	if options.MaxTokens != 0 {
		requestBody["max_tokens"] = options.MaxTokens
	}
	if options.TopP != 0 {
		requestBody["top_p"] = options.TopP
	}
	if options.N != 0 {
		requestBody["n"] = options.N
	}
	if options.IsStreaming {
		requestBody["stream"] = options.IsStreaming
	}

	resp, err := c.sendRequest(ctx, ChatCompletionsURL, requestBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
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
			errChan <- err
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

		// Only add non-default parameters
		if options.Temperature != 0 {
			requestBody["temperature"] = options.Temperature
		}
		if options.MaxTokens != 0 {
			requestBody["max_tokens"] = options.MaxTokens
		}
		if options.TopP != 0 {
			requestBody["top_p"] = options.TopP
		}
		if options.N != 0 {
			requestBody["n"] = options.N
		}

		resp, err := c.sendRequest(ctx, ChatCompletionsURL, requestBody)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		c.handleStreamResponse(ctx, resp, textChan, errChan, options.ChunkSize)
	}()

	return textChan, errChan
}

func (c *OpenAIClient) sendRequest(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	resp, err := c.httpClient.
		SetBaseURL(c.baseURL).
		SetBearerToken(c.apiKey).
		PostJSON(ctx, endpoint, body)

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func (c *OpenAIClient) handleStreamResponse(ctx context.Context, resp *http.Response, textChan chan<- string, errChan chan<- error, chunkSize int) {
	reader := bufio.NewReader(resp.Body)
	buffer := strings.Builder{}

	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

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
				errChan <- fmt.Errorf("error reading stream: %w", err)
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || line == "data: [DONE]" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				errChan <- fmt.Errorf("unexpected line format: %s", line)
				return
			}

			data := strings.TrimPrefix(line, "data: ")
			var streamResponse StreamResponse
			if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
				errChan <- fmt.Errorf("error unmarshaling stream data: %w", err)
				return
			}

			if len(streamResponse.Choices) > 0 && streamResponse.Choices[0].Delta.Content != "" {
				buffer.WriteString(streamResponse.Choices[0].Delta.Content)
				if buffer.Len() > 0 {
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
		Messages:  []Message{},
		ChunkSize: DefaultChunkSize,
	}

	for _, option := range options {
		option(&opts)
	}

	if opts.SystemPrompt != "" {
		opts.Messages = append(opts.Messages, Message{Role: MessageRoleSystem, Content: opts.SystemPrompt})
	}

	for i, content := range prompt {
		role := MessageRoleUser
		if i%2 != 0 {
			role = MessageRoleAssistant
		}

		message := Message{Role: role}
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

	requestBody := EmbeddingRequest{
		Model: model,
		Input: input,
	}

	resp, err := c.sendRequest(ctx, EmbeddingsURL, requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	embeddings := make([][]float64, len(result.Data))
	for i, data := range result.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// GenerateImage generates an image based on the provided prompt
func (c *OpenAIClient) GenerateImage(ctx context.Context, prompt string, options ...func(*ImageGenerationRequest)) ([]string, error) {
	requestBody := ImageGenerationRequest{
		Model:  DefaultImageModel,
		Prompt: prompt,
	}

	for _, option := range options {
		option(&requestBody)
	}

	resp, err := c.sendRequest(ctx, ImageGenerationURL, requestBody)
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
