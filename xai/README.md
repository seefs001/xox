# XAI Package

The XAI package provides a robust client for interacting with OpenAI's API, including support for text generation, embeddings, and image generation. It also includes integration with Midjourney for advanced image generation capabilities.

## Installation

```go
import "github.com/seefspkg/xai"
```

## Usage

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

## Error Handling

All methods return errors when appropriate. Always check and handle these errors in your code.

```go
text, err := client.GenerateText(ctx, options)
if err != nil {
    log.Fatal(err)
}
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests to our repository.

## License
