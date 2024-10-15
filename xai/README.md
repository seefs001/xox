# XAI Package

The XAI package provides a robust client for interacting with OpenAI's API, including support for text generation, embeddings, and image generation. It also includes integration with Midjourney for advanced image generation capabilities and an Agent system for complex interactions.

## Table of Contents
- [XAI Package](#xai-package)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
    - [Creating a Client](#creating-a-client)
    - [Text Generation](#text-generation)
    - [Streaming Text Generation](#streaming-text-generation)
    - [Quick Text Generation](#quick-text-generation)
    - [Creating Embeddings](#creating-embeddings)
    - [Image Generation](#image-generation)
    - [Midjourney Integration](#midjourney-integration)
  - [Advanced Usage](#advanced-usage)
    - [Custom HTTP Client](#custom-http-client)
    - [Environment Variables](#environment-variables)
    - [Agents](#agents)
    - [Collaborative Agents](#collaborative-agents)
    - [Streaming Agent Interactions](#streaming-agent-interactions)
  - [Error Handling](#error-handling)
  - [OpenAI Client Details](#openai-client-details)
    - [Structures](#structures)
    - [Key Functions](#key-functions)
      - [Chat Completions](#chat-completions)
      - [Completions](#completions)
      - [Image Generation](#image-generation-1)
      - [Embeddings](#embeddings)
      - [Audio](#audio)
    - [Configuration Options](#configuration-options)
  - [Contributing](#contributing)
  - [License](#license)

## Installation

```go
import "github.com/seefspkg/xai"
```

## Basic Usage

### Creating a Client

```go
client := xai.NewOpenAIClient(
    xai.WithAPIKey("your-api-key"),
    xai.WithBaseURL("https://api.openai.com"),
    xai.WithModel("gpt-3.5-turbo"),
    xai.WithHTTPClient(customHTTPClient),
    xai.WithDebug(true),
)
```

### Text Generation

```go
text, err := client.GenerateText(ctx, xai.TextGenerationOptions{
    Model:       "gpt-3.5-turbo",
    Messages:    []xai.Message{{Role: xai.MessageRoleUser, Content: "Hello, AI!"}},
    Temperature: 0.7,
    MaxTokens:   150,
})
```

### Streaming Text Generation

```go
textChan, errChan := client.GenerateTextStream(ctx, xai.TextGenerationOptions{
    Model:       "gpt-3.5-turbo",
    Messages:    []xai.Message{{Role: xai.MessageRoleUser, Content: "Tell me a story"}},
    Temperature: 0.7,
    MaxTokens:   500,
    ChunkSize:   100,
})

for {
    select {
    case text := <-textChan:
        fmt.Print(text)
    case err := <-errChan:
        if err != nil {
            log.Fatal(err)
        }
        return
    }
}
```

### Quick Text Generation

```go
text, err := client.QuickGenerateText(ctx, []string{"What's the weather like?"},
    xai.WithTextModel("gpt-4"),
    xai.WithChunkSize(50),
)
```

### Creating Embeddings

```go
embeddings, err := client.CreateEmbeddings(ctx, []string{"Hello, world!"}, "text-embedding-ada-002")
```

### Image Generation

```go
urls, err := client.GenerateImage(ctx, "A serene landscape with mountains",
    xai.WithImageModel("dall-e-3"),
    xai.WithImageSize("1024x1024"),
    xai.WithImageQuality("standard"),
    xai.WithImageResponseFormat("url"),
    xai.WithImageCount(1),
)
```

### Midjourney Integration

```go
// Generate image with Midjourney
response, err := client.GenerateImageWithMidjourney(ctx, "A futuristic cityscape",
    xai.WithUseMidjourney(true),
    xai.WithMidjourneyAction(xai.MidjourneyActionSubmit),
)

// Get Midjourney job status
status, err := client.GetMidjourneyStatus(ctx, jobID)

// Perform action on Midjourney image
response, err := client.ActMidjourney(ctx, actionContent, jobID)
```

## Advanced Usage

### Custom HTTP Client

```go
customClient := xhttpc.NewClient(
    xhttpc.WithTimeout(30 * time.Second),
    // Add other custom settings
)

client := xai.NewOpenAIClient(
    xai.WithHTTPClient(customClient),
)
```

### Environment Variables

The package supports loading API key and base URL from environment variables:

- `OPENAI_API_KEY`: Your OpenAI API key
- `OPENAI_API_BASE`: Base URL for the OpenAI API

### Agents

```go
agent := xai.NewAgent(client, toolExecutor,
    xai.WithAgentModel("gpt-4"),
    xai.WithAgentSystemPrompt(xai.ExpertSystemPrompt),
    xai.WithAgentMaxIterations(5),
    xai.WithAgentTemperature(0.7),
    xai.WithAgentTools(tools),
    xai.WithAgentName("ExpertAgent"),
    xai.WithAgentDebug(true),
    xai.WithAgentCallback(callbackFunc),
    xai.WithAgentMemory(xai.NewSimpleMemory(10)),
)

result, err := agent.Run(ctx, "Analyze the current economic trends")
```

### Collaborative Agents

```go
result, err := agent.CollaborateWithAgents(ctx, "Develop a marketing strategy", map[string]*xai.Agent{
    "MarketResearch": marketResearchAgent,
    "ContentCreation": contentCreationAgent,
    "Analytics": analyticsAgent,
})
```

### Streaming Agent Interactions

```go
eventChan, result, err := agent.RunWithEvents(ctx, "Explain quantum computing")

for event := range eventChan {
    // Process events (thoughts, actions, observations, etc.)
    fmt.Printf("Event: %s, Data: %v\n", event.Type, event.Data)
}
```

## Error Handling

```go
text, err := client.GenerateText(ctx, options)
if err != nil {
    log.Fatal(err)
}
```

## OpenAI Client Details

### Structures

- `OpenAIClient`: The main client struct for interacting with OpenAI API.
- `ChatCompletionMessage`: Represents a message in the chat completion.
- `Tool`: Represents a tool that can be used in chat completions.
- `Function`: Describes a function that can be called by the model.

### Key Functions

#### Chat Completions
```go
func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, req CreateChatCompletionRequest) (*CreateChatCompletionResponse, error)
func (c *OpenAIClient) CreateChatCompletionStream(ctx context.Context, req CreateChatCompletionRequest) (<-chan ChatCompletionChunk, error)
```

#### Completions
```go
func (c *OpenAIClient) CreateCompletion(ctx context.Context, req CreateCompletionRequest) (*CreateCompletionResponse, error)
```

#### Image Generation
```go
func (c *OpenAIClient) CreateImage(ctx context.Context, req CreateImageRequest) (*ImagesResponse, error)
func (c *OpenAIClient) CreateImageEdit(ctx context.Context, req CreateImageEditRequest) (*ImagesResponse, error)
func (c *OpenAIClient) CreateImageVariation(ctx context.Context, req CreateImageVariationRequest) (*ImagesResponse, error)
```

#### Embeddings
```go
func (c *OpenAIClient) CreateEmbedding(ctx context.Context, req CreateEmbeddingRequest) (*CreateEmbeddingResponse, error)
```

#### Audio
```go
func (c *OpenAIClient) CreateSpeech(ctx context.Context, req CreateSpeechRequest) ([]byte, error)
func (c *OpenAIClient) CreateTranscription(ctx context.Context, req CreateTranscriptionRequest) (interface{}, error)
```

### Configuration Options

- `WithBaseURL`: Sets the base URL for the OpenAI API.
- `WithAPIKey`: Sets the API key for authentication.
- `WithHTTPClient`: Sets a custom HTTP client.
- `WithModel`: Sets the default model for the OpenAI client.
- `WithDebug`: Enables or disables debug mode.

Example:
```go
client := NewOpenAIClient(
    WithBaseURL("https://api.openai.com"),
    WithAPIKey("your-api-key"),
    WithModel("gpt-3.5-turbo"),
    WithDebug(true),
)
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests to our repository.

## License

This project is licensed under the MIT License - see the LICENSE file for details.