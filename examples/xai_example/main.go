package main

import (
	"context"

	"github.com/seefs001/xox/xai"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xlog"
)

func main() {
	xlog.Info("Starting AI example")

	xenv.Load()
	xlog.Info("Environment variables loaded")

	client := xai.NewOpenAIClient(xai.WithDebug(true))
	xlog.Info("OpenAI client created with debug mode enabled")

	// Text generation (non-streaming)
	xlog.Info("Generating text using QuickGenerateText")
	response, err := client.QuickGenerateText(context.Background(), []string{"Hello, world!"}, xai.WithTextModel(xai.ModelGPT4o))
	if err != nil {
		xlog.Error("Error generating text", "error", err)
		return
	}
	xlog.Info("Text generation response:", "response", response)

	// Text generation (streaming)
	xlog.Info("Starting streaming text generation")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	textChan, errChan := client.QuickGenerateTextStream(ctx, []string{"Hello, world!"}, xai.WithTextModel(xai.ModelClaude35Sonnet))

	streamFinished := make(chan struct{})
	go func() {
		defer close(streamFinished)
		for {
			select {
			case text, ok := <-textChan:
				if !ok {
					xlog.Info("Stream finished")
					return
				}
				xlog.Info("Received text from stream", "text", text)
			case err, ok := <-errChan:
				if !ok {
					return
				}
				if err != nil {
					xlog.Error("Error generating text stream", "error", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	<-streamFinished

	// Image generation
	xlog.Info("Generating image")
	imageURLs, err := client.GenerateImage(context.Background(), "A beautiful sunset over the ocean", xai.WithImageModel(xai.DefaultImageModel))
	if err != nil {
		xlog.Error("Error generating image", "error", err)
		return
	}
	for i, url := range imageURLs {
		xlog.Info("Generated image URL", "index", i+1, "url", url)
	}

	// Embedding generation
	xlog.Info("Creating embeddings")
	embeddings, err := client.CreateEmbeddings(context.Background(), []string{"Hello, world!"}, xai.DefaultEmbeddingModel)
	if err != nil {
		xlog.Error("Error creating embeddings", "error", err)
		return
	}
	xlog.Info("Embedding created", "embedding_sample", embeddings[0][:5]) // Print first 5 values of the embedding

	xlog.Info("AI example completed")
}
