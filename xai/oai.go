package xai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

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

// OpenAIOption represents an option for configuring the OpenAIClient
type OpenAIOption func(*OpenAIClient)

// Define API endpoints as constants
const (
	DefaultBaseURL              = "https://api.openai.com"
	ChatCompletionsEndpoint     = "/v1/chat/completions"
	CompletionsEndpoint         = "/v1/completions"
	ImagesGenerationsEndpoint   = "/v1/images/generations"
	ImagesEditsEndpoint         = "/v1/images/edits"
	ImagesVariationsEndpoint    = "/v1/images/variations"
	EmbeddingsEndpoint          = "/v1/embeddings"
	AudioSpeechEndpoint         = "/v1/audio/speech"
	AudioTranscriptionsEndpoint = "/v1/audio/transcriptions"
)

// ChatCompletionStreamResponse represents a streaming response chunk
type ChatCompletionStreamResponse struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index        int    `json:"index"`
		Delta        Delta  `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Delta represents the content delta in a streaming response
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ChatCompletionChunk represents a streaming response chunk or an error
type ChatCompletionChunk struct {
	Response *ChatCompletionStreamResponse
	Error    error
}

// Common structures moved from xai.go to oai.go

// ChatCompletionMessage represents a message in the chat completion
type ChatCompletionMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Image      string     `json:"image,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool call in the chat completion response
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call in the tool call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool in the chat completion request
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function in the tool
type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
	Index        int                   `json:"index"`
	Logprobs     interface{}           `json:"logprobs"`
}

// Update CreateChatCompletionRequest to include tools and tool_choice
type CreateChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	Temperature float32                 `json:"temperature,omitempty"`
	TopP        float32                 `json:"top_p,omitempty"`
	N           int                     `json:"n,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
	Stop        []string                `json:"stop,omitempty"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Tools       []Tool                  `json:"tools,omitempty"`
	ToolChoice  string                  `json:"tool_choice,omitempty"`
}

// Update CreateChatCompletionResponse to include usage details
type CreateChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens            int `json:"prompt_tokens"`
		CompletionTokens        int `json:"completion_tokens"`
		TotalTokens             int `json:"total_tokens"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
}

// CreateChatCompletion creates a chat completion
func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, req CreateChatCompletionRequest) (*CreateChatCompletionResponse, error) {
	endpoint := ChatCompletionsEndpoint

	if req.Stream {
		return nil, fmt.Errorf("use CreateChatCompletionStream for streaming responses")
	}

	var resp CreateChatCompletionResponse
	err := c.sendRequestWithResp(ctx, http.MethodPost, endpoint, req, &resp)
	return &resp, err
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *OpenAIClient) CreateChatCompletionStream(ctx context.Context, req CreateChatCompletionRequest) (<-chan ChatCompletionChunk, error) {
	if !req.Stream {
		return nil, fmt.Errorf("Stream must be set to true for streaming responses")
	}

	endpoint := ChatCompletionsEndpoint
	url := c.baseURL + endpoint

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	chunkChan := make(chan ChatCompletionChunk)

	go func() {
		defer resp.Body.Close()
		defer close(chunkChan)

		reader := bufio.NewReader(resp.Body)
		buffer := strings.Builder{}
		const maxOutputInterval = 500 * time.Millisecond
		const minChunkSize = 10 // Minimum number of characters to output
		lastOutputTime := time.Now()

		flushBuffer := func() {
			if buffer.Len() > 0 {
				chunkChan <- ChatCompletionChunk{Response: &ChatCompletionStreamResponse{
					Choices: []struct {
						Index        int    `json:"index"`
						Delta        Delta  `json:"delta"`
						FinishReason string `json:"finish_reason"`
					}{
						{
							Delta: Delta{
								Content: buffer.String(),
							},
						},
					},
				}}
				buffer.Reset()
				lastOutputTime = time.Now()
			}
		}

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					chunkChan <- ChatCompletionChunk{Error: fmt.Errorf("error reading stream: %w", err)}
				}
				flushBuffer()
				return
			}

			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			line = bytes.TrimPrefix(line, []byte("data: "))
			if string(line) == "[DONE]" {
				flushBuffer()
				return
			}

			var streamResp ChatCompletionStreamResponse
			if err := json.Unmarshal(line, &streamResp); err != nil {
				chunkChan <- ChatCompletionChunk{Error: fmt.Errorf("error unmarshaling stream response: %w", err)}
				flushBuffer()
				return
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				buffer.WriteString(streamResp.Choices[0].Delta.Content)
				if buffer.Len() >= minChunkSize || time.Since(lastOutputTime) >= maxOutputInterval {
					flushBuffer()
				}
			}
		}
	}()

	return chunkChan, nil
}

// CreateCompletion creates a completion
func (c *OpenAIClient) CreateCompletion(ctx context.Context, req CreateCompletionRequest) (*CreateCompletionResponse, error) {
	endpoint := CompletionsEndpoint
	var resp CreateCompletionResponse
	err := c.sendRequestWithResp(ctx, http.MethodPost, endpoint, req, &resp)
	return &resp, err
}

// CreateImage creates an image
func (c *OpenAIClient) CreateImage(ctx context.Context, req CreateImageRequest) (*ImagesResponse, error) {
	endpoint := ImagesGenerationsEndpoint
	var resp ImagesResponse
	err := c.sendRequestWithResp(ctx, http.MethodPost, endpoint, req, &resp)
	return &resp, err
}

// CreateImageEdit creates an edited or extended image
func (c *OpenAIClient) CreateImageEdit(ctx context.Context, req CreateImageEditRequest) (*ImagesResponse, error) {
	endpoint := ImagesEditsEndpoint
	var resp ImagesResponse
	err := c.sendMultipartRequest(ctx, endpoint, req, &resp)
	return &resp, err
}

// CreateImageVariation creates a variation of a given image
func (c *OpenAIClient) CreateImageVariation(ctx context.Context, req CreateImageVariationRequest) (*ImagesResponse, error) {
	endpoint := ImagesVariationsEndpoint
	var resp ImagesResponse
	err := c.sendMultipartRequest(ctx, endpoint, req, &resp)
	return &resp, err
}

// CreateEmbedding creates an embedding vector representing the input text
func (c *OpenAIClient) CreateEmbedding(ctx context.Context, req CreateEmbeddingRequest) (*CreateEmbeddingResponse, error) {
	endpoint := EmbeddingsEndpoint
	var resp CreateEmbeddingResponse
	err := c.sendRequestWithResp(ctx, http.MethodPost, endpoint, req, &resp)
	return &resp, err
}

// CreateSpeech generates audio from the input text
func (c *OpenAIClient) CreateSpeech(ctx context.Context, req CreateSpeechRequest) ([]byte, error) {
	endpoint := AudioSpeechEndpoint
	resp, err := c.sendRawRequest(ctx, http.MethodPost, endpoint, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// CreateTranscription transcribes audio into the input language
func (c *OpenAIClient) CreateTranscription(ctx context.Context, req CreateTranscriptionRequest) (interface{}, error) {
	endpoint := AudioTranscriptionsEndpoint
	var resp interface{}
	err := c.sendMultipartRequest(ctx, endpoint, req, &resp)
	return resp, err
}

func (c *OpenAIClient) sendRequestWithResp(ctx context.Context, method, endpoint string, reqBody, respBody interface{}) error {
	url := c.baseURL + endpoint

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	return nil
}

func (c *OpenAIClient) sendMultipartRequest(ctx context.Context, endpoint string, reqBody, respBody interface{}) error {
	url := c.baseURL + endpoint

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	switch req := reqBody.(type) {
	case CreateImageEditRequest:
		addFormField(w, "prompt", req.Prompt)
		addFormField(w, "n", fmt.Sprintf("%d", req.N))
		addFormField(w, "size", req.Size)
		addFormField(w, "response_format", req.ResponseFormat)
		addFormFile(w, "image", req.Image)
		if req.Mask != "" {
			addFormFile(w, "mask", req.Mask)
		}
	case CreateImageVariationRequest:
		addFormField(w, "n", fmt.Sprintf("%d", req.N))
		addFormField(w, "size", req.Size)
		addFormField(w, "response_format", req.ResponseFormat)
		addFormFile(w, "image", req.Image)
	case CreateTranscriptionRequest:
		addFormField(w, "model", req.Model)
		addFormField(w, "language", req.Language)
		addFormField(w, "prompt", req.Prompt)
		addFormField(w, "response_format", req.ResponseFormat)
		addFormField(w, "temperature", fmt.Sprintf("%f", req.Temperature))
		addFormFile(w, "file", req.File)
	default:
		return fmt.Errorf("unsupported request type for multipart request")
	}

	err := w.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	return nil
}

func (c *OpenAIClient) sendRawRequest(ctx context.Context, method, endpoint string, reqBody interface{}) (*http.Response, error) {
	url := c.baseURL + endpoint

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.httpClient.Do(req)
}

func addFormField(w *multipart.Writer, fieldName, value string) {
	w.WriteField(fieldName, value)
}

func addFormFile(w *multipart.Writer, fieldName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	part, err := w.CreateFormFile(fieldName, filePath)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// Request and response structs for OpenAI API endpoints

type CreateCompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float32  `json:"temperature,omitempty"`
	TopP        float32  `json:"top_p,omitempty"`
	N           int      `json:"n,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	LogProbs    int      `json:"logprobs,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type CreateCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		LogProbs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type CreateImageRequest struct {
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type CreateImageEditRequest struct {
	Image          string `json:"image"`
	Mask           string `json:"mask,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type CreateImageVariationRequest struct {
	Image          string `json:"image"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type ImagesResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url,omitempty"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
}

type CreateEmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`
	User           string      `json:"user,omitempty"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

type CreateEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type CreateSpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float32 `json:"speed,omitempty"`
}

type CreateTranscriptionRequest struct {
	File           string  `json:"file"`
	Model          string  `json:"model"`
	Language       string  `json:"language,omitempty"`
	Prompt         string  `json:"prompt,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Temperature    float32 `json:"temperature,omitempty"`
}
